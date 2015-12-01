// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"strings"

	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
	"github.com/nanobox-io/nanobox-server/util/worker"
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

	_, err := docker.InspectContainer("build1")
	if err != nil {
		util.UpdateStatus(j, "unavailable")
		return
	}

	// parse the boxfile
	box := combinedBox()

	// define the build payload
	j.payload = map[string]interface{}{
		"platform":    "local",
		"app":         config.App(),
		"dns":         []string{config.App() + ".dev"},
		"port":        "8080",
		"boxfile":     box.Node("build").Parsed,
		"logtap_host": config.LogtapHost,
	}

	// grab the environment data from all service containers

	worker := worker.New()
	worker.Blocking = true
	worker.Concurrent = true

	//
	serviceEnvs := []*ServiceEnv{}

	serviceContainers, _ := docker.ListContainers("service")
	for _, container := range serviceContainers {

		s := ServiceEnv{UID: container.Config.Labels["uid"]}
		serviceEnvs = append(serviceEnvs, &s)

		worker.Queue(&s)
	}

	worker.Process()

	evars := DefaultEVars(box)

	failedEnv := false
	for _, env := range serviceEnvs {
		if !env.Success {
			util.HandleError(stylish.ErrorHead("Failed to configure %v's environment variables", env.UID))
			util.HandleError(stylish.ErrorBody(""))
			failedEnv = true
			continue
		}

		for key, val := range env.EVars {
			evars[strings.ToUpper(env.UID+"_"+key)] = val
		}
	}
	if failedEnv {
		util.UpdateStatus(j, "errored")

		return
	}

	j.payload["env"] = evars

	if err := j.RunBuild(); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	restarts := []*Restart{}

	// find code containers and run the restart hook
	containers, _ := docker.ListContainers("code")
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

func (j *Build) RunBuild() error {
	// run sync hook (blocking)
	if _, err := script.Exec("default-sync", "build1", j.payload); err != nil {
		return err
	}

	// run build hook (blocking)
	if _, err := script.Exec("default-build", "build1", j.payload); err != nil {
		return err
	}

	// run publish hook (blocking)
	if _, err := script.Exec("default-publish", "build1", j.payload); err != nil {
		return err
	}

	// run cleanup script (blocking)
	if _, err := script.Exec("default-cleanup", "build1", j.payload); err != nil {
		return err
	}

	return nil
}
