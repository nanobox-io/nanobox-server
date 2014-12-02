package util

import (
	"os"
	"os/exec"

	"github.com/nanobox-core/mist"
)

type (
	Network struct{}
)

func (c *Network) Install(m *mist.Mist) {

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//
	if out, err := exec.Command(cwd + "/priv/network/config.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/priv/network/broadcast.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/priv/network/cleanup.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

}
