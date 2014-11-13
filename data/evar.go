package data

import (
	// "fmt"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/nanobox-core/nanobox-server/db"
)

type (

	//
	EVars struct {
		EVars []EVar
	}

	//
	EVar struct {
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
	EVarCreateOptions struct {
		AppID string `json:"app_id,omitempty"`
		Title string `json:"title,omitempty"`
		Value string `json:"value,omitempty"`
	}

	//
	EVarUpdateOptions struct {
		Title string `json:"title,omitempty"`
		Value string `json:"value,omitempty"`
	}
)

// List
func (e *EVars) List(driver *db.Driver) {
	trans := db.Transaction{Action: "readall", Collection: "evars", Container: &e.EVars}
	driver.Transact(trans)
}

// Create
func (e *EVar) Create(driver *db.Driver) {

	e.ID = uuid.New()
	e.CreatedAt = time.Now()

	trans := db.Transaction{Action: "write", Collection: "evars", Resource: e.ID, Container: e}
	e.save(trans, driver)
}

// Get
func (e *EVar) Get(resourceID string, driver *db.Driver) {
	trans := db.Transaction{Action: "read", Collection: "evars", Resource: resourceID, Container: e}
	driver.Transact(trans)
}

// Update
func (e *EVar) Update(resourceID string, driver *db.Driver) {
	trans := db.Transaction{Action: "read", Collection: "evars", Resource: resourceID, Container: e}
	e.save(trans, driver)
}

// Destroy
func (e *EVar) Destroy(resourceID string, driver *db.Driver) {
	trans := db.Transaction{Action: "delete", Collection: "evars", Resource: resourceID}
	driver.Transact(trans)
}

// private

// save
func (e *EVar) save(trans db.Transaction, driver *db.Driver) {
	e.UpdatedAt = time.Now()

	driver.Transact(trans)
}
