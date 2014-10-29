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
