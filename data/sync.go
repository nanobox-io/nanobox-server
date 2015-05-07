package data

import (
	// "time"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/hookyd/go-client"
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
	range container, _ := tasks.ListContainers() {

		tasks.RemoveContainer(container(id))
	}

	// create a build container
	con, err := tasks.CreateContainer("engine-image") // i dont have an image name yet
	if err != nil {
		ch <- "could not create build image container"
		config.Log.Error("could not create build image container(%s)", err.Error())
		return
	}

	addr := con.NetworkSettings.IPAddress

	h := hooky.Hooky{
		Host: addr,
		Port: 1234, // dont know the port
	}

	box := boxfile.NewFromPath("some path")
	
	// run build hooks

	// combine boxfiles


	// build containers according to boxfile

	// run hooks in new containers

	// before deploy hooks
	// set routing to web components
	// after deploy hooks

}
