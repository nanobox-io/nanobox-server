package api

import (
	"net/http"

	"github.com/nanobox-core/nanobox-server/data"
	"github.com/nanobox-core/nanobox-server/workers"
	"github.com/nanobox-core/utils"
)

// ListEVars
func (api *API) ListEVars(rw http.ResponseWriter, req *http.Request) {
	evars := &data.EVars{}
	evars.List(api.DB)
	utils.WriteResponse(evars.EVars, rw, http.StatusOK)

	//
	worker := workers.EVar{
		Action: "list",
		ID:     "n",
	}

	//
	go workers.Process(worker, api.Mist)
}

// CreateEVar
func (api *API) CreateEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	utils.ParseBody(req, evar)
	evar.Create(api.DB)
	utils.WriteResponse(evar, rw, http.StatusCreated)
}

// GetEVar
func (api *API) GetEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Get(req.URL.Query().Get(":slug"), api.DB)
	utils.WriteResponse(evar, rw, http.StatusOK)
}

// UpdateEVar
func (api *API) UpdateEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Get(req.URL.Query().Get(":slug"), api.DB)
	utils.ParseBody(req, evar)
	evar.Update(req.URL.Query().Get(":slug"), api.DB)
	utils.WriteResponse(evar, rw, http.StatusOK)
}

// DeleteEVar
func (api *API) DeleteEVar(rw http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Destroy(req.URL.Query().Get(":slug"), api.DB)
	utils.WriteResponse(evar, rw, http.StatusOK)
}
