package adm

import (
	"os"
	"os/exec"
)

type (
	Network struct{}
)

func (c *Network) Install(mist chan<- string) error {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/config.sh").Output(); err != nil {
		return err
	} else {
		mist<- string(out)
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/broadcast.sh").Output(); err != nil {
		return err
	} else {
		mist<- string(out)
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/cleanup.sh").Output(); err != nil {
		return err
	} else {
		mist<- string(out)
	}

	return nil

}
