// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hookyd/go-client"
	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/pagodabox/nanobox-server/worker"
)

//
type Sync struct {
	Id    string
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

	// wipe the previous deploy data if reset == true
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

	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// build the payload and parse boxfile
	logDebug("[NANOBOX :: SYNC] Parsed Boxfile %+v\n", box.Parsed)
	payload := map[string]interface{}{
		"app":        config.App,
		"dns":        []string{config.App + ".gonano.io"},
		"env":        map[string]string{"APP_NAME": config.App},
		"port":       "8080",
		"boxfile":    box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	logDebug("[NANOBOX :: SYNC] Build Hook Payload: %+v\n", payload)
	cPayload, err := json.Marshal(payload)
	logDebug("[NANOBOX :: SYNC] Build Hook Payload (json): %s\n", cPayload)

	h, err := hooky.New(addr, "5540")
	if err != nil {
		handleError("[NANOBOX :: SYNC] hooky connection for (build)", err)
		s.updateStatus("complete")
		return
	}

	// run configure hook
	logInfo("[NANOBOX :: SYNC] running configure hook")
	response, err := h.Run("configure", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(configure) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (configure): %+v\n", response)

	// run prepare hook
	logInfo("[NANOBOX :: SYNC] running prepare hook")
	response, err = h.Run("prepare", cPayload, s.Id)
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(prepare) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (prepare): %+v\n", response)

	// run boxfile hook
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

	// add the misssing storage nodes to the boxfile
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		logDebug("[NANOBOX :: SYNC] looking at node (%s)\n", node)
		if (name == "web" || name == "worker") && box.Node(node).Value("network_dirs") != nil {
			found := false
			for _, storage := range box.Node(node).Node("network_dirs").Nodes() {
				found = true
				if !box.Node(storage).Valid {
					box.Parsed[storage] = map[string]interface{}{
						"found": true,
					}
				}
			}
			// if i dont find anything but they did have a network_dirs.. just try adding a new one
			if !found {
				if !box.Node("nfs1").Valid {
					box.Parsed["nfs1"] = map[string]interface{}{
						"found": true,
					}
				}
			}
		}
	}

	// remove any containers no longer in the boxfile
	serviceContainers, _ := tasks.ListContainers("service")
	for _, container := range serviceContainers {
		if !box.Node(container.Labels["uid"]).Valid {
			logDebug("[NANOBOX :: SYNC] removing service(%s)", container.Labels["uid"])
			tasks.RemoveContainer(container.Id)
		}
	}

	work := worker.New()
	work.Blocking = true
	work.Concurrent = true

	serviceStarts := []*ServiceStart{}

	// build service containers according to boxfile
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		logDebug("[NANOBOX :: SYNC] looking at node (%s)\n", node)
		if node != "nanobox" &&
			node != "global" &&
			node != "build" &&
			name != "web" &&
			name != "worker" {
			if _, err := tasks.GetContainer(node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node,
				}
				logDebug("[NANOBOX :: SYNC] creating new service (%+v)", s)
				serviceStarts = append(serviceStarts, &s)

				work.Queue(&s)
			}
		}
	}

	work.Process()

	for _, serv := range serviceStarts {
		if !serv.Success {
			handleError("[NANOBOX :: SYNC] A Service was not started correctly ("+serv.Uid+")", err)
			s.updateStatus("complete")
			return
		}
	}

	// grab the environment data from all service containers
	evars := payload["env"].(map[string]string)

	serviceEnvs := []*ServiceEnv{}
	serviceContainers, _ = tasks.ListContainers("service")
	for _, container := range serviceContainers {
		dc, _ := tasks.GetDetailedContainer(container.Id)
		
		s := ServiceEnv{
			Uid:  container.Labels["uid"],
			Addr: dc.NetworkSettings.IPAddress,
		}
		logDebug("[NANOBOX :: SYNC] creating new service (%+v)", s)
		serviceEnvs = append(serviceEnvs, &s)

		work.Queue(&s)
	}

	work.Process()

	for _, env := range serviceEnvs {
		if !env.Success {
			handleError("[NANOBOX :: SYNC] A Service didnt return evars correctly ("+env.Uid+")", err)
			s.updateStatus("complete")
			return
		}

		for key, val := range env.EnvVars {
			evars[strings.ToUpper(env.Uid+"_"+key)] = val
		}
	}

	payload["env"] = evars

	// run build hook
	logInfo("[NANOBOX :: SYNC] running build hook")
	pload, _ := json.Marshal(payload)
	logDebug("[NANOBOX :: SYNC] revised payload: %s", pload)
	response, err = h.Run("build", pload, "3")
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(build) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (build): %+v\n", response)

	// run cleanup hook
	logInfo("[NANOBOX :: SYNC] running cleanup hook")
	if response, err := h.Run("cleanup", pload, "4"); err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC] hook(cleanup) problem(%+v)", response), err)
		s.updateStatus("complete")
		return
	}
	logDebug("[NANOBOX :: SYNC] Hook Response (cleanup): %+v\n", response)

	// remove build container
	logInfo("[NANOBOX :: SYNC] remove build container")

	tasks.RemoveContainer(con.Id)

	// build new code containers

	codeServices := []*ServiceStart{}
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		if name == "web" || name == "worker" {
			if _, err := tasks.GetContainer(node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node,
					EnvVars: evars,
				}
				codeServices = append(codeServices, &s)
				work.Queue(&s)
			}
		}
	}

	work.Process()

	for _, serv := range codeServices {
		if !serv.Success {
			handleError("A Service was not started correctly ("+serv.Uid+")", err)
			s.updateStatus("complete")
			return
		}
	}

	// run before deploy hooks
	for _, node := range box.Nodes() {
		n := node
		bd := box.Node(n).Value("before_deploy")
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
		ad := box.Node(n).Value("after_deploy")
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
