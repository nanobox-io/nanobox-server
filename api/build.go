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

// CreateBuild
func (api *API) CreateBuild(rw http.ResponseWriter, req *http.Request) {

	//
	build := jobs.Build{
		ID:    newUUID(),
		Reset: (req.FormValue("reset") == "true"),
	}

	//
	api.Worker.QueueAndProcess(&build)

	//
	rw.Write([]byte("{\"id\":\"" + build.ID + "\", \"status\":\"created\"}"))
}
