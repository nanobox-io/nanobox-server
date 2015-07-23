// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

//
import (
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Startup struct{}

// process on startup
func (j *Startup) Process() {

	// start all the old containers
	containers, _ := util.ListContainers()
	for _, container := range containers {
		err := util.StartContainer(container.ID)
		if err != nil {
			config.Log.Error(err.Error())
		}
	}

	// when we start we need to reset up the routing
	if container, err := util.GetContainer("web1"); err == nil {
		dc, _ := util.InspectContainer(container.ID)

		config.Router.AddTarget("/", "http://"+dc.NetworkSettings.IPAddress+":8080")
		config.Router.Handler = nil
	} else {
		config.Router.Handler = router.NoDeploy{}
	}

	// we also need to set up a ssh tunnel for each running docker container
	for _, container := range containers {
		dc, err := util.InspectContainer(container.ID)
		if err == nil {
			config.Router.AddForward(dc.NetworkSettings.IPAddress + ":22")
		}
	}
}
