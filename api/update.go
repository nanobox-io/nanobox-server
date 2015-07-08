// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"net/http"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/data"
)

// CreateEVar
func (api *API) UpdateImages(rw http.ResponseWriter, req *http.Request) {
	config.Log.Debug("[NANOBOX :: API] Deploy create\n")

	imageUpdate := data.ImageUpdate{}
	api.Worker.QueueAndProcess(&imageUpdate)

	rw.Write([]byte("{\"id\":\"1\"}"))
}
