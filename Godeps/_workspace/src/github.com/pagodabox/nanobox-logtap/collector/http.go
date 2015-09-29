// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package collector

import (
	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/nanobox-logtap"
	"io"
	"io/ioutil"
	"net"
	"net/http"
)

// create and return a http handler that can be dropped into an api.
func GenerateHttpCollector(kind string, l *logtap.Logtap) http.HandlerFunc {
	headerName := "X-" + kind + "-Id"
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		logLevel := lumber.LvlInt(r.Header.Get("X-Log-Level"))
		header := r.Header.Get(headerName)
		if header == "" {
			header = kind
		}
		l.Publish(header, logLevel, string(body))
	}
}

func StartHttpCollector(kind, address string, l *logtap.Logtap) (io.Closer, error) {
	httpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	go http.Serve(httpListener, GenerateHttpCollector(kind, l))
	return httpListener, nil
}
