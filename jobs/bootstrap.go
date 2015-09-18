// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
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
	if err := util.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// if the build image doesn't exist it needs to be downloaded
	if !util.ImageExists("nanobox/build") {
		util.LogInfo(stylish.Bullet("Pulling the latest build image (this will take awhile)... "))
		util.InstallImage("nanobox/build")
	}

	// create a build container
	util.LogInfo(stylish.Bullet("Creating build container..."))
	_, err := util.CreateContainer(util.CreateConfig{Image: "nanobox/build", Category: "bootstrap", Name: "bootstrap1"})
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
	if _, err := util.ExecHook("default-bootstrap", "bootstrap1", payload); err != nil {
		util.HandleError(stylish.Error("Failed to run bootstrap hook", err.Error()))
		util.UpdateStatus(j, "errored")
	}

	util.RemoveContainer("bootstrap1")

	util.UpdateStatus(j, "complete")
}
