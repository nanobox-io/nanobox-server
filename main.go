package main

import (
	"fmt"
	"os"
)

//

func main() {

	// command line args w/o program
	args := os.Args[1:]

	// set default config file
	configFile := "default.conf"

	// override default if config file provided
	if len(args) >= 1 {
		configFile = args[0]
	}

	// `parse config file
	configMap, err := parseConfig(configFile)
	if err != nil {
		fmt.Printf("Unable to parse config file: %s", err)
		os.Exit(1)
	}

	// success!
	fmt.Println(configMap)
}
