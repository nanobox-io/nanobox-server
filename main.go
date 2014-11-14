package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/pat"

	"github.com/nanobox-core/nanobox-server/api"
	"github.com/nanobox-core/nanobox-server/db"
)

type (

	//
	Nanobox struct {
		api api.API
		db  db.Driver

		opts map[string]string
	}
)

//
func main() {

	// everything inside of nanobox should only be created once
	nanobox := &Nanobox{
		api: api.API{},
		db:  db.Driver{},
	}

	// boot sequence...

	// parse the provided config file or use defautls
	if err := nanobox.config(); err != nil {
		fmt.Println("Unable to configure. Aborting... ", err)
		os.Exit(1)
	}

	// initialize the database driver
	if status := nanobox.db.Init(nanobox.opts); status != 0 {
		fmt.Println("Unable to initialize database driver. Aborting...")
		os.Exit(status)
	}

	// initialize the api
	if status := nanobox.api.Init(nanobox.opts, &nanobox.db); status != 0 {
		fmt.Println("Unable to initialize API. Aborting...")
		os.Exit(status)
	}

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
	fmt.Printf("\nListening at %v\n", n.api.Server.Addr)

	//
	http.Handle("/", p)
	err := http.ListenAndServe(n.api.Server.Addr, nil)
	if err != nil {
		return err
	}

	return nil
}
