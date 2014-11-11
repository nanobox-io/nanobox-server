package models

import (
	"encoding/JSON"
	"fmt"
	// "io"
	// "net/http"
	"time"

	"github.com/nanobox-core/nanobox-server/db"

	"code.google.com/p/go-uuid/uuid"
)

type (

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

//
var EVarChan chan int

//
func init() {
  EVarChan = make(chan int, 1)
}

// List
func (m *EVar) List() {
  var evars []EVar
  do(func(){db.ReadAll("evars", &evars, EVarChan)})
}

// Create
func (m *EVar) Create(body []byte) {

	evar := &EVar{}

	if err := json.Unmarshal(body, evar); err != nil {
		panic(err)
	}

	evar.ID = uuid.New()
	evar.CreatedAt = time.Now()

	evar.save()
}

// Get
func (m *EVar) Get(id string) {
	evar := &EVar{}
	do(func(){db.Read("evars", id, evar, EVarChan)})
}

// Update
func (m *EVar) Update(id string, body []byte) {
	evar := &EVar{}

	db.Read("evars", id, evar, EVarChan)

	if err := json.Unmarshal(body, evar); err != nil {
		panic(err)
	}

	evar.save()
}

// Destroy
func (m *EVar) Destroy(id string) {
	do(func(){db.Delete("evars", id, EVarChan)})
}

// Save
func (m *EVar) Save() {
	m.save()
}

// Count
func (m *EVar) Count() {

}

// private

// save
func (m *EVar) save() {
	m.UpdatedAt = time.Now()

	do(func(){db.Write("evars", m.ID, m, EVarChan)})
}

// do
func do(action func()) {
  go action()
  <- EVarChan
}
