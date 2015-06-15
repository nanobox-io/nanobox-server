// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/pagodabox/nanobox-server/config"
)

type Startup struct {
}

// process on startup
func (s *Startup) Process() {
	// when we start we need to reset up the routing
	if container, err := tasks.GetContainer("web1"); err == nil {
		dc, _ := tasks.GetDetailedContainer(container.Id)

		config.Router.AddTarget("/", dc.NetworkSettings.IPAddress)
		config.Router.Handler = nil
	} else {
		config.Router.Handler = router.NoDeploy{}
	}
}
