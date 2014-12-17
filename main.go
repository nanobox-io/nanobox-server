package main

import (
	"fmt"
	"os"

	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/config"
)

//
const Version = "0.0.1"

//
func main() {

	//
	config := config.Init()

	//
	api := api.Init()

	//
	// start nanobox
	if err := api.Start(config.Port); err != nil {
		fmt.Printf("Unable to start API: %v", err)
		os.Exit(1)
	}
}
