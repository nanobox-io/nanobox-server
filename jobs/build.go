// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
)

//
type Build struct {
	ID    string
	Reset bool

	payload map[string]interface{}
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Build) Process() {
	// add a lock so the service wont go down whil im running
	util.Lock()
	defer util.Unlock()

	_, err := util.InspectContainer("build1")
	if err != nil {
		util.UpdateStatus(j, "unavailable")
		return
	}

	// parse the boxfile
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// define the build payload
	j.payload = map[string]interface{}{
		"platform":    "local",
		"app":         config.App,
		"dns":         []string{config.App + ".dev"},
		"port":        "8080",
		"boxfile":     box.Node("build").Parsed,
		"logtap_host": config.LogtapHost,
	}
	evar := map[string]interface{}{}
	if box.Node("env").Valid {
		evar = box.Node("env").Parsed
	}
	evar["APP_NAME"] = config.App
	j.payload["env"] = evar

	// run sync hook (blocking)
	if _, err := util.ExecHook("default-sync", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run sync hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run build hook (blocking)
	if _, err := util.ExecHook("default-build", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run build hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run publish hook (blocking)
	if _, err := util.ExecHook("default-publish", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run publish hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	worker := util.NewWorker()
	worker.Blocking = true
	worker.Concurrent = true

	restarts := []*Restart{}

	// find code containers and run the restart hook
	containers, _ := util.ListContainers("code")
	for _, container := range containers {

		uid := container.Config.Labels["uid"]

		r := Restart{UID: uid}
		restarts = append(restarts, &r)
		worker.Queue(&r)
	}

	worker.Process()

	// ensure all services started correctly before continuing
	for _, restart := range restarts {
		if !restart.Success {
			util.HandleError(stylish.ErrorHead("Failed to restart %v", restart.UID))
			util.HandleError(stylish.ErrorBody("unsuccessful restart"))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	util.UpdateStatus(j, "complete")
}
