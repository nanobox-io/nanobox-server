// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"fmt"
	"regexp"
	"time"
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

func (s *ServiceStart) deployLog(message string) {
	config.Logtap.Publish("deploy", message)
}

func (s *ServiceStart) handleError(message string, err error) {
	s.deployLog(message)
	errMessage := ""
	if err == nil {
		errMessage = "noerror"
	} else {
		errMessage = err.Error()
	}
 	config.Log.Error("%s (%s)\n", message, errMessage)
}

func (s *ServiceStart) Process() {
	config.Log.Debug("[NANOBOX :: SYNC :: SERVICE] Started\n")

	s.Success = false
	// start the container
	image := regexp.MustCompile(`\d+`).ReplaceAllString(s.Uid, "")
	m := map[string]string{"uid": s.Uid}
	if image == "web" || image == "worker" || image == "tcp" {
		image = "code"
		m["code"] = "true"
	} else {
		m["service"] = "true"
	}
	config.Log.Debug("%#v %#v\n\n\n", image, m)
	container, err := tasks.CreateContainer("nanobox/"+image, m)
	if err != nil {
		s.handleError("[NANOBOX :: SYNC :: SERVICE] container create failed", err)
		return
	}

	addr := container.NetworkSettings.IPAddress

	h := hooky.Hooky{
		Host: addr,
		Port: "5540",
	}

	payload := map[string]interface{}{
		"boxfile": s.Boxfile.Parsed,
		"logvac_uri": config.LogtapURI,
		"uid": s.Uid,
	}

	pString, _ := json.Marshal(payload)
	time.Sleep(10 * time.Second)

	response, err := h.Run("code-configure", pString, "1")
	if err != nil || response.Exit != 0 {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC :: SERVICE] hook problem(%#v)", response), err)
		return
	}

	response, err = h.Run("code-start", "{}", "2")
	if err != nil {
		s.handleError(fmt.Sprintf("[NANOBOX :: SYNC :: SERVICE] hook problem(%#v)", response), err)
		return
	}

	if m["service"] == "true" {
		response, err = h.Run("code-environment", "{}", "3")
		if err != nil || response.Exit != 0 {
			s.handleError(fmt.Sprintf("[NANOBOX :: SYNC :: SERVICE] hook problem(%#v)", response), err)
			return
		}
		if err := json.Unmarshal([]byte(response.Out), &s.EnvVars); err != nil {
			s.handleError("[NANOBOX :: SYNC :: SERVICE] couldnt unmarshel evars from server", err)
			return
		}
	}

	s.Success = true
	config.Log.Debug("[NANOBOX :: SYNC :: SERVICE] service started perfectly(%#v)", s)
}