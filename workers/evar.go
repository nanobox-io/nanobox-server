package workers

import (
	"fmt"
	"time"

	"github.com/nanobox-core/mist"
	// "github.com/nanobox-core/nanobox-server/util"
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

	m.Publish([]string{"a"}, "A1")
	time.Sleep(2 * time.Second)

	m.Publish([]string{"a"}, "A2")
	time.Sleep(2 * time.Second)

	m.Publish([]string{"b"}, "B1")
	time.Sleep(1 * time.Second)

	m.Publish([]string{"a"}, "A3")
	time.Sleep(2 * time.Second)

	m.Publish([]string{"a"}, "A4")
	time.Sleep(2 * time.Second)

	m.Publish([]string{"b"}, "B2")
	time.Sleep(1 * time.Second)

	m.Publish([]string{"a"}, "done")
	time.Sleep(1 * time.Second)

	for i := 0; i < 3; i++ {
		m.Publish([]string{"b"}, "looping...")
		time.Sleep(1 * time.Second)
	}

	m.Publish([]string{"b"}, "done")
	time.Sleep(1 * time.Second)

	// container := util.Container{}
	// network := util.Network{}

	//
	// container.Install()

	//
	// network.Install()
}
