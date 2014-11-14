package workers

import (
	"fmt"
	"time"
)

type (

	//
	EVar struct {
		ID     string
		Action string
	}
)

// Start
func (w EVar) Start(done chan<- bool) {

	//
	if debugging {
		fmt.Printf("EVar worker '%v' is working job '%v'...\n.", w.ID, w.Action)
	}

	//
	switch w.Action {

	//
	case "create":
		w.create()

	//
	case "update":
		w.update()
	}

	// release...
	done <- true
}

//
func (w EVar) create() {
	fmt.Println("CREATE!")
	time.Sleep(time.Second * 1)
}

//
func (w EVar) update() {
	fmt.Println("UPDATE!")
	time.Sleep(time.Second * 1)
}
