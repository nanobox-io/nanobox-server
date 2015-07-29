// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

type ServiceEnv struct {
	deploy Deploy

	EVars   map[string]string
	UID     string
	Success bool
}

func (j *ServiceEnv) Process() {
	j.Success = false

	// run environment hook (blocking)
	if out, err := util.ExecHook("environment", j.UID, nil); err != nil {
		util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", j.UID), err.Error()), "")
		// util.UpdateStatus(j.deploy, "errored")
		return
	} else {
		if err := json.Unmarshal(out, &j.EVars); err != nil {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", j.UID), err.Error()), "")
			// util.UpdateStatus(j.deploy, "errored")
			return
		}
	}

	// if a service doesnt have a port we cant continue
	if j.EVars["port"] == "" {
		util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's tunnel", j.UID), "no port given in environment"), "")
		return
	}

	// now we need to set the host in the evars as well as create a tunnel port in the router
	container, err := util.InspectContainer(j.UID)
	if err != nil {
		util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's tunnel", j.UID), err.Error()), "")
	}

	j.EVars["host"] = container.NetworkSettings.IPAddress
	if config.Router.GetForward(j.UID) == nil {
		config.Router.AddForward(j.UID, j.EVars["host"]+":"+j.EVars["port"])
	}

	j.Success = true
}
