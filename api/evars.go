package api

import (
	"net/http"
	"time"

	"github.com/nanobox-core/nanobox-server/data"
	// "github.com/nanobox-core/nanobox-server/worker"
)

// ListEVars
func (api *API) ListEVars(rw http.ResponseWriter, req *http.Request) {
	evars := &[]data.EVar{}

	//
	if err := data.List("evars", evars); err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	//
	writeBody(evars, rw, http.StatusOK)
}

// CreateEVar
func (api *API) CreateEVar(rw http.ResponseWriter, req *http.Request) {

	//
	evar := &data.EVar{
		ID:        newUUID(),
		CreatedAt: time.Now(),
	}

	//
	if err := parseBody(req, evar); err != nil {
		rw.WriteHeader(422)
		rw.Write([]byte(""))
		return
	}

	if err := data.Save(evar); err != nil {
		rw.WriteHeader(422)
		rw.Write([]byte(""))
		return
	}

	//
	writeBody(evar, rw, http.StatusCreated)
}

// GetEVar
func (api *API) GetEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{
		ID:        req.URL.Query().Get(":slug"),
		UpdatedAt: time.Now(),
	}

	//
	if err := data.Get(evar); err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	//
	writeBody(evar, rw, http.StatusOK)
}

// UpdateEVar
func (api *API) UpdateEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{
		ID:        req.URL.Query().Get(":slug"),
		UpdatedAt: time.Now(),
	}

	//
	if err := data.Get(evar); err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	//
	if err := parseBody(req, evar); err != nil {
		rw.WriteHeader(422)
		rw.Write([]byte(""))
		return
	}

	//
	if err := data.Update(evar); err != nil {
		rw.WriteHeader(422)
		rw.Write([]byte(""))
		return
	}

	//
	writeBody(evar, rw, http.StatusAccepted)
}

// DeleteEVar
func (api *API) DeleteEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{
		ID: req.URL.Query().Get(":slug"),
	}

	//
	if err := data.Destroy(evar); err != nil {
		rw.WriteHeader(422)
		rw.Write([]byte(""))
		return
	}

	//
	writeBody(evar, rw, http.StatusOK)
}
