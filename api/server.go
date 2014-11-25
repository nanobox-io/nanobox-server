package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/pat"

	"github.com/nanobox-core/mist"
	"github.com/nanobox-core/nanobox-server/db"
	"github.com/nanobox-core/nanobox-server/utils"
)

const (
	DefaultHost = "localhost"
	DefaultPort = "1757"
	DefaultAddr = DefaultHost + ":" + DefaultPort
)

//
type (

	// Server represents
	API struct {
		Server *Server
		Driver *db.Driver
		Mist   *mist.Mist
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
func (api *API) Init(opts map[string]string, driver *db.Driver, mist *mist.Mist) int {
	fmt.Println("Initializing API...")

	debugging = (opts["debugging"] == "true")

	// = nanobox.db
	api.Driver = driver

	// = nanobox.mist
	api.Mist = mist

	//
	api.Server = &Server{
		host: utils.SetOption(opts["host"], DefaultHost),
		port: utils.SetOption(opts["port"], DefaultHost),
		Addr: utils.SetOption((opts["host"] + ":" + opts["port"]), DefaultAddr),
	}

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
%+v

Response:
--------------------------------------------------------------------------------
%+v


`, req, w)

		}

		fn(w, req)
	}
}
