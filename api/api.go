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

	"github.com/pagodabox/nanobox-server/data"
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

	startup := data.Startup{}
	api.Worker.QueueAndProcess(&startup)

	//
	routes, err := api.registerRoutes()
	if err != nil {
		return err
	}

	//
	config.Log.Info("[NANOBOX :: API] Listening on port %v\n", port)

	// blocking...
	if err := http.ListenAndServe(":"+port, routes); err != nil {
		return err
	}

	return nil
}

// registerRoutes
func (api *API) registerRoutes() (*pat.Router, error) {
	config.Log.Debug("[NANOBOX :: API] Registering routes...\n")

	//
	router := pat.New()

	// will need a /services/ and /services/name
	router.Post("/deploys", api.handleRequest(api.CreateDeploy))
	router.Post("/update", api.handleRequest(api.UpdateImages))
	router.Get("/services", api.handleRequest(api.ListServices))
	return router, nil
}

// handleRequest
func (api *API) handleRequest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		config.Log.Debug(`
Request:
--------------------------------------------------------------------------------
%+v

`, req)

		//
		fn(rw, req)

		config.Log.Debug(`
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
