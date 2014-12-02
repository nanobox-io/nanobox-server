package workers

import (
	"fmt"

	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/adm"
)

type (

	//
	EVar struct {
		Action string
		ID     string
		// Adapter
	}
)

// Start
func (w EVar) Start(done chan<- bool, m *mist.Mist) {

	//
	if debugging {
		fmt.Printf("EVar worker '%v' is working job '%v'...\n\n", w.ID, w.Action)
	}

	//
	switch w.Action {

	//
	case "list":
		w.list(m)
	}

	// release...
	done <- true
}

//
func (w EVar) list(m *mist.Mist) {

	container := util.Container{}
	network := util.Network{}

	//
	m.Publish([]string{"evars"}, "Doing container things...")
	container.Install(m)

	//
	m.Publish([]string{"evars"}, "Doing install things...")
	network.Install(m)

	m.Publish([]string{"evars"}, "\n\r")
}
