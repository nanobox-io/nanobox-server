package util

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type (
	Network struct{}
)

func (c *Network) Install() {

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/network/config.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/network/broadcast.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/network/cleanup.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

}
