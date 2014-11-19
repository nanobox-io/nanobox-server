package workers

import (
	"fmt"

	"github.com/nanobox-core/nanobox-server/mist"
	"github.com/nanobox-core/nanobox-server/util"
)

type (

	//
	EVar struct {
		Action string
		ID     string
	}
)

// Start
func (w EVar) Start(done chan<- bool) {

	//
	if debugging {
		fmt.Printf("EVar worker '%v' is working job '%v'...\n\n", w.ID, w.Action)
	}

	// subscribe
	mist.Subscribe(`{command:"subscribe", filters:["evar"]}`)

	//
	switch w.Action {

	//
	case "list":
		w.list()
	}

	// unsubscribe
	mist.Unsubscribe(`{command:"unsubscribe", filters:["evar"]}`)

	// release...
	done <- true
}

//
func (w EVar) list() {

	container := util.Container{}
	network := util.Network{}

	//
	container.Install()

	//
	network.Install()
}
