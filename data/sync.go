// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"fmt"
	"strings"
	"regexp"
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
	Reset bool
}

func (s *Sync) updateStatus(status string) {
	config.Mist.Publish([]string{"sync"}, fmt.Sprintf(`{"model":"Sync", "action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, s.Id, status))
}

// this method syncronies your docker containers
// with the boxfile specification
func (s *Sync) Process() {
	// clear the deploy log
	config.Logtap.Drains["history"].(*logtap.HistoricalDrain).ClearDeploy()

	logInfo("Starting A new deploy")

	// set routing to watch logs
	logDebug("[NANOBOX :: SYNC] setting routes")
	config.Router.Handler = router.DeployInProgress{}

	// remove all code containers
	logInfo("[NANOBOX :: SYNC] clearing old containers")

	containers, _ := tasks.ListContainers("code", "build")
	logDebug("[NANOBOX :: SYNC] containers (%#v)\n", containers)
	for _, container := range containers {
		logDebug("[NANOBOX :: SYNC] clean container %#v\n", container)
		err := tasks.RemoveContainer(container.Id)
		if err != nil {
			handleError("[NANOBOX :: SYNC] There is a problem removing old docker containers", err)
			s.updateStatus("complete")
			return
		}
	}

	// Make sure we have the directories
	if err := tasks.CreateDirs(); err != nil {
		handleError("[NANOBOX :: SYNC] Failed to create directories", err)
		s.updateStatus("complete")
		return
	}

	// make sure we have the directories and wipe the deploy from the previous
	// deploy
	if s.Reset {
		logInfo("[NANOBOX :: SYNC] Resetting cache and code directories")
		if err := tasks.Clean(); err != nil {
			handleError("[NANOBOX :: SYNC] Could not reset code directories", err)
			s.updateStatus("complete")
			return
		}
		logDebug("[NANOBOX :: SYNC] Cache and Code directories are clean")
	}

	// create a build container
	logDebug("[NANOBOX :: SYNC] creating container\n")
	if !tasks.ImageExists("nanobox/build") {
		logInfo("[NANOBOX :: SYNC] Pulling the latest build image... This could take a while.")
	}
	
	con, err := tasks.CreateContainer("nanobox/build", map[string]string{"build": "true", "uid": "build"})
	if err != nil {
		handleError("[NANOBOX :: SYNC] could not create build image container", err)
		s.updateStatus("complete")
		return
	}

	logInfo("[NANOBOX :: SYNC] container created")
	logDebug("[NANOBOX :: SYNC] New build cotainer: %#v\n", con)

	addr := con.NetworkSettings.IPAddress

	box := boxfile.NewFromPath("/vagrant/code/"+config.App+"/Boxfile")

	logDebug("[NANOBOX :: SYNC] Parsed Boxfile %+v\n", box.Parsed)
	payload := map[string]interface{}{
		"app": config.App,
		"dns": []string{config.App+".gonano.io"},
		"env": map[string]string{"APP_NAME":config.App},
		"port":"8080",
		"boxfile": box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	logInfo("[NANOBOX :: SYNC] running sniff hook")

	logDebug("[NANOBOX :: SYNC] Build Hook Payload: %+v\n", payload)
	cPayload, err := json.Marshal(payload)
	logDebug("[NANOBOX :: SYNC] Build Hook Payload (json): %s\n", cPayload)

	h, err := hooky.New(addr, "5540")
	if err != nil {
		handleError("[NANOBOX :: SYNC] hooky connection for (build)", err)
		s.updateStatus("complete")
		return
	}

	logInfo("[NANOBOX :: SYNC] running configure hook")
	response, err := h.Run("configure", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(configure) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (configure): %+v\n", response)

	logInfo("[NANOBOX :: SYNC] running prepare hook")
	response, err = h.Run("prepare", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(prepare) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (prepare): %+v\n", response)

	logInfo("[NANOBOX :: SYNC] running boxfile hook")
	response, err = h.Run("boxfile", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(boxfile) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	} else {
		// combine boxfiles
		logDebug("[NANOBOX :: SYNC] Hook Response (boxfile): %+v\n", response)
		hookBox := boxfile.New([]byte(response.Out))
		logDebug("[NANOBOX :: SYNC] Boxfile from hook: %+v\n", hookBox)
		box.Merge(hookBox)
		logDebug("[NANOBOX :: SYNC] Merged Boxfile: %+v\n", box)
	}
	

	serviceWorker := worker.New()
	serviceWorker.Blocking = true
	serviceWorker.Concurrent = true

	serviceStarts := []*ServiceStart{}

	// build containers according to boxfile
	// run hooks in new containers (services)
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		logDebug("[NANOBOX :: SYNC] looking at node (%s)\n", node)
		if node != "nanobox" &&
			node != "global" &&
			node != "build" &&
			name != "web" &&
			name != "worker" {
			if _, err := tasks.GetContainer("nanobox/"+node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node,
				}
				logDebug("[NANOBOX :: SYNC] sure didnt (%+v)", s)
				serviceStarts = append(serviceStarts, &s)

				serviceWorker.Queue(&s)
			}
		}
	}

	serviceWorker.Process()

	evars := payload["env"].(map[string]string)

	for _, serv := range serviceStarts {
		if !serv.Success {
			handleError("[NANOBOX :: SYNC] A Service was not started correctly ("+serv.Uid+")", err)			
			s.updateStatus("complete")
			return
		}

		for key, val := range serv.EnvVars {
			evars[strings.ToUpper(serv.Uid+"_"+key)] = val
		}
	}

	payload["env"] = evars

	logInfo("[NANOBOX :: SYNC] running build hook")
	pload, _ := json.Marshal(payload)
	response, err = h.Run("build", pload, "3")
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (build): %+v\n", response)

	logInfo("[NANOBOX :: SYNC] running cleanup hook")
	if response, err := h.Run("cleanup", pload, "4"); err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(cleanup) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (cleanup): %+v\n", response)

	// remove build
	logInfo("[NANOBOX :: SYNC] remove build container")

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
			handleError("A Service was not started correctly ("+serv.Uid+")", err)
			s.updateStatus("complete")
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

				h, err := hooky.New(addr, "5540")
				if err != nil {
					handleError("[NANOBOX :: SYNC] hooky connection for ("+n+")", err)
					s.updateStatus("complete")
					return
				}
				pload, _ := json.Marshal(map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda})

				if response, err := h.Run("before_deploy", pload, "0"); err != nil || response.Exit != 0 {
					logInfo(fmt.Sprintf("[NANOBOX :: SYNC] hook(before_deploy) problem(%+v) err(%s)", response, err))
				}

			}
		}
	}

	// set routing to web components
	logInfo("[NANOBOX :: SYNC] set routing")
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

				h, err := hooky.New(addr, "5540")
				if err != nil {
					handleError("[NANOBOX :: SYNC] hooky connection for ("+n+")", err)
					s.updateStatus("complete")
					return
				}
				pload, _ := json.Marshal(map[string]interface{}{"after_deploy": ad, "after_deploy_all": ada})

				if response, err := h.Run("after_deploy", pload, "0"); err != nil || response.Exit != 0 {
					logInfo(fmt.Sprintf("[NANOBOX :: SYNC] hook(after_deploy) problem(%+v) err(%s)", response, err))
				}

			}
		}
	}

	logInfo("[NANOBOX :: SYNC] sync complete")
	s.updateStatus("complete")
}
