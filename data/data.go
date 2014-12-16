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
	Driver.Transact(scribble.Transaction{Action: "readall", Collection: collection, Container: &v})
	return nil
}

// Save
func Save(v Model) error {
	Driver.Transact(scribble.Transaction{Action: "write", Collection: v.Collection(), Resource: v.Id(), Container: &v})
	return nil
}

// Get
func Get(v Model) error {
	Driver.Transact(scribble.Transaction{Action: "read", Collection: v.Collection(), Resource: v.Id(), Container: &v})
	return nil
}

// Update
func Update(v Model) error {
	Save(v)
	return nil
}

// Destroy
func Destroy(v Model) error {
	Driver.Transact(scribble.Transaction{Action: "delete", Collection: v.Collection(), Resource: v.Id(), Container: &v})
	return nil
}
