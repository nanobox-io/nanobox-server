// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"encoding/json"
	"strconv"

	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
)

type ServiceEnv struct {
	deploy  Deploy
	EVars   map[string]string
	UID     string
	Success bool
}

func (j *ServiceEnv) Process() {
	j.Success = false

	// run environment hook (blocking)
	if out, err := script.Exec("environment", j.UID, nil); err != nil {
		util.HandleError(stylish.ErrorHead("Failed to configure %v's environment variables", j.UID))
		util.HandleError(stylish.ErrorBody(err.Error()))
		util.UpdateStatus(&j.deploy, "errored")
		return
	} else {
		config.Log.Info("getting port data: %s", out)
		if err := json.Unmarshal(out, &j.EVars); err != nil {
			util.HandleError(stylish.ErrorHead("Failed to configure %v's environment variables", j.UID))
			util.HandleError(stylish.ErrorBody(err.Error()))
			util.UpdateStatus(&j.deploy, "errored")
			return
		}
	}
	config.Log.Debug("getting port data: %+v", j.EVars)
	// if a service doesnt have a port we cant continue
	if j.EVars["PORT"] == "" {
		util.HandleError(stylish.ErrorHead("Failed to configure %v's tunnel", j.UID))
		util.HandleError(stylish.ErrorBody("no port given in environment"))
		return
	}

	// now we need to set the host in the evars as well as create a tunnel port in the router
	container, err := docker.InspectContainer(j.UID)
	if err != nil {
		util.HandleError(stylish.ErrorHead("Failed to configure %v's tunnel", j.UID))
		util.HandleError(stylish.ErrorBody(err.Error()))
	}
	config.Log.Debug("container: %+v", container)

	j.EVars["HOST"] = container.NetworkSettings.IPAddress
	err = util.AddForward(j.EVars["PORT"], j.EVars["HOST"], j.EVars["PORT"])
	if err != nil {
		port, _ := strconv.Atoi(j.EVars["PORT"])
		for i := 1; i <= 10; i++ {
			err = util.AddForward(strconv.Itoa(port+i), j.EVars["HOST"], j.EVars["PORT"])
			if err == nil {
				break
			}
		}
		if err != nil {
			util.HandleError(stylish.Error("Failed to setup forward for service", err.Error()))
		}
	}

	j.Success = true
}
