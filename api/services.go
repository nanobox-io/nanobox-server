// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

// CreateEVar
func (api *API) ListServices(rw http.ResponseWriter, req *http.Request) {
	config.Log.Debug("[NANOBOX :: API] List Services\n")

	containers, _ := util.ListContainers()
	data := []map[string]string{}
	for _, container := range containers {
		dc, _ := util.InspectContainer(container.ID)

		c := container.Config.Labels
		c["ip"] = dc.NetworkSettings.IPAddress
		tunnelPort := config.Router.GetLocalPort(c["ip"] + ":22")
		c["tunnel_port"] = strconv.Itoa(tunnelPort)
		data = append(data, c)
	}

	b, err := json.Marshal(data)
	if err != nil {
		config.Log.Error("[NANOBOX :: API] list services (%s)", err.Error())
	}
	rw.Write(b)
}
