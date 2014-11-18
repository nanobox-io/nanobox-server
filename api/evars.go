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
func (api *API) ListEVars(res http.ResponseWriter, req *http.Request) {
	evars := &data.EVars{}
	evars.List(api.Driver)
	utils.WriteResponse(evars.EVars, res, http.StatusOK)
}

// CreateEVar
func (api *API) CreateEVar(res http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	utils.ParseBody(req, evar)
	evar.Create(api.Driver)
	utils.WriteResponse(evar, res, http.StatusCreated)

	//
	worker := workers.EVar{
		Action: "create",
		ID: evar.ID,
		Response: res,
	}

	workers.Process(worker)
}

// GetEVar
func (api *API) GetEVar(res http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Get(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, res, http.StatusOK)
}

// UpdateEVar
func (api *API) UpdateEVar(res http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	utils.ParseBody(req, evar)
	evar.Update(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, res, http.StatusOK)
}

// DeleteEVar
func (api *API) DeleteEVar(res http.ResponseWriter, req *http.Request) {
	evar := &data.EVar{}
	evar.Destroy(req.URL.Query().Get(":slug"), api.Driver)
	utils.WriteResponse(evar, res, http.StatusOK)
}
