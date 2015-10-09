// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"fmt"
	"net/http"

	"github.com/nanobox-io/nanobox-server/util"
)

func (api *API) Suspend(rw http.ResponseWriter, req *http.Request) {
	if util.LockCount() <= 0 {
		return
	}

	writeBody(map[string]string{"error": fmt.Sprintf("Current lock count: %d", util.LockCount())}, rw, http.StatusNotAcceptable)
}

func (api *API) LockCount(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%d", util.LockCount())
}

// keeps a lock open as long as the connection is established with
// my service
func (api *API) Lock(rw http.ResponseWriter, req *http.Request) {
	util.Lock()
	defer util.Unlock()

	cNotify := rw.(http.CloseNotifier)
	<-cNotify.CloseNotify()
}
