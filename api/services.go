package api

import (
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
)

// CreateEVar
func (api *API) ListServices(rw http.ResponseWriter, req *http.Request) {
	config.Log.Debug("[NANOBOX :: API] List Services\n")

	containers, _ := tasks.ListContainers()
	data := []map[string]string{}
	for _, container := range containers {
		dc, _ := tasks.GetDetailedContainer(container.Id)

		c := container.Labels
		c["ip"] = dc.NetworkSettings.IPAddress
		tunnelPort := config.Router.GetLocalPort(c["ip"]+":22")
		c["tunnel_port"] = strconv.Itoa(tunnelPort)
		data = append(data, c)
	}

	j, err := json.Marshal(data)
	if err != nil {
		config.Log.Error("[NANOBOX :: API] list services (%s)", err.Error())
	}
	rw.Write(j)
}
