package controllers

import (
	// "fmt"
	"io/ioutil"
	"net/http"

	"github.com/nanobox-core/nanobox-server/models"
)

//
type EVars struct {
	model models.EVar
}

// Index
func (c *EVars) Index(w http.ResponseWriter, req *http.Request) {
	c.model.List()
}

// Create
func (c *EVars) Create(w http.ResponseWriter, req *http.Request) {

	//
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	c.model.Create(b)
}

// Show
func (c *EVars) Show(w http.ResponseWriter, req *http.Request) {
	c.model.Get(req.URL.Query().Get(":slug"))
}

// Update
func (c *EVars) Update(w http.ResponseWriter, req *http.Request) {

	//
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	c.model.Update(req.URL.Query().Get(":slug"), b)
}

// Delete
func (c *EVars) Delete(w http.ResponseWriter, req *http.Request) {
	c.model.Destroy(req.URL.Query().Get(":slug"))
}
