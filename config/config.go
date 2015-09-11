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
	"time"

	"github.com/jcelliott/lumber"

	"github.com/pagodabox/golang-hatchet"
	"github.com/pagodabox/golang-mist"
	"github.com/pagodabox/nanobox-logtap"
	logapi "github.com/pagodabox/nanobox-logtap/api"
	"github.com/pagodabox/nanobox-logtap/archive"
	"github.com/pagodabox/nanobox-logtap/collector"
	"github.com/pagodabox/nanobox-logtap/drain"
	"github.com/pagodabox/nanobox-router"
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

	Log.Info("LogtapHost: " + LogtapHost)
	App, err = appName()
	for err != nil {
		Log.Error("error: %s\n", err.Error())
		time.Sleep(time.Second)
		App, err = appName()
	}

	// create new router
	err = router.StartHTTP(":"+Ports["router"])
	if err != nil {
		Log.Error("error: %s\n", err.Error())
	}
	
	// create a new mist and start listening for messages at *:1445
	Mist = mist.New()
	Mist.Listen(Ports["mist"])

	// create new logtap; // we don't need to defer a close here, because this want
	// to live as long as the server
	Logtap = logtap.New(Log)
	// defer Logtap.Close()

	//
	console := drain.AdaptLogger(Log)
	Logtap.AddDrain("console", console)

	// define logtap collectors/drains; we don't need to defer Close() anything here,
	// because these want to live as long as the server
	if _, err := collector.SyslogUDPStart("app", Ports["logtap"], Logtap); err != nil {
		panic(err)
	}

	//
	if _, err := collector.SyslogTCPStart("app", Ports["logtap"], Logtap); err != nil {
		panic(err)
	}

	// we will be adding a 0 to the end of the logtap port because we cant have 2 tcp listeneres
	// on the same port
	if _, err := collector.StartHttpCollector("deploy", Ports["logtap"]+"0", Logtap); err != nil {
		panic(err)
	}

	//
	db, err := archive.NewBoltArchive("/tmp/bolt.db")
	LogHandler = logapi.GenerateArchiveEndpoint(db)

	//
	Logtap.AddDrain("historical", db.Write)
	Logtap.AddDrain("mist", drain.AdaptPublisher(Mist))
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

	// for _, file := range files {
	// 	Log.Info("%s: %s\n\n", file.Name(), file.IsDir())
	// }

	if len(files) < 1 || !files[0].IsDir() {
		return "", errors.New("There is no code in your /vagrant/code/ folder")
	}

	return files[0].Name(), nil
}
