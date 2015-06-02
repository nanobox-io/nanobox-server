// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.


//
package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/pat"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/worker"
)

// structs
type (

	//
	API struct {
		Worker *worker.Worker
	}
)

func Init() *API {
	//
	api := &API{
		Worker: worker.New(),
	}

	return api
}

// Start
func (api *API) Start(port string) error {
	config.Log.Info("[NANOBOX :: API] Starting server...\n")

	//
	routes, err := api.registerRoutes()
	if err != nil {
		return err
	}

	//
	config.Log.Info("[NANOBOX :: API] Listening on port %v\n", port)

	// blocking...
	http.Handle("/", routes)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	return nil
}

// registerRoutes
func (api *API) registerRoutes() (*pat.Router, error) {
	config.Log.Debug("[NANOBOX :: API] Registering routes...\n")

	//
	router := pat.New()

	// evars
	router.Delete("/evars/{slug}", api.handleRequest(api.DeleteEVar))
	router.Put("/evars/{slug}", api.handleRequest(api.UpdateEVar))
	router.Get("/evars/{slug}", api.handleRequest(api.GetEVar))
	router.Post("/evars", api.handleRequest(api.CreateEVar))
	router.Get("/evars", api.handleRequest(api.ListEVars))

	return router, nil
}

// handleRequest
func (api *API) handleRequest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		config.Log.Info(`
Request:
--------------------------------------------------------------------------------
%+v

`, req)

		//
		fn(rw, req)

		config.Log.Info(`
Response:
--------------------------------------------------------------------------------
%+v

`, rw)
	}
}

// helpers

// newUUID
func newUUID() string {
	return uuid.New()
}

// parseBody
func parseBody(req *http.Request, v interface{}) error {

	//
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	defer req.Body.Close()

	//
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

// writeBody
func writeBody(v interface{}, rw http.ResponseWriter, status int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(b)

	return nil
}
