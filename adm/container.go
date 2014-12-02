package util

import (
	"os"
	"os/exec"

	"github.com/nanobox-core/mist"
)

type (
	Container struct{}
)

func (c *Container) Install(m *mist.Mist) {

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//
	if out, err := exec.Command(cwd + "/priv/container/download.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/priv/container/install.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/priv/container/start.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

	//
	if out, err := exec.Command(cwd + "/priv/container/cleanup.sh").Output(); err != nil {
		panic(err)
	} else {
		m.Publish([]string{"evars"}, string(out))
	}

}
