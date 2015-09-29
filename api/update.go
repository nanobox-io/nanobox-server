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

// UpdateImages
func (api *API) UpdateImages(rw http.ResponseWriter, req *http.Request) {

	//
	api.Worker.QueueAndProcess(&jobs.ImageUpdate{})

	//
	rw.Write([]byte("{\"id\":\"1\", \"status\":\"created\"}"))
}
