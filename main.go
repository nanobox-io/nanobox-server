package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/pat"

	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/db"
	"github.com/nanobox-core/nanobox-server/workers"
)

type (

	//
	Nanobox struct {
		api     api.API         //
		db      db.Driver       //
		mist    mist.Mist       //
		workers workers.Factory //

		opts map[string]string //
	}
)

//
func main() {

	// everything inside of nanobox should only be created once
	nanobox := &Nanobox{
		api:     api.API{},
		db:      db.Driver{},
		workers: workers.Factory{},
	}

	// boot sequence...
	// all of these steps should happen exactly once

	// parse the provided config file or use defautls
	if err := nanobox.config(); err != nil {
		fmt.Println("Unable to configure. Aborting... ", err)
		os.Exit(1)
	}

	// initialize the database driver
	if status := nanobox.db.Init(nanobox.opts); status != 0 {
		fmt.Println("Unable to initialize database driver, aborting...")
		os.Exit(status)
	}

	// initialize workers
	if status := nanobox.workers.Init(nanobox.opts); status != 0 {
		fmt.Println("Unable to initialize workers, aborting...")
		os.Exit(status)
	}

	// create new mist
	mist, err := mist.New(nanobox.opts)
	if err != nil {
		fmt.Println("Unable to create new Mist, aborting...")
		os.Exit(1)
	}

	nanobox.mist = mist

	// initialize the api, providing it with a database driver, and mist adapter
	if status := nanobox.api.Init(nanobox.opts, &nanobox.db, &nanobox.mist); status != 0 {
		fmt.Println("Unable to initialize API, aborting...")
		os.Exit(status)
	}

	//
	// start nanobox
	if err := nanobox.Start(); err != nil {
		fmt.Printf("Unable to start nanobox: %v", err)
		os.Exit(1)
	}
}

// Start
func (n *Nanobox) Start() error {

	//
	p := pat.New()

	//
	fmt.Println("Registering routes...")
	api.InitRoutes(p, &n.api)

	fmt.Println("Starting server...")
	fmt.Printf("Nanobox listening at %v\n", n.api.Server.Addr)

	//
	http.Handle("/", p)
	err := http.ListenAndServe(n.api.Server.Addr, nil)
	if err != nil {
		return err
	}

	return nil
}
