package main

import (
	"fmt"
	"os"
)

//
func main() {
	configMap, err := parseConfig(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(configMap)
}
