package data

import (
	"regexp"
	"encoding/json"

	"github.com/hookyd/go-client"
	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/pagodabox/nanobox-server/config"
)

type ServiceStart struct {
	Boxfile boxfile.Boxfile
	Uid     string
	Success bool
	EnvVars map[string]string
}

// hooks
	// configure
	// start
	// environment

func (s *ServiceStart) Process() {
	config.Log.Debug("[NANOBOX :: SYNC :: SERVICE] Started\n")
	ch := make(chan string)
	defer close(ch)

	go func() {
		for data := range ch {
			config.Mist.Publish([]string{"sync"}, data)
		}
	}()

	s.Success = false
	// start the container
	image := regexp.MustCompile(`\d+`).ReplaceAllString(s.Uid, "")
	m := map[string]string{"uid": s.Uid}
	if image == "web" || image == "worker" || image == "tcp" {
		image = "base"
		m["code"] = "true"
	} else {
		m["service"] = "true"
	}
	config.Log.Debug("%#v %#v\n\n\n", image, m)
	container, err := tasks.CreateContainer("nanobox/"+image, m)
	if err != nil {
		config.Log.Error("[NANOBOX :: SYNC :: SERVICE] container create failed:%s", err.Error())
		return
	}

	addr := container.NetworkSettings.IPAddress

	h := hooky.Hooky{
		Host: addr,
		Port: "1234", // dont know the port
	}

	payload := map[string]interface{}{
		"boxfile": s.Boxfile.Parsed,
		"logvac_uri": config.LogtapURI,
		"uid": s.Uid,
	}

	pString, _ := json.Marshal(payload)

	response, err := h.Run("configure", pString, "1")
	if err != nil {
		config.Log.Error("[NANOBOX :: SYNC :: SERVICE] hook response(%#v) err(%#v)", response, err)
		// return
	}

	response, err = h.Run("start", "{}", "2")
	if err != nil {
		config.Log.Error("[NANOBOX :: SYNC :: SERVICE] hook response(%#v) err(%#v)", response, err)
		// return
	}

	if m["service"] == "true" {
		response, err = h.Run("environment", "{}", "3")
		if err != nil {
			config.Log.Error("[NANOBOX :: SYNC :: SERVICE] hook response(%#v) err(%#v)", response, err)
			// return
		}
		if err := json.Unmarshal([]byte(response.Out), &s.EnvVars); err != nil {
			ch <- "couldnt un marshel evars from server"
			// return
		}
	}

	s.Success = true
	config.Log.Debug("[NANOBOX :: SYNC :: SERVICE] service started perfectly(%#v)", s)
}