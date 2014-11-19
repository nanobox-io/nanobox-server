package util

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type (
	Container struct{}
)

func (c *Container) Install() {

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/container/download.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/container/install.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/container/start.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

	//
	time.Sleep(time.Second * 1)
	if out, err := exec.Command(cwd + "/priv/container/cleanup.sh").Output(); err != nil {
		panic(err)
	} else {
		fmt.Println("OUT: ", string(out))
	}

}

// for i := 0; i < len(slice)/2; i++ {
//  slice[i]
//  slice[len(slice)-i] = slice[len(slice)-i]
//  slice[i]
// }
