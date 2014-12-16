package main

import (
	"fmt"
	"os"

	"github.com/jcelliott/lumber"

	// "github.com/nanobox-core/logtap"
	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/worker"
	"github.com/nanobox-core/router"
	"github.com/nanobox-core/scribble"
)

const (
	DefaultAPIHost     = "0.0.0.0"
	DefaultAPIPort     = "1757"
	DefaultLogTapPort  = "514"
	DefaultMistPort    = "1445"
	DefaultRouterPort  = "80"
	DefaultScribbleDir = "./tmp/db"
	DefaultLogtapDir   = "./tmp/logs"
)

//
func main() {

	//
	logger := lumber.NewConsoleLogger(lumber.INFO)

	// boot sequence...
	// all of these steps should happen exactly once

	// parse the provided config file or use defautls
	config, err := ParseConfig()
	if err != nil {
		fmt.Println("Failed to configure, Aborting... ", err)
		os.Exit(1)
	}

	// create new scribble
	scribble, err := scribble.New(config.scribbleDir, logger)
	if err != nil {
		fmt.Println("Unable to create new Scribble, aborting...")
		os.Exit(1)
	}

	// create new logtap
	// logtap, err := logtap.New(config.logtapPort, logger)
	// if err != nil {
	// 	fmt.Println("Unable to initialize database driver, aborting...")
	// 	os.Exit(1)
	// }

	// create new router
	router := router.New(config.routerPort, logger)

	// create new mist
	mist := mist.New(config.mistPort, logger)

	// create a worker
	worker := worker.New(mist, logger)

	// everything inside of nanobox should only be created once
	nanobox := &api.API{
		Driver: scribble,
		// Logtap : logtap,
		Router: router,
		Mist:   mist,
		Worker: worker,
		Log:    logger,
	}

	//
	// start nanobox
	if err := nanobox.Start(config.port); err != nil {
		fmt.Printf("Unable to start nanobox: %v", err)
		os.Exit(1)
	}
}
