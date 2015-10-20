// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package config

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	// "time"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/golang-hatchet"
	"github.com/nanobox-io/golang-mist"
	"github.com/nanobox-io/nanobox-logtap"
)

//
var (
	App        string
	LogtapHost string
	Ports      map[string]string
	IP         string

	Log        hatchet.Logger
	Logtap     *logtap.Logtap
	Mist       *mist.Mist
	LogHandler http.HandlerFunc
)

//
func init() {

	// create an error object
	var err error

	Log = lumber.NewConsoleLogger(lumber.INFO)

	//
	Ports = map[string]string{
		"api":    ":1757",
		"logtap": ":514",
		"mist":   ":1445",
		"router": "60000",
	}

	IP, err = externalIP()
	if err != nil {
		Log.Error("error: %s\n", err.Error())
	}

	LogtapHost = IP

	App, err = appName()
	// for err != nil {
	// 	Log.Error("error: %s\n", err.Error())
	// 	time.Sleep(time.Second)
	// 	App, err = appName()
	// }

	Mist = mist.New()
	Logtap = logtap.New(Log)
}

//
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			if strings.HasPrefix(ip.String(), "10") {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

//
func appName() (string, error) {
	files, err := ioutil.ReadDir("/vagrant/code/")
	if err != nil {
		return "", err
	}

	if len(files) < 1 || !files[0].IsDir() {
		return "", errors.New("There is no code in your /vagrant/code/ folder")
	}

	return files[0].Name(), nil
}
