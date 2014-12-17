package data

import (
	"github.com/nanobox-core/nanobox-server/config"
	"github.com/nanobox-core/scribble"
)

type (

	//
	Model interface {
		Collection() string
		Id() string
	}
)

// List
func List(collection string, v interface{}) error {
	t := scribble.Transaction{Action: "readall", Collection: collection, Container: &v}
	return config.Scribble.Transact(t)
}

// Save
func Save(v Model) error {
	t := scribble.Transaction{Action: "write", Collection: v.Collection(), ResourceID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}

// Get
func Get(v Model) error {
	t := scribble.Transaction{Action: "read", Collection: v.Collection(), ResourceID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}

// Update
func Update(v Model) error {
	return Save(v)
}

// Destroy
func Destroy(v Model) error {
	t := scribble.Transaction{Action: "delete", Collection: v.Collection(), ResourceID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}
