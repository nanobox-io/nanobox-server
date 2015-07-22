// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"fmt"
	"regexp"

	"github.com/samalba/dockerclient"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
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

	var ci *dockerclient.ContainerInfo
	var err error

	j.Success = false

	util.LogInfo(stylish.Bullet(fmt.Sprintf("Starting %v...", j.UID)))

	// start the container
	image := regexp.MustCompile(`\d+`).ReplaceAllString(j.UID, "")

	switch image {

	//
	case "web", "worker", "tcp":
		ci, err = util.CreateCodeContainer(j.UID)
		if err != nil {
			util.HandleError(fmt.Sprintf("Failed to create %v", j.UID), "")
			// util.UpdateStatus(j.deploy, "errored")
			return
		}

		//
	default:
		ci, err = util.CreateServiceContainer(j.UID, "nanobox/"+image)
		if err != nil {
			util.HandleError(fmt.Sprintf("Failed to create %v", j.UID), "")
			// util.UpdateStatus(j.deploy, "errored")
			return
		}
	}

	if err := util.StartContainer(ci.Id); err != nil {
		util.HandleError(fmt.Sprintf("Failed to create %v", j.UID), "")
		// util.UpdateStatus(j.deploy, "errored")
		return
	}

	// payload
	payload := map[string]interface{}{
		"boxfile":    j.Boxfile.Parsed,
		"logtap_uri": config.LogtapURI,
		"uid":        j.UID,
	}

	// adds to the payload storage information if storage is required
	needsStorage := false
	storage := map[string]map[string]string{}
	for key, val := range j.EVars {
		matched, _ := regexp.MatchString(`NFS\d+_HOST`, key)
		if matched {
			needsStorage = true
			nfsUid := regexp.MustCompile(`_HOST`).ReplaceAllString(key, "")
			host := map[string]string{"host": val}
			storage[nfsUid] = host
		}
	}

	if needsStorage {
		payload["storage"] = storage
	}

	// run configure hook (blocking)
	if _, err := util.ExecHook("configure", j.UID, payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// util.UpdateStatus(j.deploy, "errored")
		return
	}

	// run start hook (blocking)
	if _, err := util.ExecHook("start", j.UID, payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// util.UpdateStatus(j.deploy, "errored")
		return
	}

	// if we make it to the end it was a success!
	j.Success = true

	util.LogDebug("   [√] SUCCESS\n")
}