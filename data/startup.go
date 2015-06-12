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
