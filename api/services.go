// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Service struct {
	CreatedAt time.Time
	IP        string
	Name      string
	Port      int
}

// ListServices
func (api *API) ListServices(rw http.ResponseWriter, req *http.Request) {

	// a list of services to be returned in the response
	services := []Service{}

	// interate over each container building a corresponding service for that container
	// and then add it to the list of services that will be passed back as the
	// response
	containers, _ := util.ListContainers()
	for _, container := range containers {

		// a 'service' representing the container
		name := strings.Replace(container.Name, "/", "", 1)
		service := Service{
			CreatedAt: container.Created,
			IP:        container.NetworkSettings.IPAddress,
			Name:      name,
			Port:      config.Router.GetLocalPort(name),
		}

		// add the service to the list to be returned
		services = append(services, service)
	}

	// marshall the services to json
	b, err := json.Marshal(services)
	if err != nil {
		config.Log.Error("[NANOBOX :: API] list services (%s)", err.Error())
	}

	// return the list of services
	rw.Write(b)
}
