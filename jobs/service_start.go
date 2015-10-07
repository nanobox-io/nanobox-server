// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"regexp"
	"strings"

	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
)

//
type ServiceStart struct {
	deploy Deploy

	Boxfile boxfile.Boxfile
	EVars   map[string]string
	Success bool
	UID     string
}

//
func (j *ServiceStart) Process() {
	// var ci *docker.Container
	var err error

	j.Success = false

	createConfig := util.CreateConfig{UID: j.UID, Name: j.Boxfile.StringValue("name")}

	image := regexp.MustCompile(`\d+`).ReplaceAllString(j.UID, "")
	if image == "web" || image == "worker" || image == "tcp" || image == "udp" {
		createConfig.Category = "code"
		image = "code"
	} else {
		createConfig.Category = "service"
	}

	if !util.ImageExists("nanobox/" + image) {
		util.LogInfo(stylish.SubBullet("- Pulling the latest %s image (this may take awhile)... ", image))
		util.InstallImage("nanobox/" + image)
	}	

	extra := strings.Trim(strings.Join([]string{j.Boxfile.VersionValue("version"), j.Boxfile.StringValue("stability")}, "-"), "-")
	if extra != "" {
		image = image + ":" + extra
	}

	createConfig.Image = "nanobox/" + image

	util.LogDebug(stylish.SubBullet("- Image name: %v", createConfig.Image))

	util.LogInfo(stylish.SubBullet("- Creating %v container", j.UID))

	// start the container
	if _, err = util.CreateContainer(createConfig); err != nil {
		util.HandleError(stylish.ErrorHead("Failed to create %v container", j.UID))
		util.HandleError(stylish.ErrorBody(err.Error()))
		util.UpdateStatus(&j.deploy, "errored")
		return
	}

	// payload
	payload := map[string]interface{}{
		"platform":    "local",
		"boxfile":     j.Boxfile.Parsed,
		"logtap_host": config.LogtapHost,
		"uid":         j.UID,
	}

	// adds to the payload storage information if storage is required
	needsStorage := false
	storage := map[string]map[string]string{}
	for key, val := range j.EVars {
		matched, _ := regexp.MatchString(`NFS\d+_HOST`, key)
		if matched {
			needsStorage = true
			nfsUid := strings.ToLower(regexp.MustCompile(`_HOST`).ReplaceAllString(key, ""))
			host := map[string]string{"host": val}
			storage[nfsUid] = host
		}
	}

	if needsStorage {
		payload["storage"] = storage
	}

	// run configure hook (blocking)
	if data, err := util.ExecHook("default-configure", j.UID, payload); err != nil {
		util.LogDebug("Failed Hook Output:\n%s\n", data)
		util.HandleError(stylish.Error("Configure hook failed", err.Error()))
		util.UpdateStatus(&j.deploy, "errored")
		return
	}

	util.LogInfo(stylish.SubBullet("- Starting %v service", j.UID))
	
	// run start hook (blocking)
	if data, err := util.ExecHook("default-start", j.UID, payload); err != nil {
		util.LogDebug("Failed Hook Output:\n%s\n", data)
		util.HandleError(stylish.Error("Start hook failed", err.Error()))
		util.UpdateStatus(&j.deploy, "errored")
		return
	}

	// if we make it to the end it was a success!
	j.Success = true

	util.LogDebug("   [√] SUCCESS\n")
}
