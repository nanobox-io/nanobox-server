package api

import (
	// "fmt"
	"io/ioutil"
	"net/http"

	"github.com/nanobox-core/nanobox-server/data"
)

// ListEVars
func (api *API) ListEVars(w http.ResponseWriter, req *http.Request) {

	evars := &data.EVars{}
	evars.List(api.Driver)
}

// CreateEVar
func (api *API) CreateEVar(w http.ResponseWriter, req *http.Request) {

	//
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	evar := &data.EVar{}
	evar.Create(b, api.Driver)
}

// GetEVar
func (api *API) GetEVar(w http.ResponseWriter, req *http.Request) {
	resourceID := req.URL.Query().Get(":slug")

	evar := &data.EVar{}
	evar.Get(resourceID, api.Driver)
}

// UpdateEVar
func (api *API) UpdateEVar(w http.ResponseWriter, req *http.Request) {

	//
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	resourceID := req.URL.Query().Get(":slug")

	evar := &data.EVar{}
	evar.Update(resourceID, b, api.Driver)
}

// DeleteEVar
func (api *API) DeleteEVar(w http.ResponseWriter, req *http.Request) {
	resourceID := req.URL.Query().Get(":slug")

	evar := &data.EVar{}
	evar.Destroy(resourceID, api.Driver)
}
