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

var createTex = sync.Mutex{}

var execWait = sync.WaitGroup{}

var execKeys = map[string]string{}

func (api *API) Run(rw http.ResponseWriter, req *http.Request) {
	name := req.FormValue("container")
	if name != "" {
		api.Exec(rw, req)
		return
	}

	box := mergedBox()

	containerControl := false

	createTex.Lock()
	// if there is no exec 1 it needs to be created and this thread needs to remember
	// to shut it down when its done conatinerControl is used for that purpose
	container, err := docker.GetContainer("exec1")
	if err != nil {
		containerControl = true
		cmd := []string{"/bin/sleep", "365d"}

		image := "nanobox/build"
		if stab := box.Node("build").StringValue("stability"); stab != "" {
			image = image + ":" + stab
		}

		container, err = docker.CreateContainer(docker.CreateConfig{Image: image, Category: "exec", UID: "exec1", Cmd: cmd})
		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}

		// run the default-user hook to get ssh keys setup
		out, err := script.Exec("default-user", "exec1", fs.UserPayload())
		if err != nil {
			util.LogDebug("Failed script output: \n %s", out)
		}
		fmt.Println(string(out))
	}

	createTex.Unlock()

	api.Exec(rw, req)

	if containerControl {
		execWait.Wait()
		docker.RemoveContainer(container.ID)
	}
}

func (api *API) LibDirs(rw http.ResponseWriter, req *http.Request) {
	writeBody(fs.LibDirs(), rw, http.StatusOK)
}

func (api *API) FileChange(rw http.ResponseWriter, req *http.Request) {
	fs.Touch(req.FormValue("filename"))
	writeBody(nil, rw, http.StatusOK)
}

func (api *API) KillRun(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("signal recieved: %s\n", req.FormValue("signal"))
	err := docker.KillContainer("exec1", req.FormValue("signal"))
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
	err := docker.ResizeContainerTTY("exec1", h, w)
	fmt.Println(err)
}

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
		execKeys[name] = exec.ID
		defer delete(execKeys, name)
		docker.RunExec(exec, conn, conn, conn)
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

	err := docker.ResizeExecTTY(execKeys[name], h, w)
	fmt.Println(err)
}

func mergedBox() (box boxfile.Boxfile) {
	box = boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")
	if out, err := script.Exec("boxfile", "build1", map[string]interface{}{}); err == nil {
		box.Merge(boxfile.New([]byte(out)))
	}
	return
}
