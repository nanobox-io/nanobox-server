// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.


//
package data

import (
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/golang-scribble"
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
	config.Log.Debug("[NANOBOX :: DATA] Save resource: %+v\n", v)

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
