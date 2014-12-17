package adm

import (
	"os"
	"os/exec"
)

type (
	Container struct{}
)

func (c *Container) Install(ch chan<- string) error {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/download.sh").Output(); err != nil {
		return err
	} else {
		ch <- string(out)
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/install.sh").Output(); err != nil {
		return err
	} else {
		ch <- string(out)
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/start.sh").Output(); err != nil {
		return err
	} else {
		ch <- string(out)
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/cleanup.sh").Output(); err != nil {
		return err
	} else {
		ch <- string(out)
	}

	return nil

}
