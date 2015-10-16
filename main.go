// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package main

import (
	"os"

	logapi "github.com/nanobox-io/nanobox-logtap/api"
	"github.com/nanobox-io/nanobox-logtap/archive"
	"github.com/nanobox-io/nanobox-logtap/collector"
	"github.com/nanobox-io/nanobox-logtap/drain"
	"github.com/nanobox-io/nanobox-router"
	"github.com/nanobox-io/nanobox-server/api"
	"github.com/nanobox-io/nanobox-server/config"
)

//
func main() {

	// create a new mist and start listening for messages at *:1445
	config.Mist.Listen(config.Ports["mist"])

	setupLogtap()

	// create new router
	err := router.StartHTTP(":" + config.Ports["router"])
	if err != nil {
		config.Log.Error("error: %s\n", err.Error())
	}

	// initialize the api and set up routing
	api := api.Init()

	// start nanobox
	if err := api.Start(config.Ports["api"]); err != nil {
		config.Log.Fatal("[nanobox/main.go] Unable to start API, aborting...\n%v\n", err)
		os.Exit(1)
	}
}

func setupLogtap() {
	//
	console := drain.AdaptLogger(config.Log)
	config.Logtap.AddDrain("console", console)

	// define logtap collectors/drains; we don't need to defer Close() anything here,
	// because these want to live as long as the server
	if _, err := collector.SyslogUDPStart("app", config.Ports["logtap"], config.Logtap); err != nil {
		panic(err)
	}

	//
	if _, err := collector.SyslogTCPStart("app", config.Ports["logtap"], config.Logtap); err != nil {
		panic(err)
	}

	// we will be adding a 0 to the end of the logtap port because we cant have 2 tcp listeneres
	// on the same port
	if _, err := collector.StartHttpCollector("deploy", config.Ports["logtap"]+"0", config.Logtap); err != nil {
		panic(err)
	}

	//
	db, err := archive.NewBoltArchive("/tmp/bolt.db")
	if err != nil {
		panic(err)
	}
	config.LogHandler = logapi.GenerateArchiveEndpoint(db)

	//
	config.Logtap.AddDrain("historical", db.Write)
	config.Logtap.AddDrain("mist", drain.AdaptPublisher(config.Mist))

}
