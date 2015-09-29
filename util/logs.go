// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"fmt"

	"github.com/nanobox-io/nanobox-router"
	"github.com/nanobox-io/nanobox-server/config"
)

// LogDebug
func LogDebug(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", 1, fmt.Sprintf(f, v...))
}

// LogInfo
func LogInfo(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", 2, fmt.Sprintf(f, v...))
}

// LogWarn
func LogWarn(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", 3, fmt.Sprintf(f, v...))
}

// LogError
func LogError(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", 4, fmt.Sprintf(f, v...))
}

// LogFatal
func LogFatal(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", 5, fmt.Sprintf(f, v...))
}

// HandleError
func HandleError(msg string) {
	LogDebug(msg)
	router.ErrorHandler = router.FailedDeploy{}
}
