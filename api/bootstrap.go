// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"net/http"

	"github.com/pagodabox/nanobox-server/jobs"
)

// CreateDeploy
func (api *API) CreateBootstrap(rw http.ResponseWriter, req *http.Request) {

	bootstrap := jobs.Bootstrap{
		ID:     newUUID(),
		Engine: req.FormValue("engine"),
	}
	api.Worker.QueueAndProcess(&bootstrap)

	rw.Write([]byte("{\"id\":\"" + bootstrap.ID + "\", \"status\":\"created\"}"))
}
