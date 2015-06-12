package api

import (
	"net/http"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/data"
)

// CreateEVar
func (api *API) CreateDeploy(rw http.ResponseWriter, req *http.Request) {
	config.Log.Debug("[NANOBOX :: API] Deploy create\n")

	sync := data.Sync{}
	api.Worker.QueueAndProcess(&sync)

	rw.Write([]byte("dude it worked!"))
}
