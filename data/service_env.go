// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"encoding/json"
	"fmt"

	"github.com/hookyd/go-client"
)

type ServiceEnv struct {
	Uid     string
	Addr    string
	Success bool
	EnvVars map[string]string
}

func (s *ServiceEnv) Process() {
	if s.EnvVars == nil {
		s.EnvVars = map[string]string{}
	}

	s.Success = false

	// create hooky connection
	h, err := hooky.New(s.Addr, "5540")
	if err != nil {
		handleError("[NANOBOX :: SYNC :: "+s.Uid+"] hooky connection failed", err)
	}

	logInfo("[NANOBOX :: SYNC :: %s] running environment hook", s.Uid)
	response, err := h.Run("environment", "{}", "3")
	if err != nil || response.Exit != 0 {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] hook problem(%#v)", s.Uid, response), err)
		return
	}
	logDebug("[NANOBOX :: SYNC :: %s] Hook Response (environment): %+v\n", s.Uid, response)
	if err := json.Unmarshal([]byte(response.Out), &s.EnvVars); err != nil {
		handleError(fmt.Sprintf("[NANOBOX :: SYNC :: %s] couldnt unmarshel evars from server", s.Uid), err)
		return
	}
	s.Success = true
}
