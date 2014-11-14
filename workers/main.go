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

//
func (w *Factory) Init(opts map[string]string) int {
	fmt.Println("Initializing workers...")

	debugging = (opts["debugging"] == "true")

	done = make(chan bool, 5)

	//
	return 0
}

//
func Run(v Worker) {

	//
	go v.Start(done)

	// wait...
	<-done
}
