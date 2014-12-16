package main

import (
	"fmt"
	"os"

	"github.com/jcelliott/lumber"

	// "github.com/nanobox-core/logtap"
	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/worker"
	// "github.com/nanobox-core/router"
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
	log := lumber.NewConsoleLogger(lumber.INFO)

	// boot sequence...
	// all of these steps should happen exactly once

	// parse the provided config file or use defautls
	config, err := ParseConfig()
	if err != nil {
		fmt.Println("Unable to configure. Aborting... ", err)
		os.Exit(1)
	}

	// create new scribble
	scribble, err := scribble.New(config.scribbleDir, log)
	if err != nil {
		fmt.Println("Unable to initialize database driver, aborting...")
		os.Exit(1)
	}

	// create new logtap
	// logtap, err := logtap.New(config.logtapPort, log)
	// if err != nil {
	// 	fmt.Println("Unable to initialize database driver, aborting...")
	// 	os.Exit(1)
	// }

	// create new router
	// router, err := router.New(config.routerPort, log)
	// if err != nil {
	// 	fmt.Println("Unable to initialize database driver, aborting...")
	// 	os.Exit(1)
	// }

	// create new mist
	mist, err := mist.New(config.mistPort, log)
	if err != nil {
		fmt.Println("Unable to create new Mist, aborting...")
		os.Exit(1)
	}

	// create a worker
	worker, err := worker.New(mist, log)
	if err != nil {
		fmt.Println("Unable to initialize API, aborting...")
		os.Exit(1)
	}

	// everything inside of nanobox should only be created once
	nanobox := &api.API{
		Driver: scribble,
		// Logtap : logtap,
		// Router : router,
		Mist:   mist,
		Worker: worker,
		Log:    log,
	}

	//
	// start nanobox
	if err := nanobox.Start("1757"); err != nil {
		fmt.Printf("Unable to start nanobox: %v", err)
		os.Exit(1)
	}
}
