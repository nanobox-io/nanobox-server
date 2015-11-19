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

	"github.com/nanobox-io/nanobox-router"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
)

//
type Service struct {
	CreatedAt time.Time
	IP        string
	UID       string
	Name      string `json:",omitempty"`
	Ports     []int
	Username  string            `json:",omitempty"`
	Password  string            `json:",omitempty"`
	EnvVars   map[string]string `json:",omitempty"`
}

func (api *API) ListRoutes(rw http.ResponseWriter, req *http.Request) {
	writeBody(router.Routes(), rw, http.StatusOK)
}

func (api *API) ListVips(rw http.ResponseWriter, req *http.Request) {
	vips, _ := util.ListVips()
	writeBody(vips, rw, http.StatusOK)	
}

// ListServices
func (api *API) ListServices(rw http.ResponseWriter, req *http.Request) {

	// a list of services to be returned in the response
	services := []Service{}

	// interate over each container building a corresponding service for that container
	// and then add it to the list of services that will be passed back as the
	// response
	containers, _ := docker.ListContainers("service")
	for _, container := range containers {

		// a 'service' representing the container
		// uid := strings.Replace(container.Name, "/", "", 1)
		service := Service{
			CreatedAt: container.Created,
			IP:        container.NetworkSettings.IPAddress,
			UID:       container.Config.Labels["uid"],
			Name:      container.Config.Labels["name"],
		}

		ports := []int{}
		vips, _ := util.ListVips()
		for _, vip := range vips {
			for _, server := range vip.Servers {
				if server.Host == service.IP {
					ports = append(ports, vip.Port)
				}
			}
		}
		service.Ports = ports

		// run environment hook (blocking)
		if out, err := script.Exec("environment", container.ID, nil); err == nil {
			config.Log.Info("getting port data: %s", out)
			uidlessEvar := map[string]string{}
			if err := json.Unmarshal(out, &uidlessEvar); err == nil {
				service.Password = uidlessEvar["PASS"]
				service.Username = uidlessEvar["USER"]
				upUid := strings.ToUpper(service.UID)
				evars := map[string]string{}
				for key, value := range uidlessEvar {
					evars[upUid+"_"+key] = value
				}
				service.Password = evars["PASS"]
				service.Username = evars["USER"]
				service.EnvVars = evars
			}
		}
		config.Log.Info("service: %+v", service)

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
