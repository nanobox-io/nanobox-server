package data

import (
	"encoding/JSON"
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
func (e *EVars) List(driver *db.Driver) error {
	trans := db.Transaction{Action: "readall", Collection: "evars", Container: e.EVars}

	driver.Transact(trans)

	return nil
}

// Create
func (e *EVar) Create(body []byte, driver *db.Driver) error {

	if err := json.Unmarshal(body, e); err != nil {
		panic(err)
	}

	e.ID = uuid.New()
	e.CreatedAt = time.Now()

	e.save(driver)

	return nil
}

// Get
func (e *EVar) Get(resourceID string, driver *db.Driver) error {
	trans := db.Transaction{Action: "read", Collection: "evars", Resource: resourceID, Container: e}
	driver.Transact(trans)

	return nil
}

// Update
func (e *EVar) Update(resourceID string, body []byte, driver *db.Driver) error {
	trans := db.Transaction{Action: "read", Collection: "evars", Resource: resourceID, Container: e}
	driver.Transact(trans)

	if err := json.Unmarshal(body, e); err != nil {
		panic(err)
	}

	e.save(driver)

	return nil
}

// Destroy
func (e *EVar) Destroy(resourceID string, driver *db.Driver) error {
	trans := db.Transaction{Action: "delete", Collection: "evars", Resource: resourceID}
	driver.Transact(trans)

	return nil
}

// private

// save
func (e *EVar) save(driver *db.Driver) {
	e.UpdatedAt = time.Now()

	trans := db.Transaction{Action: "write", Collection: "evars", Resource: e.ID, Container: e}
	driver.Transact(trans)
}
