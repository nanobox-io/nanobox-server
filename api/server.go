package api

import(
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
    debugging string
  }
)

// Init
func (api *API) Init(opts map[string]string, driver *db.Driver) int {

	// = nanobox.db
	api.Driver = driver

	//
	api.Server = &Server{}

	api.Server.debugging = opts["debugging"]

  api.Server.host = opts["host"]
  api.Server.port = opts["port"]
  api.Server.Addr = opts["host"] + ":" + opts["port"]

	//
	return 0
}

// registerRoutes
func InitRoutes(p *pat.Router, api *API) {

	watch := (api.Server.debugging == "true")

  // evars
  p.Delete("/evars/{slug}", handle(api.DeleteEVar, watch))
  p.Put("/evars/{slug}", handle(api.UpdateEVar, watch))
  p.Get("/evars/{slug}", handle(api.GetEVar, watch))
  p.Post("/evars", handle(api.CreateEVar, watch))
  p.Get("/evars", handle(api.ListEVars, watch))
}

// handle
func handle(fn func (http.ResponseWriter, *http.Request), watch bool) http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {

  	if watch {
	  	fmt.Printf(`
Request:
--------------------------------------------------------------------------------
%#v

Response:
--------------------------------------------------------------------------------
%#v


`, req, w )

	  }

    fn(w, req)
  }
}
