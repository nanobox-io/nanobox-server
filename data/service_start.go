// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hookyd/go-client"
	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
)

type ServiceStart struct {
	Boxfile boxfile.Boxfile
	Uid     string
	Success bool
	EnvVars map[string]string
}

func (s *ServiceStart) Process() {
	if s.EnvVars == nil {
		s.EnvVars = map[string]string{}
	}
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

	// create hooky connection
	h, err := hooky.New(addr, "5540")
	if err != nil {
		handleError("[NANOBOX :: SYNC :: "+s.Uid+"] hooky connection failed", err)
	}

	// payload
	payload := map[string]interface{}{
		"boxfile":    s.Boxfile.Parsed,
		"logtap_uri": config.LogtapURI,
		"uid":        s.Uid,
	}

	// adds to the payload storage information if storage is required
	needsStorage := false
	storage := map[string]map[string]string{}
	for key, val := range s.EnvVars {
		matched, _ := regexp.MatchString(`NFS\d+_HOST`, key)
		if matched {
			needsStorage = true
			nfsUid := regexp.MustCompile(`_HOST`).ReplaceAllString(key, "")
			host := map[string]string{"host": val}
			storage[nfsUid] = host
		}
	}

	if needsStorage {
		payload["storage"] = storage
	}

	pString, err := json.Marshal(payload)
	if err != nil {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] json error", s.Uid), err)
		return
	}

	logInfo("[NANOBOX :: SYNC :: %s] running configure hook", s.Uid)
	response, err := h.Run("configure", pString, "1")
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] hook problem(%#v)", s.Uid, response), err)
		return
	}
	logDebug("[NANOBOX :: SYNC :: %s] Hook Response (configure): %+v\n", s.Uid, response)

	logInfo("[NANOBOX :: SYNC :: %s] running start hook", s.Uid)
	response, err = h.Run("start", "{}", "2")
	if err != nil {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] hook problem(%#v)", s.Uid, response), err)
		return
	}
	logDebug("[NANOBOX :: SYNC :: %s] Hook Response (start): %+v\n", s.Uid, response)

	// if we make it to the end it was a success!
	s.Success = true
	logDebug("[NANOBOX :: SYNC :: %s] service started perfectly(%#v)", s.Uid ,s)
}
