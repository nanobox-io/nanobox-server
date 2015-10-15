// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/fs"
	"github.com/nanobox-io/nanobox-server/util/script"
)

//
type Bootstrap struct {
	ID     string
	Engine string
}

// Bootstrap the code according to the engine provided
func (j *Bootstrap) Process() {

	// Make sure we have the directories
	util.LogDebug(stylish.Bullet("Ensure directories exist on host..."))
	if err := fs.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// if the build image doesn't exist it needs to be downloaded
	if !docker.ImageExists("nanobox/build") {
		util.LogInfo(stylish.Bullet("Pulling the latest build image (this will take awhile)... "))
		docker.InstallImage("nanobox/build")
	}

	// create a build container
	util.LogInfo(stylish.Bullet("Creating build container..."))
	_, err := docker.CreateContainer(docker.CreateConfig{Image: "nanobox/build", Category: "bootstrap", UID: "bootstrap1"})
	if err != nil {
		util.HandleError(stylish.Error("Failed to create build container", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// define the deploy payload
	payload := map[string]interface{}{
		"platform":    "local",
		"engine":      j.Engine,
		"logtap_host": config.LogtapHost,
	}

	// run configure hook (blocking)
	if _, err := script.Exec("default-bootstrap", "bootstrap1", payload); err != nil {
		util.HandleError(stylish.Error("Failed to run bootstrap hook", err.Error()))
		util.UpdateStatus(j, "errored")
	}

	docker.RemoveContainer("bootstrap1")

	util.UpdateStatus(j, "complete")
}
