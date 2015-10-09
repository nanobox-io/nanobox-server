// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package main

import (
	"os"

	"github.com/nanobox-io/nanobox-server/api"
	"github.com/nanobox-io/nanobox-server/config"
)

//
func main() {

	// initialize the api and set up routing
	api := api.Init()

	// start nanobox
	if err := api.Start(config.Ports["api"]); err != nil {
		config.Log.Fatal("[nanobox/main.go] Unable to start API, aborting...\n%v\n", err)
		os.Exit(1)
	}
}
