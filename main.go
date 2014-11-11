package main

import (
	"fmt"
	"os"
)

var router Router

//
func main() {

	// command line args w/o program
	args := os.Args[1:]

	// set default config file
	config := "default.conf"

	// override default if config file provided
	if len(args) >= 1 {
		config = args[0]
	}

	// parse config file
	opts, err := parseConfig(config)
	if err != nil {
		fmt.Printf("Unable to parse config file: %s", err)
		os.Exit(1)
	}

	//
	router.host = opts["host"]
	router.port = opts["port"]
	router.addr = router.host + ":" + router.port

	if err := router.start(); err != nil {
		fmt.Printf("Unable to start router: %s", err)
		os.Exit(1)
	}
}
