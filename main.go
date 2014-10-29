package main

import (
	"fmt"
	"os"
)

//
func main() {

	var fileName string
	//If no args, parse default.conf
	if len(os.Args) < 2 {
		fileName = "default.conf"
	} else {
		fileName = os.Args[1]
	}
	configMap, err := parseConfig(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(configMap)
}

// func main() {

//   // command line args w/o program
//   args := os.Args[1:]

//   // set default config file
//   configFile := "default.conf"

//   // override default if config file provided
//   if len(args) >= 1 {
//     configFile = args[0]
//   }

//   // parse config file
//   configMap, err := parseConfig(configFile)
//   if err != nil {
//     fmt.Printf("Unable to parse config file: %s", err)
//     os.Exit(1)
//   }

//   // success!
//   fmt.Println(configMap)
// }
