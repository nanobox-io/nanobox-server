package api

import (
	// "encoding/JSON"
	// "fmt"
	"net/http"

	"github.com/nanobox-core/nanobox-server/data"
	"github.com/nanobox-core/nanobox-server/utils"
	"github.com/nanobox-core/nanobox-server/workers"
)

// ListEVars
func (api *API) ListEVars(w http.ResponseWriter, req *http.Request) {
	evars := &data.EVars{}
	evars.List(api.Driver)
	utils.WriteResponse(evars.EVars, w, http.StatusOK)
}

// CreateEVar
func (api *API) CreateEVar(w http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	utils.ParseBody(req, evar)
	evar.Create(api.Driver)
	utils.WriteResponse(evar, w, http.StatusCreated)

	//
	worker := workers.EVar{
		ID:     evar.ID,
		Action: "create",
	}

	workers.Run(worker)
}

// GetEVar
func (api *API) GetEVar(w http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Get(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, w, http.StatusOK)
}

// UpdateEVar
func (api *API) UpdateEVar(w http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	utils.ParseBody(req, evar)
	evar.Update(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, w, http.StatusOK)
}

// DeleteEVar
func (api *API) DeleteEVar(w http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Destroy(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, w, http.StatusOK)
}
