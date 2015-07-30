// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"fmt"

	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
)

// LogFatal
func LogFatal(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", logtap.FATAL, fmt.Sprintf(f, v...))
}

// LogError
func LogError(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", logtap.ERROR, fmt.Sprintf(f, v...))
}

// LogWarn
func LogWarn(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", logtap.WARN, fmt.Sprintf(f, v...))
}

// LogInfo
func LogInfo(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", logtap.INFO, fmt.Sprintf(f, v...))
}

// LogDebug
func LogDebug(f string, v ...interface{}) {
	config.Logtap.Publish("deploy", logtap.DEBUG, fmt.Sprintf(f, v...))
}

// HandleError
func HandleError(msg string) {
	LogError(msg)
	config.Router.Handler = router.FailedDeploy{}
}
