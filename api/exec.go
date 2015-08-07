// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

// starter is a bandaid for a problem with docker
// where it will not send the first data until it recieves
// something on stdin
type starter struct {
	thing   io.Reader
	started bool
}

func (s *starter) Read(p []byte) (int, error) {
	if !s.started {
		s.started = true
		p[0] = '\n'
		return len([]byte("\n")), nil
	}
	return s.thing.Read(p)
}

func (api *API) Exec(rw http.ResponseWriter, req *http.Request) {
	util.RemoveContainer("exec1")
	conn, _, err := rw.(http.Hijacker).Hijack()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	defer conn.Close()

	cmd := []string{"/bin/bash"}
	additionalCmd := req.FormValue("cmd")
	if additionalCmd != "" {
		cmd = append(cmd, "-c", additionalCmd)
	}

	container, err := util.CreateContainer(util.CreateConfig{Category: "exec", Name: "exec1", Cmd: cmd})
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}

	go util.AttachToContainer(container.ID, conn, conn, conn)

	forwards := []string{}
	if req.FormValue("forward") != "" {
		forwards = append(forwards, strings.Split(req.FormValue("forward"), ",")...)
	}

	box := mergedBox()
	if boxForwards, ok := box.Node("console").Value("forwards").([]interface{}); ok {
		for _, boxFInterface := range boxForwards {
			if boxForward, ok := boxFInterface.(string); ok {
				forwards = append(forwards, boxForward)
			}
		}
	}

	// maybe add a forward port mapping
	for _, rule := range forwards {
		strSlice := strings.Split(rule, ":")
		switch len(strSlice) {
		case 1:
			portInt, err := strconv.Atoi(strSlice[0])
			if err != nil {
				continue
			}
			port, err := config.Router.AddForward("enter-"+rule, portInt, container.NetworkSettings.IPAddress+":"+strSlice[0])
			if err != nil {
				fmt.Fprintf(conn, "could not establish forward: %s\r\n", rule)
				continue
			}
			if port != portInt {
				fmt.Fprintf(conn, "requested port(%d) taken, we gave you %d instead\r\n", portInt, port)
			}
			defer config.Router.RemoveForward("enter-" + rule)
		case 2:
			portInt, _ := strconv.Atoi(strSlice[0])
			port, err := config.Router.AddForward("enter-"+rule, portInt, container.NetworkSettings.IPAddress+":"+strSlice[1])
			if err != nil {
				fmt.Fprintf(conn, "could not establish forward: %s\r\n", rule)
				continue
			}
			if port != portInt {
				fmt.Fprintf(conn, "requested port(%d) taken, we gave you %d instead\r\n", portInt, port)
			}
			defer config.Router.RemoveForward("enter-" + rule)
		}
	}

	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})
	// s := &starter{thing: conn}

	util.WaitContainer(container.ID)
	util.RemoveContainer(container.ID)
}

func (api *API) KillExec(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("signal recieved: %s\n", req.FormValue("signal"))
	err := util.KillContainer("exec1", req.FormValue("signal"))
	fmt.Println(err)
}

func (api *API) ResizeExec(rw http.ResponseWriter, req *http.Request) {
	h, _ := strconv.Atoi(req.FormValue("h"))
	w, _ := strconv.Atoi(req.FormValue("w"))
	if h == 0 || w == 0 {
		return
	}
	err := util.ResizeContainerTTY("exec1", h, w)
	fmt.Println(err)
}

func mergedBox() (box boxfile.Boxfile) {
	box = boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")
	if out, err := util.ExecHook("boxfile", "build1", map[string]interface{}{}); err == nil {
		box.Merge(boxfile.New([]byte(out)))
	}
	return
}
