// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	// "time"
	"fmt"
	"strings"
	"regexp"
	"time"
	"encoding/json"

	"github.com/hookyd/go-client"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/pagodabox/nanobox-server/worker"
)

//
type Sync struct {
	Id string
}

func (s *Sync) deployLog(message string) {
	config.Logtap.Publish("deploy", message)
}

func (s *Sync) handleError(message string, err error) {
	s.deployLog(message)
	errMessage := ""
	if err == nil {
		errMessage = "noerror"
	} else {
		errMessage = err.Error()
	}

	config.Log.Error("%s (%s)\n", message, errMessage)
	config.Router.Handler = router.FailedDeploy{}
	time.Sleep(10 * time.Second)
	s.updateStatus("complete")
}

func (s *Sync) updateStatus(status string) {
	config.Mist.Publish([]string{"sync"}, fmt.Sprintf(`"action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, s.Id, status))
}

// this method syncronies your docker containers
// with the boxfile specification
func (s *Sync) Process() {
	// clear the deploy log
	config.Logtap.Drains["history"].(*logtap.HistoricalDrain).ClearDeploy()

	config.Log.Debug("[NANOBOX :: SYNC] Started\n")
	s.deployLog("Starting A new deploy")

	// set routing to watch logs
	s.deployLog("[NANOBOX :: SYNC] setting routes")
	config.Log.Debug("[NANOBOX :: SYNC] setting routes\n")
	config.Router.Handler = router.DeployInProgress{}

	// remove all code containers
	s.deployLog("[NANOBOX :: SYNC] clearing old containers")
	config.Log.Debug("[NANOBOX :: SYNC] clearing old containers")

	containers, _ := tasks.ListContainers("code", "build")
	config.Log.Debug("[NANOBOX :: SYNC] containers (%#v)", containers)
	for _, container := range containers {
		config.Log.Debug("[NANOBOX :: SYNC] clean container %#v\n", container)
		err := tasks.RemoveContainer(container.Id)
		if err != nil {
			s.handleError("[NANOBOX :: SYNC] There is a problem removing old docker containers", err)
			return
		}
	}

	// make sure we have the directories and wipe the deploy from the previous
	// deploy
	if err := tasks.Clean(); err != nil {
		s.handleError("[NANOBOX :: SYNC] Could not clean code directories", err)
		return
	}

	// create a build container
	config.Log.Debug("[NANOBOX :: SYNC] creating container")
	if !tasks.ImageExists("nanobox/build") {
		s.deployLog("Pulling the latest build image... This could take a while.")
	}
	
	con, err := tasks.CreateContainer("nanobox/build", map[string]string{"build": "true", "uid": "build"})
	if err != nil {
		s.handleError("[NANOBOX :: SYNC] could not create build image container", err)
		return
	}

	s.deployLog("[NANOBOX :: SYNC] container created")
	config.Log.Debug("[NANOBOX :: SYNC] container created %#v", con)

	addr := con.NetworkSettings.IPAddress

	h := hooky.Hooky{
		Host: addr,
		Port: "5540",
	}

	box := boxfile.NewFromPath("/vagrant/code/"+config.App+"/Boxfile")

	config.Log.Info("boxfile after parsing!! %#v\n", box)
	payload := map[string]interface{}{
		"app": config.App,
		"dns": []string{config.App+".gonano.io"},
		"env": map[string]string{"APP_NAME":config.App},
		"port":"8080",
		"boxfile": box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	s.deployLog("[NANOBOX :: SYNC] running sniff hook")
	config.Log.Info("[NANOBOX :: SYNC] running sniff hook")

	config.Log.Info("[NANOBOX :: SYNC] pload: %#v", payload)
	cPayload, err := json.Marshal(payload)
	config.Log.Info("[NANOBOX :: SYNC] json: %s, %#v", cPayload, err)

	s.deployLog("[NANOBOX :: SYNC] running configure hook")
	time.Sleep(10 * time.Second)
	if response, err := h.Run("build-configure", cPayload, s.Id); err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build-configure) problem(exit: %d, out: %s)", response.Exit, response.Out), err)
		return
	}

	s.deployLog("[NANOBOX :: SYNC] running prepare hook")
	if response, err := h.Run("build-prepare", cPayload, s.Id); err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build-prepare) problem(exit: %d, out: %s)", response.Exit, response.Out), err)
		return
	}

	s.deployLog("[NANOBOX :: SYNC] running boxfile hook")
	response, err := h.Run("build-boxfile", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build-boxfile) problem(exit: %d, out: %s)", response.Exit, response.Out), err)
		return
	} else {
		// combine boxfiles
		s.deployLog(fmt.Sprintf("boxfile hook response: %#v\n", response))
		config.Log.Debug("boxfile %s\n", response.Out)
		hookBox := boxfile.New([]byte(response.Out))
		config.Log.Debug("hookbox: %#v", hookBox)
		box.Merge(hookBox)
	}

	serviceWorker := worker.New()
	serviceWorker.Blocking = true
	serviceWorker.Concurrent = true

	serviceStarts := []*ServiceStart{}

	// build containers according to boxfile
	// run hooks in new containers (services)
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		config.Log.Debug("[NANOBOX :: SYNC] looking at node (%s)", node)
		if node != "nanobox" &&
			node != "global" &&
			node != "build" &&
			name != "web" &&
			name != "worker" {
				config.Log.Debug("[NANOBOX :: SYNC] looks good making sure we dont have one already")
			if _, err := tasks.GetContainer("nanobox/"+node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node,
				}
				config.Log.Debug("[NANOBOX :: SYNC] sure didnt (%#v)", s)
				serviceStarts = append(serviceStarts, &s)

				serviceWorker.Queue(&s)
			}
		}
	}

	serviceWorker.Process()

	evars := payload["env"].(map[string]string)

	for _, serv := range serviceStarts {
		if !serv.Success {
			s.handleError("[NANOBOX :: SYNC] A Service was not started correctly ("+serv.Uid+")", err)			
			return
		}

		for key, val := range serv.EnvVars {
			evars[strings.ToUpper(serv.Uid+"_"+key)] = val
		}
	}

	payload["env"] = evars

	s.deployLog("[NANOBOX :: SYNC] running build hook")
	pload, _ := json.Marshal(payload)
	response, err = h.Run("build-build", pload, "3")
	if err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build-build) problem(exit: %d, out: %s)", response.Exit, response.Out), err)
		return
	}

	s.deployLog("[NANOBOX :: SYNC] running cleanup hook")
	if response, err := h.Run("build-cleanup", pload, "4"); err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build-cleanup) problem(exit: %d, out: %s)", response.Exit, response.Out), err)
		return
	}

	// remove build
	s.deployLog("[NANOBOX :: SYNC] remove build container")

	config.Log.Debug("[NANOBOX :: SYNC] remove build container")
	tasks.RemoveContainer(con.Id)

	// run hooks in new containers (code)
	codeWorker := worker.New()
	codeWorker.Blocking = true
	codeWorker.Concurrent = true

	codeServices := []*ServiceStart{}
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		if name == "web" || name == "worker" {
			if _, err := tasks.GetContainer(node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node,
				}
				codeServices = append(codeServices, &s)
				codeWorker.Queue(&s)
			}
		}
	}

	codeWorker.Process()

	for _, serv := range codeServices {
		if !serv.Success {
			s.handleError("A Service was not started correctly ("+serv.Uid+")", err)
			return
		}
	}

	// before deploy hooks
	for _, node := range box.Nodes() {
		n := node
		bd  := box.Node(n).Value("before_deploy")
		bda := box.Node(n).Value("before_deploy_all")
		if bd != nil || bda != nil {
			if container, err := tasks.GetContainer(n); err == nil {
				dc, _ := tasks.GetDetailedContainer(container.Id)
				addr := dc.NetworkSettings.IPAddress

				h := hooky.Hooky{
					Host: addr,
					Port: "1234", // dont know the port
				}

				pload, _ := json.Marshal(map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda})

				if response, err := h.Run("code-before_deploy", pload, "0"); err != nil || response.Exit != 0 {
					s.deployLog(fmt.Sprintf("[NANOBOX :: SYNC] hook(code-before_deploy) problem(exit: %d, out: %s) err(%#v)", response.Exit, response.Out, err))
				}

			}
		}
	}

	// set routing to web components
	s.deployLog("[NANOBOX :: SYNC] set routing")
	config.Log.Debug("[NANOBOX :: SYNC] set routing")
	if container, err := tasks.GetContainer("web1"); err == nil {
		dc, _ := tasks.GetDetailedContainer(container.Id)

		config.Router.AddTarget("/", "http://"+dc.NetworkSettings.IPAddress+":8080")
		config.Router.Handler = nil
	}
	
	// after deploy hooks
	for _, node := range box.Nodes() {
		n := node
		ad  := box.Node(n).Value("after_deploy")
		ada := box.Node(n).Value("after_deploy_all")
		if ad != nil || ada != nil {
			if container, err := tasks.GetContainer(n); err == nil {
				dc, _ := tasks.GetDetailedContainer(container.Id)
				addr := dc.NetworkSettings.IPAddress

				h := hooky.Hooky{
					Host: addr,
					Port: "1234", // dont know the port
				}

				pload, _ := json.Marshal(map[string]interface{}{"after_deploy": ad, "after_deploy_all": ada})

				if response, err := h.Run("code-after_deploy", pload, "0"); err != nil || response.Exit != 0 {
					s.deployLog(fmt.Sprintf("[NANOBOX :: SYNC] hook(code-after_deploy) problem(exit: %d, out: %s) err(%#v)", response.Exit, response.Out, err))
				}

			}
		}
	}

	s.deployLog("[NANOBOX :: SYNC] sync complete")
	config.Log.Debug("[NANOBOX :: SYNC] sync complete")
	s.updateStatus("complete")

}
