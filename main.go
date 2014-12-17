package main

import (
	"fmt"
	"os"

	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/config"
	"github.com/nanobox-core/nanobox-server/worker"
)

//
const Version = "0.0.1"

//
func main() {

	//
	config := config.Init()

	//
	nanobox := &api.API{
		Worker: worker.New(),
	}

	//
	// start nanobox
	if err := nanobox.Start(config.Port); err != nil {
		fmt.Printf("Unable to start nanobox: %v", err)
		os.Exit(1)
	}
}
