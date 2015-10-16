// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util/fs"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
)

var developTex = sync.Mutex{}

func (api *API) Develop(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		config.Log.Debug("form parsing error: \n %s", err.Error())
	}
	// force the develop route to go into a dev1 container
	req.Form["container"] = []string{"dev1"}

	box := mergedBox()

	containerControl := false

	developTex.Lock()
	// if there is no dev1 it needs to be created and this thread needs to remember
	// to shut it down when its done conatinerControl is used for that purpose
	container, err := docker.GetContainer("dev1")
	if err != nil {
		containerControl = true
		cmd := []string{"/bin/sleep", "365d"}

		image := "nanobox/build"
		if stab := box.Node("build").StringValue("stability"); stab != "" {
			image = image + ":" + stab
		}

		container, err = docker.CreateContainer(docker.CreateConfig{Image: image, Category: "dev", UID: "dev1", Cmd: cmd})
		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}

		// run the default-user hook to get ssh keys setup
		out, err := script.Exec("default-user", "dev1", fs.UserPayload())
		if err != nil {
			config.Log.Debug("Failed script output: \n %s", out)
		}
		fmt.Println(string(out))
	}

	developTex.Unlock()

	api.Exec(rw, req)

	if containerControl {
		execWait.Wait()
		docker.RemoveContainer(container.ID)
	}
}


func mergedBox() (box boxfile.Boxfile) {
	box = boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")
	if out, err := script.Exec("boxfile", "build1", map[string]interface{}{}); err == nil {
		box.Merge(boxfile.New([]byte(out)))
	}
	return
}
