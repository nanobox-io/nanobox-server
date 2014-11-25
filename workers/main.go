package workers

import (
	"fmt"
	"github.com/nanobox-core/mist"
)

//
type (

	//
	Factory struct{}

	//
	Worker interface {
		Start(chan<- bool, *mist.Mist)
	}
)

//
var (
	debugging bool
	done      chan bool
)

// Init
func (w *Factory) Init(opts map[string]string) int {
	fmt.Println("Initializing workers...")

	debugging = (opts["debugging"] == "true")

	done = make(chan bool)

	//
	return 0
}

// Process
func Process(v Worker, m *mist.Mist) {

	fmt.Println("Proccessing worker...")

	//
	v.Start(done, m)

	// wait...
	<-done
}
