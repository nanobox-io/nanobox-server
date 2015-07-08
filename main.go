// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package main

import (
	"os"

	"github.com/pagodabox/nanobox-server/api"
	"github.com/pagodabox/nanobox-server/config"
)

//
const Version = "0.0.1"

//
func main() {

	// read the config file and create configurations
	config.Init()

	// initialize the api and set up routing
	api := api.Init()

	// start nanobox
	if err := api.Start(config.APIPort); err != nil {
		config.Log.Fatal("[NANOBOX] Unable to start API, aborting...\n%v\n", err)
		os.Exit(1)
	}
}
