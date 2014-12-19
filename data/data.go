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
	t := scribble.Transaction{Operation: "readall", Collection: collection, Container: &v}
	return config.Scribble.Transact(t)
}

// Save
func Save(v Model) error {
	config.Log.Debug("[NANOBOX :: DATA] Save resource: %+v\n", v)

	t := scribble.Transaction{Operation: "write", Collection: v.Collection(), RecordID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}

// Get
func Get(v Model) error {
	t := scribble.Transaction{Operation: "read", Collection: v.Collection(), RecordID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}

// Update
func Update(v Model) error {
	return Save(v)
}

// Destroy
func Destroy(v Model) error {
	t := scribble.Transaction{Operation: "delete", Collection: v.Collection(), RecordID: v.Id(), Container: &v}
	return config.Scribble.Transact(t)
}
