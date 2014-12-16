package util

import (
	"os"
	"os/exec"

	"github.com/nanobox-core/mist"
)

type (
	Container struct{}
)

func (c *Container) Install(m *mist.Mist) error {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/download.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/install.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/start.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/scripts/container/cleanup.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	return nil

}
