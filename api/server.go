package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/pat"

	"github.com/nanobox-core/nanobox-server/db"
)

//
type (

	// Server represents
	API struct {
		Server *Server
		Driver *db.Driver
	}

	// Server represents
	Server struct {
		Addr string
		host string
		port string
	}
)

//
var (
	debugging bool
)

// Init
func (api *API) Init(opts map[string]string, driver *db.Driver) int {
	fmt.Println("Initializing API...")

	debugging = (opts["debugging"] == "true")

	// = nanobox.db
	api.Driver = driver

	//
	api.Server = &Server{}
	api.Server.host = opts["host"]
	api.Server.port = opts["port"]
	api.Server.Addr = opts["host"] + ":" + opts["port"]

	//
	return 0
}

// registerRoutes
func InitRoutes(p *pat.Router, api *API) {

	// evars
	p.Delete("/evars/{slug}", handle(api.DeleteEVar, debugging))
	p.Put("/evars/{slug}", handle(api.UpdateEVar, debugging))
	p.Get("/evars/{slug}", handle(api.GetEVar, debugging))
	p.Post("/evars", handle(api.CreateEVar, debugging))
	p.Get("/evars", handle(api.ListEVars, debugging))
}

// handle
func handle(fn func(http.ResponseWriter, *http.Request), watch bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		if watch {
			fmt.Printf(`
Request:
--------------------------------------------------------------------------------
%#v

Response:
--------------------------------------------------------------------------------
%#v


`, req, w)

		}

		fn(w, req)
	}
}
