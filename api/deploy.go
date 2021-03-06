// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"net/http"

	"github.com/nanobox-io/nanobox-server/jobs"
)

// CreateDeploy
func (api *API) CreateDeploy(rw http.ResponseWriter, req *http.Request) {

	//
	deploy := jobs.Deploy{
		ID:    newUUID(),
		Reset: (req.FormValue("reset") == "true"),
		Run:   (req.FormValue("run") == "true"),
	}

	//
	api.Worker.QueueAndProcess(&deploy)

	//
	rw.Write([]byte("{\"id\":\"" + deploy.ID + "\", \"status\":\"created\"}"))
}
