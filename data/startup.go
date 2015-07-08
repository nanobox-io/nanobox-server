// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
)

type Startup struct {
}

// process on startup
func (s *Startup) Process() {

	// start all the old containers
	containers, _ := tasks.ListContainers()
	for _, container := range containers {
		err := tasks.StartContainer(container.Id)
		if err != nil {
			config.Log.Error(err.Error())
		}
	}

	// when we start we need to reset up the routing
	if container, err := tasks.GetContainer("web1"); err == nil {
		dc, _ := tasks.GetDetailedContainer(container.Id)

		config.Router.AddTarget("/", "http://"+dc.NetworkSettings.IPAddress+":8080")
		config.Router.Handler = nil
	} else {
		config.Router.Handler = router.NoDeploy{}
	}

	// we also need to set up a ssh tunnel for each running docker container
	for _, container := range containers {
		dc, err := tasks.GetDetailedContainer(container.Id)
		if err == nil {
			config.Router.AddForward(dc.NetworkSettings.IPAddress + ":22")
		}
	}
}
