// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package script

//
import (
	"encoding/json"
	"fmt"

	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
)

// Exec executes a script using docker
// It is not in the docker package becuase it isnt a function of docker
// but more a function of our system
// it makes more sense to do script.Exec then docker.ExecScript
// it is alos a var instead of a package function so we can swap it out for a
// mock function in tests.
var Exec = func(name, container string, payload map[string]interface{}) ([]byte, error) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	// marshal the payload
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	out, err := docker.ExecInContainer(container, "/opt/bin/"+name, string(b))
	if err != nil {
		util.LogDebug("Failed script output(%s): \n %s", name, out)
		util.HandleError(stylish.Error(fmt.Sprintf("Failed to run %s script", name), err.Error()))
	}
	return out, err
}
