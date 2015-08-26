// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"fmt"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Build struct {
	ID    string
	Reset bool

	payload map[string]interface{}
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Build) Process() {

	// parse the boxfile
	util.LogInfo(stylish.Bullet("Parsing Boxfile..."))
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// define the build payload
	j.payload = map[string]interface{}{
		"app":        config.App,
		"dns":        []string{config.App + ".nano.dev"},
		"port":       "8080",
		"boxfile":    box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}
	var env map[string]interface{}{}
	if box.Node("env").Valid {
		env = box.Node("env").Parsed
	}
	env["APP_NAME"] = config.App
	j.payload["env"] = env

	// run sync hook (blocking)
	if _, err := util.ExecHook("sync", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run sync hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run build hook (blocking)
	if _, err := util.ExecHook("build", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run build hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run publish hook (blocking)
	if _, err := util.ExecHook("publish", "build1", j.payload); err != nil {
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
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to restart %v", restart.UID), "unsuccessful restart"))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	util.UpdateStatus(j, "complete")
}
