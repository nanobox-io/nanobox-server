package data

import (
	"time"

	"github.com/nanobox-core/nanobox-server/config"
	"github.com/nanobox-core/nanobox-server/tasks"
)

//
type EVar struct {
	AppID     string    `json:"app_id"`     //
	CreatedAt time.Time `json:"created_at"` //
	ID        string    `json:"id"`         //
	Internal  bool      `json:"internal"`   //
	ServiceID string    `json:"service_id"` //
	Title     string    `json:"title"`      //
	UpdatedAt time.Time `json:"updated_at"` //
	Value     string    `json:"value"`      //
}

//
func (e *EVar) Collection() string {
	return "evars"
}

//
func (e *EVar) Id() string {
	return e.ID
}

//
func (e *EVar) Process() {

	ch := make(chan string)
	defer close(ch)

	go func() {
		for data := range ch {
			config.Mist.Publish([]string{e.Collection()}, data)
		}
	}()

	container := tasks.Container{}
	network := tasks.Network{}

	//
	ch <- "Doing container things..."
	container.Install(ch)

	//
	ch <- "Doing install things..."
	network.Install(ch)

	//
	ch <- "DONE"
}
