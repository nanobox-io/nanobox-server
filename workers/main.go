package workers

import (
	"fmt"
)

//
type (

	//
	Factory struct{}

	//
	Worker interface {
		Start(chan<- bool)
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
func Process(v Worker) {

	//
	go v.Start(done)

	// wait...
	<-done
}
