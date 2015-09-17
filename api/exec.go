// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

var execKeys = map[string]string{}

func (api *API) Run(rw http.ResponseWriter, req *http.Request) {
	name := req.FormValue("container")
	if name != "" {
		api.Exec(rw, req)
		return
	}

	// if an exec1 already exists we need to drop into that
	// container instead of failing or destroying the existing
	// container
	container, err := util.GetContainer("exec1")
	if err == nil {
		api.Exec(rw, req)
		return
	}	

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

	container, err = util.CreateContainer(util.CreateConfig{Image: "nanobox/build", Category: "exec", Name: "exec1", Cmd: cmd})
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
			// fromPort, toIp, toPort string
			err := util.AddForward(strSlice[0], container.NetworkSettings.IPAddress, strSlice[0])
			if err != nil {
				fmt.Fprintf(conn, "could not establish forward: %s\r\n", rule)
				continue
			}
			defer util.RemoveForward(container.NetworkSettings.IPAddress)
		case 2:
			err := util.AddForward(strSlice[0], container.NetworkSettings.IPAddress, strSlice[1])
			if err != nil {
				fmt.Fprintf(conn, "could not establish forward: %s\r\n", rule)
				continue
			}
			defer util.RemoveForward(container.NetworkSettings.IPAddress)
		}
	}

	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})

	util.WaitContainer(container.ID)
	util.RemoveContainer(container.ID)
}

func (api *API) LibDirs(rw http.ResponseWriter, req *http.Request) {
	writeBody(util.LibDirs(), rw, http.StatusOK)
}

func (api *API) FileChange(rw http.ResponseWriter, req *http.Request) {
	util.Touch(req.FormValue("filename"))
	writeBody(nil, rw, http.StatusOK)
}

func (api *API) KillRun(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("signal recieved: %s\n", req.FormValue("signal"))
	err := util.KillContainer("exec1", req.FormValue("signal"))
	fmt.Println(err)
}

func (api *API) ResizeRun(rw http.ResponseWriter, req *http.Request) {
	if req.FormValue("container") != "" {
		api.ResizeExec(rw, req)
		return
	}
	h, _ := strconv.Atoi(req.FormValue("h"))
	w, _ := strconv.Atoi(req.FormValue("w"))
	if h == 0 || w == 0 {
		return
	}
	err := util.ResizeContainerTTY("exec1", h, w)
	fmt.Println(err)
}

// proxy an exec request to docker. This allows us to have the same
// exec power but with added security.
func (api *API) Exec(rw http.ResponseWriter, req *http.Request) {
	name := req.FormValue("container")
	if name == "" {
		name = "exec1"
	}

	conn, _, err := rw.(http.Hijacker).Hijack()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	defer conn.Close()

	cmd := []string{"/bin/bash"}
	if additionalCmd := req.FormValue("cmd"); additionalCmd != "" {
		cmd = append(cmd, "-c", additionalCmd)
	}

	container, err := util.GetContainer(name)
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}

	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})

	exec, err := util.CreateExec(container.ID, cmd, true, true, true)
	if err == nil {
		execKeys[name] = exec.ID
		defer delete(execKeys, name)
		util.RunExec(exec, conn, conn, conn)
	}
}

// necessary for anything using a windowing system through the exec.
func (api *API) ResizeExec(rw http.ResponseWriter, req *http.Request) {
	name := req.FormValue("container")
	if execKeys[name] == "" {
		time.Sleep(1 * time.Second)
	}
	if name == "" || execKeys[name] == "" {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	h, _ := strconv.Atoi(req.FormValue("h"))
	w, _ := strconv.Atoi(req.FormValue("w"))
	if h == 0 || w == 0 {
		return
	}

	err := util.ResizeExecTTY(execKeys[name], h, w)
	fmt.Println(err)
}

func mergedBox() (box boxfile.Boxfile) {
	box = boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")
	if out, err := util.ExecHook("boxfile", "build1", map[string]interface{}{}); err == nil {
		box.Merge(boxfile.New([]byte(out)))
	}
	return
}
