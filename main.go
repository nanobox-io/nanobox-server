package main

import (
	"os"

	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/config"
)

//
const Version = "0.0.1"

//
func main() {

	//
	config.Init()

	//
	api := api.Init()

	//
	// start nanobox
	if err := api.Start(config.APIPort); err != nil {
		config.Log.Fatal("[NANOBOX] Unable to start API, aborting...\n%v\n", err)
		os.Exit(1)
	}
}
