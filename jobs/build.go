// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

import (
	"fmt"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Build struct {
	ID      string
	Payload map[string]interface{}
	Reset   bool
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Build) Process() {

	// parse the boxfile
	util.LogInfo(stylish.Bullet("Parsing Boxfile..."))
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// define the build payload
	j.Payload = map[string]interface{}{
		"app":        config.App,
		"dns":        []string{config.App + ".nano.dev"},
		"env":        map[string]string{"APP_NAME": config.App},
		"port":       "8080",
		"boxfile":    box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	// run sync hook (blocking)
	if _, err := util.ExecHook("sync", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run build hook (blocking)
	if _, err := util.ExecHook("build", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run publish hook (blocking)
	if _, err := util.ExecHook("publish", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// find code containers and run the restart hook
	// todo: maybe move this into it's own job
	containers, _ := util.ListContainers("code")
	for _, container := range containers {
		uid := container.Labels["uid"]

		util.LogInfo(stylish.Bullet(fmt.Sprintf("Restarting app in %s container...", uid)))

		// restart payload
		payload := map[string]interface{}{
			"boxfile":    box.Node(uid).Parsed,
			"logtap_uri": config.LogtapURI,
			"uid":        uid,
		}

		// run restart hook (blocking)
		if _, err := util.ExecHook("restart", uid, payload); err != nil {
			util.LogInfo("ERROR %v\n", err)
			// j.updateStatus("errored")
			return
		}

	}

	j.updateStatus("complete")
}

//
func (j *Build) updateStatus(status string) {
	config.Mist.Publish([]string{"job", "deploy"}, fmt.Sprintf(`{"model":"Build", "action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, j.ID, status))
}
