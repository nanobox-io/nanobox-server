// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/data"
)

// CreateEVar
func (api *API) CreateDeploy(rw http.ResponseWriter, req *http.Request) {
	config.Log.Debug("[NANOBOX :: API] Deploy create\n")

	sync := data.Sync{
		Id:    uuid.New(),
		Reset: req.FormValue("reset") == "true",
	}
	api.Worker.QueueAndProcess(&sync)

	rw.Write([]byte("{\"id\":\"" + sync.Id + "\", \"status\":\"created\"}"))
}
