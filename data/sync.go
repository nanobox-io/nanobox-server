package data

import (
	"time"

	"github.com/nanobox-core/nanobox-server/config"
	"github.com/nanobox-core/nanobox-server/tasks"
)

//
type Sync struct {
}

func (s *Sync) Process() {

	ch := make(chan string)
	defer close(ch)

	go func() {
		for data := range ch {
			config.Mist.Publish([]string{"sync"}, data)
		}
	}()

	// set routing to watch logs

	// remove all existing containers

	// create a build container

	// run build hooks

	// combine boxfiles

	// build containers according to boxfile

	// run hooks in new containers

	// before deploy hooks
	// set routing to web components
	// after deploy hooks

}
