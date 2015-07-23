// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/util"
)

type ServiceEnv struct {
	deploy Deploy

	EVars   map[string]string
	UID     string
	Success bool
}

func (j *ServiceEnv) Process() {
	j.Success = false

	// run environment hook (blocking)
	if out, err := util.ExecHook("environment", j.UID, nil); err != nil {
		util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", j.UID), err.Error()), "")
		// util.UpdateStatus(j.deploy, "errored")
		return
	} else {
		if err := json.Unmarshal(out, &j.EVars); err != nil {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", j.UID), err.Error()), "")
			// util.UpdateStatus(j.deploy, "errored")
			return
		}
	}

	j.Success = true
}
