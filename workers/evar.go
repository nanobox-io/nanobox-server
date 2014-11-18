package workers

import (
	"fmt"
	"net/http"
	"time"
)

type (

	//
	EVar struct {
		Action string
		ID string
		Response http.ResponseWriter
	}
)

// Start
func (w EVar) Start(done chan<- bool) {

	//
	if debugging {
		fmt.Printf("EVar worker '%v' is working job '%v'...\n.", w.ID, w.Action)
	}

	// subscribe

	//
	switch w.Action {

	//
	case "create":
		w.create()

	//
	case "update":
		w.update()
	}

	// unsubscribe

	// release...
	done <- true
}

//
func (w EVar) create() {
	fmt.Println("CREATE!")

	//
	time.Sleep(time.Second * 1)
	fmt.Println("UPDATE 1")
	w.Response.Write([]byte(`{"message": "update 1"}`))
	w.Response.(http.Flusher).Flush()

  //
	time.Sleep(time.Second * 1)
	fmt.Println("UPDATE 2")
	w.Response.Write([]byte(`{"message": "update 2"}`))
	w.Response.(http.Flusher).Flush()

  //
	time.Sleep(time.Second * 1)
	fmt.Println("UPDATE 3")
	w.Response.Write([]byte(`{"message": "update 3"}`))
	w.Response.(http.Flusher).Flush()

	//
	time.Sleep(time.Second * 1)
	fmt.Println("UPDATE 4")
	w.Response.Write([]byte(`{"message": "update 4"}`))
	w.Response.(http.Flusher).Flush()

	//
	time.Sleep(time.Second * 1)
	fmt.Println("DONE")
	w.Response.Write([]byte(`{"message": "done"}`))
	w.Response.(http.Flusher).Flush()
}

//
func (w EVar) update() {
	fmt.Println("UPDATE!")
	time.Sleep(time.Second * 1)
}
