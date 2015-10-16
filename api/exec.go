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
	"sync"
	"time"

	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/fs"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
)

var execWait = sync.WaitGroup{}

var execKeys = map[string]string{}

func (api *API) LibDirs(rw http.ResponseWriter, req *http.Request) {
	writeBody(fs.LibDirs(), rw, http.StatusOK)
}

func (api *API) FileChange(rw http.ResponseWriter, req *http.Request) {
	fs.Touch(req.FormValue("filename"))
	writeBody(nil, rw, http.StatusOK)
}

// func (api *API) KillRun(rw http.ResponseWriter, req *http.Request) {
// 	fmt.Printf("signal recieved: %s\n", req.FormValue("signal"))
// 	err := docker.KillContainer("exec1", req.FormValue("signal"))
// 	fmt.Println(err)
// }

// func (api *API) ResizeRun(rw http.ResponseWriter, req *http.Request) {
// 	if req.FormValue("container") != "" {
// 		api.ResizeExec(rw, req)
// 		return
// 	}
// 	h, _ := strconv.Atoi(req.FormValue("h"))
// 	w, _ := strconv.Atoi(req.FormValue("w"))
// 	if h == 0 || w == 0 {
// 		return
// 	}
// 	err := docker.ResizeContainerTTY("exec1", h, w)
// 	fmt.Println(err)
// }

// proxy an exec request to docker. This allows us to have the same
// exec power but with added security.
func (api *API) Exec(rw http.ResponseWriter, req *http.Request) {
	execWait.Add(1)
	util.Lock()
	defer execWait.Done()
	defer util.Unlock()
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

	container, err := docker.GetContainer(name)
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}

	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})

	exec, err := docker.CreateExec(container.ID, cmd, true, true, true)
	if err == nil {
		pid := req.FormValue("pid")
		execKeys[pid] = exec.ID
		defer delete(execKeys, pid)
		docker.RunExec(exec, conn, conn, conn)
	}
}

// necessary for anything using a windowing system through the exec.
func (api *API) ResizeExec(rw http.ResponseWriter, req *http.Request) {
	pid := req.FormValue("pid")
	if execKeys[pid] == "" {
		time.Sleep(1 * time.Second)
	}
	if pid == "" || execKeys[pid] == "" {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	h, _ := strconv.Atoi(req.FormValue("h"))
	w, _ := strconv.Atoi(req.FormValue("w"))
	if h == 0 || w == 0 {
		return
	}

	err := docker.ResizeExecTTY(execKeys[pid], h, w)
	fmt.Println(err)
}