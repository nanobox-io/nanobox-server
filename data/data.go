package data

import (
	"github.com/nanobox-core/scribble"
)

type (

	//
	Model interface {
		Collection() string
		Id() string
	}
)

//
var Driver *scribble.Driver

// List
func List(collection string, v interface{}) error {
	return Driver.Transact(scribble.Transaction{Action: "readall", Collection: collection, Container: &v})
}

// Save
func Save(v Model) error {
	return Driver.Transact(scribble.Transaction{Action: "write", Collection: v.Collection(), ResourceID: v.Id(), Container: &v})
}

// Get
func Get(v Model) error {
	return Driver.Transact(scribble.Transaction{Action: "read", Collection: v.Collection(), ResourceID: v.Id(), Container: &v})
}

// Update
func Update(v Model) error {
	return Save(v)
}

// Destroy
func Destroy(v Model) error {
	return Driver.Transact(scribble.Transaction{Action: "delete", Collection: v.Collection(), ResourceID: v.Id(), Container: &v})
}
