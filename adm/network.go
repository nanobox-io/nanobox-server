package util

import (
	"os"
	"os/exec"

	"github.com/nanobox-core/mist"
)

type (
	Network struct{}
)

func (c *Network) Install(m *mist.Mist) error {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/config.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/broadcast.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/scripts/network/cleanup.sh").Output(); err != nil {
		return err
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	return nil

}
