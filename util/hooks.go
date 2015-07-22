// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package util

//
import (
	"encoding/json"
)

// Exec
func ExecHook(hook, container string, payload map[string]interface{}) ([]byte, error) {

	// marshal the payload
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return ExecInContainer(container, "/opt/bin/"+hook, string(b))
}

// Run
func RunHook(hook, container, img string, payload map[string]interface{}) ([]byte, error) {

	// marshal the payload
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return RunInContainer(container, img, "/opt/bin/"+hook, string(b))
}
