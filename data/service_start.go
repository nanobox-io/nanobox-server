// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"fmt"
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

func (s *ServiceStart) Process() {
	logInfo("[NANOBOX :: SYNC :: %s] Started", s.Uid)

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
	container, err := tasks.CreateContainer("nanobox/"+image, m)
	if err != nil {
		handleError("[NANOBOX :: SYNC :: "+s.Uid+"] container create failed", err)
		return
	}

	addr := container.NetworkSettings.IPAddress

	h, err := hooky.New(addr, "5540")
	if err != nil {
		handleError("[NANOBOX :: SYNC :: "+s.Uid+"] hooky connection failed", err)
	}

	payload := map[string]interface{}{
		"boxfile": s.Boxfile.Parsed,
		"logtap_uri": config.LogtapURI,
		"uid": s.Uid,
	}

	pString, _ := json.Marshal(payload)

	logInfo("[NANOBOX :: SYNC :: %s] running configure hook", s.Uid)
	response, err := h.Run("configure", pString, "1")
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: SERVICE] hook problem(%#v)", response), err)
		return
	}
	logDebug("[NANOBOX :: SYNC :: %s] Hook Response (configure): %+v\n", s.Uid, response)

	logInfo("[NANOBOX :: SYNC :: %s] running start hook", s.Uid)
	response, err = h.Run("start", "{}", "2")
	if err != nil {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] hook problem(%#v)", s.Uid, response), err)
		return
	}
	logDebug("[NANOBOX :: SYNC :: %s] Hook Response (start): %+v\n", response)

	if m["service"] == "true" {
		logInfo("[NANOBOX :: SYNC :: %s] running environment hook", s.Uid)
		response, err = h.Run("environment", "{}", "3")
		if err != nil || response.Exit != 0 {
			handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] hook problem(%#v)", s.Uid, response), err)
			return
		}
		logDebug("[NANOBOX :: SYNC :: %s] Hook Response (environment): %+v\n", s.Uid, response)
		if err := json.Unmarshal([]byte(response.Out), &s.EnvVars); err != nil {
			handleError("[NANOBOX :: SYNC :: SERVICE] couldnt unmarshel evars from server", err)
			return
		}
	}

	s.Success = true
	logDebug("[NANOBOX :: SYNC :: SERVICE] service started perfectly(%#v)", s)
}