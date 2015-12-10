// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"net/http"
	"sync"

	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/jobs"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/fs"
	"github.com/nanobox-io/nanobox-server/util/script"
)

var developTex = sync.Mutex{}

func (api *API) Develop(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		config.Log.Debug("form parsing error: \n %s", err.Error())
	}

	// force the exec route to go into a dev1 container
	req.Form["container"] = []string{"dev1"}

	box := jobs.CombinedBoxfile(false)

	image := "nanobox/build"
	if stab := box.Node("build").StringValue("stability"); stab != "" {
		image = image + ":" + stab
	}

	control, err := ensureContainer(image, req.FormValue("dev_config"))
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	}

	api.Exec(rw, req)

	if control {
		execWait.Wait()
		docker.RemoveContainer("dev1")
	}
}

func ensureContainer(image, dev_config string) (control bool, err error) {
	developTex.Lock()
	defer developTex.Unlock()
	// if there is no dev1 it needs to be created and this thread needs to remember
	// to shut it down when its done conatinerControl is used for that purpose
	container, err := docker.GetContainer("dev1")
	if err != nil || !container.State.Running {
		config.Log.Debug("develop controlling container")
		config.Log.Debug("develop container: %+v", container)
		if container != nil {
			config.Log.Debug("develop container config: %+v", container.Config)
			config.Log.Debug("develop container host: %+v", container.HostConfig)
		}

		if container != nil && !container.State.Running {
			config.Log.Debug("removing old dev1")
			err = docker.RemoveContainer(container.ID)
			if err != nil {
				config.Log.Debug("develop remove containter: %s", err.Error())
				return false, err
			}
		}
		control = true
		err = nil

		// give docker a new set of lib dirs
		jobs.SetLibDirs()

		container, err = docker.CreateContainer(docker.CreateConfig{Image: image, Category: "dev", UID: "dev1"})
		if err != nil {
			config.Log.Debug("develop create containter: %s", err.Error())
			return false, err
		}

		// run the default-user hook to get ssh keys setup
		out, err := script.Exec("default-user", "dev1", fs.UserPayload())
		if err != nil {
			config.Log.Debug("Failed script output: \n %s", out)
			config.Log.Debug("out: %s", string(out))
		}

		pload := map[string]interface{}{
			"boxfile": jobs.CombinedBoxfile(false).Node("dev").Parsed,
			"dev_config": dev_config,
		}

		out, err = script.Exec("dev-prepare", "dev1", pload)
		if err != nil {
			config.Log.Debug("Failed script output: \n %s", out)
			config.Log.Debug("out: %s", string(out))
		}

	}
	return
}
