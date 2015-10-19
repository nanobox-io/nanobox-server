// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package script

//
import (
	"encoding/json"
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

	return docker.ExecInContainer(container, "/opt/bin/"+name, string(b))
}
