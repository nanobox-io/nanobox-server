// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

//
import (
	"github.com/pagodabox/nanobox-boxfile"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Startup struct{}

// process on startup
func (j *Startup) Process() {
	config.Log.Info("starting startup job")

	util.RemoveContainer("exec1")
	// TODO get the boxfile. merge with build boxfile(if any) and call:
	// configureRoutes(box)
	// configurePorts(box)
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// if i have a build make sure to merge the boxfile
	_, err := util.InspectContainer("build1")
	if err == nil {
		if out, err := util.ExecHook("default-boxfile", "build1", nil); err == nil {
			box.Merge(boxfile.New([]byte(out)))
		}
	}

	configureRoutes(box)
	configurePorts(box)

	// we also need to set up a ssh tunnel for each running docker container
	// this is easiest to do by creating a ServiceEnv job and working it
	worker := util.NewWorker()
	worker.Blocking = true
	worker.Concurrent = true

	serviceContainers, _ := util.ListContainers("service")
	for _, container := range serviceContainers {
		s := ServiceEnv{UID: container.Config.Labels["uid"]}
		worker.Queue(&s)
	}

	worker.Process()
}
