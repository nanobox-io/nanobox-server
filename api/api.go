package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/pat"

	"code.google.com/p/go-uuid/uuid"

	"github.com/nanobox-core/hatchet"
	// "github.com/nanobox-core/logtap"
	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/data"
	"github.com/nanobox-core/nanobox-server/worker"
	// "github.com/nanobox-core/router"
	"github.com/nanobox-core/scribble"
)

// structs
type (

	//
	API struct {
		Log hatchet.Logger
		// logtap *logtap.Logtap  //
		Mist *mist.Mist //
		// router *router.Router  //
		Driver *scribble.Driver //
		Worker *worker.Worker   //
	}
)

// Start
func (api *API) Start(port string) error {
	fmt.Println("Starting server...")

	//
	data.Driver = api.Driver

	//
	routes, err := api.RegisterRoutes()
	if err != nil {
		return err
	}

	//
	fmt.Printf("Nanobox listening at %v\n", port)

	// blocking...
	http.Handle("/", routes)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		return err
	}

	return nil
}

// RegisterRoutes
func (api *API) RegisterRoutes() (*pat.Router, error) {
	fmt.Println("Registering routes...")

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
	return func(w http.ResponseWriter, req *http.Request) {

		api.Log.Info(`
Request:
--------------------------------------------------------------------------------
%+v

Response:
--------------------------------------------------------------------------------
%+v

`, req, w)

		fn(w, req)
	}
}

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
