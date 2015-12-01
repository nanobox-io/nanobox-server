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
	"sync"
	"time"
	"bufio"

	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/fs"
)

var execWait = sync.WaitGroup{}

var execKeys = map[string]string{}

func (api *API) LibDirs(rw http.ResponseWriter, req *http.Request) {
	writeBody(fs.LibDirs(), rw, http.StatusOK)
}

func (api *API) FileChange(rw http.ResponseWriter, req *http.Request) {
	fn := func(file string) {
		<-time.After(time.Second)
		fs.Touch(file)
	}

	// read the file from the header so the old way works
	file := req.FormValue("filename")
	if file != "" {
		go fn(file)
	}

	// read all the body parts and post the files
	body := bufio.NewReader(req.Body)
	
	for line, _, err := body.ReadLine(); err != io.EOF; line, _, err = body.ReadLine() {
		fmt.Println("file:", string(line))
		if len(line) != 0 {
			go fn(string(line))
		}
	}

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
	// if name == "" {
	// 	name = "dev1"
	// }

	conn, br, err := rw.(http.Hijacker).Hijack()
	if err != nil {
		config.Log.Debug("exec hijack error: %s", err.Error())
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
		config.Log.Debug("exec get container: %s", err.Error())
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
		docker.RunExec(exec, io.MultiReader(br, conn), conn, conn)
	}
}

// necessary for anything using a windowing system through the exec.
func (api *API) ResizeExec(rw http.ResponseWriter, req *http.Request) {
	pid := req.FormValue("pid")
	// give it 10 seconds to show up
	for i := 0; (i < 20 || execKeys[pid] == ""); i++ {
			<-time.After(1 * time.Second)
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
	fmt.Println("resize error:", err)
}
