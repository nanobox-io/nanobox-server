// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"fmt"

	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-router"

	"github.com/pagodabox/nanobox-server/config"
)

func logFatal(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	config.Logtap.Publish("deploy", logtap.CRITICAL, s)
}

func logError(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	config.Logtap.Publish("deploy", logtap.ERROR, s)
}

func logWarn(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	config.Logtap.Publish("deploy", logtap.WARNING, s)
}

func logInfo(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	config.Logtap.Publish("deploy", logtap.INFORMATIONAL, s)
}

func logDebug(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	config.Logtap.Publish("deploy", logtap.DEBUG, s)
}

func handleError(message string, err error) {
	errMessage := ""
	if err != nil {
		errMessage = err.Error()
	}

	logError("%s (%s)\n", message, errMessage)
	config.Router.Handler = router.FailedDeploy{}
}