// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

//
import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Deploy struct {
	ID      string
	Payload map[string]interface{}
	Reset   bool
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Deploy) Process() {

	// clear the deploy log
	config.Logtap.Drains["history"].(*logtap.HistoricalDrain).ClearDeploy()

	// set routing to watch logs
	util.LogDebug(stylish.Bullet("Watching logs at /deploys..."))
	config.Router.Handler = router.DeployInProgress{}

	// remove all code containers
	util.LogInfo(stylish.Bullet("Removing containers from previous deploy..."))
	containers, _ := util.ListContainers("code", "build")
	for _, container := range containers {
		if err := util.RemoveContainer(container.Id); err != nil {
			util.HandleError(stylish.Error("Failed to remove old containers", err.Error()), "")
			j.updateStatus("errored")
			return
		}
	}

	// Make sure we have the directories
	util.LogDebug(stylish.Bullet("Ensure directories exist on host..."))
	if err := util.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()), "")
		j.updateStatus("errored")
		return
	}

	// wipe the previous deploy data if reset == true
	if j.Reset {
		util.LogInfo(stylish.Bullet("Resetting cache and code directories"))
		if err := util.Clean(); err != nil {
			util.LogInfo(stylish.Warning(fmt.Sprintf("Failed to reset cache and code directories:\n%v", err.Error())))
		}
	}

	// if the build image doesn't exist it needs to be downloaded
	if !util.ImageExists("nanobox/build") {
		util.LogInfo(stylish.Bullet("Pulling the latest build image (this will take awhile)... "))
		util.InstallImage("nanobox/build")
	}

	// create a build container
	util.LogInfo(stylish.Bullet("Creating build container..."))
	_, err := util.CreateBuildContainer("build1")
	if err != nil {
		util.HandleError(stylish.Error("Failed to create build container", err.Error()), "")
		// j.updateStatus("errored")
		return
	}

	// parse the boxfile
	util.LogDebug(stylish.Bullet("Parsing Boxfile..."))
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	// todo: move this to a function
	// define the deploy payload
	j.Payload = map[string]interface{}{
		"app":        config.App,
		"dns":        []string{config.App + ".nano.dev"},
		"env":        map[string]string{"APP_NAME": config.App},
		"port":       "8080",
		"boxfile":    box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	// run configure hook (blocking)
	if _, err := util.ExecHook("configure", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run sync hook (blocking)
	if _, err := util.ExecHook("sync", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run detect hook (blocking)
	if _, err := util.ExecHook("detect", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run setup hook (blocking)
	if _, err := util.ExecHook("setup", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// run boxfile hook (blocking)
	if out, err := util.ExecHook("boxfile", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return

		// if the hook runs succesfully merge the boxfiles
	} else {
		util.LogDebug(stylish.Bullet("Merging Boxfiles..."))
		util.LogDebug("BOXFILE STUFF! %v\n", string(out))
		box.Merge(boxfile.New([]byte(out)))
	}

	// add the missing storage nodes to the boxfile
	// todo: move this logic into the Boxfile project
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		if (name == "web" || name == "worker") && box.Node(node).Value("network_dirs") != nil {
			found := false
			for _, storage := range box.Node(node).Node("network_dirs").Nodes() {
				found = true
				if !box.Node(storage).Valid {
					box.Parsed[storage] = map[string]interface{}{"found": true}
				}
			}

			// if i dont find anything but they did have a network_dirs.. just try adding a new one
			if !found {
				if !box.Node("nfs1").Valid {
					box.Parsed["nfs1"] = map[string]interface{}{"found": true}
				}
			}
		}
	}

	// remove any containers no longer in the boxfile
	util.LogDebug(stylish.Bullet("Removing old containers..."))
	serviceContainers, _ := util.ListContainers("service")
	for _, container := range serviceContainers {
		if !box.Node(container.Labels["uid"]).Valid {
			util.RemoveContainer(container.Id)
		}
	}

	worker := util.NewWorker()
	worker.Blocking = true
	worker.Concurrent = true

	//
	serviceStarts := []*ServiceStart{}

	// build service containers according to boxfile
	// todo: move this logic into the Boxfile project
	util.LogInfo(stylish.Bullet("Creating new services..."))
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		if node != "nanobox" &&
			node != "global" &&
			node != "build" &&
			name != "web" &&
			name != "worker" {

			// container doesn't exist so we need to create it
			if _, err := util.GetContainer(node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					UID:     node,
					EVars:   map[string]string{},
				}

				serviceStarts = append(serviceStarts, &s)

				worker.Queue(&s)
			}
		}
	}

	util.LogInfo(stylish.Bullet("Starting services"))
	util.LogDebug("SERVICES! %#v\n", serviceStarts)
	worker.Process()

	// ensure all services started correctly before continuing
	for _, starts := range serviceStarts {
		if !starts.Success {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to start %v", starts.UID), err.Error()), "")
			// j.updateStatus("errored")
			return
		}
	}

	// grab the environment data from all service containers
	evars := j.Payload["env"].(map[string]string)

	//
	serviceEnvs := []*ServiceEnv{}

	util.LogDebug(stylish.Bullet("Creating new services..."))
	serviceContainers, _ = util.ListContainers("service")
	for _, container := range serviceContainers {

		s := ServiceEnv{UID: container.Labels["uid"]}

		serviceEnvs = append(serviceEnvs, &s)

		worker.Queue(&s)
	}

	worker.Process()

	for _, env := range serviceEnvs {
		if !env.Success {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", env.UID), err.Error()), "")
			// j.updateStatus("errored")
			return
		}

		for key, val := range env.EVars {
			evars[strings.ToUpper(env.UID+"_"+key)] = val
		}
	}

	j.Payload["env"] = evars

	// run prepare hook (blocking)
	if _, err := util.ExecHook("prepare", "build1", j.Payload); err != nil {
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

	// run cleanup hook (blocking)
	if _, err := util.ExecHook("cleanup", "build1", j.Payload); err != nil {
		util.LogInfo("ERROR %v\n", err)
		// j.updateStatus("errored")
		return
	}

	// build new code containers
	// todo: move the Boxfile logic into the Boxfile project
	codeServices := []*ServiceStart{}
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
		if name == "web" || name == "worker" {
			if _, err := util.GetContainer(node); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					UID:     node,
					EVars:   evars,
				}

				codeServices = append(codeServices, &s)

				worker.Queue(&s)
			}
		}
	}

	worker.Process()

	for _, serv := range codeServices {
		if !serv.Success {
			util.HandleError("A Service was not started correctly ("+serv.UID+")", err.Error())
			// j.updateStatus("errored")
			return
		}
	}

	// todo: we need to launch a new "code" image and run these in there, probably just using a RunHook instead
	// run before deploy hooks
	for _, node := range box.Nodes() {
		bd := box.Node(node).Value("before_deploy")
		bda := box.Node(node).Value("before_deploy_all")
		if bd != nil || bda != nil {

			// run before deploy hook (blocking)
			if _, err := util.ExecHook("before_deploy", "web1", map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda}); err != nil {
				util.LogInfo("ERROR %v\n", err)
				// j.updateStatus("errored")
				return
			}
		}
	}

	// set routing to web components
	util.LogDebug(stylish.Bullet("Configure routing..."))
	if container, err := util.GetContainer("web1"); err == nil {
		dc, _ := util.InspectContainer(container.Id)

		config.Router.AddTarget("/", "http://"+dc.NetworkSettings.IPAddress+":8080")
		config.Router.Handler = nil
	}

	// todo: we need to launch a new "code" image and run these in there, probably just using a RunHook instead
	// after deploy hooks
	for _, node := range box.Nodes() {
		ad := box.Node(node).Value("after_deploy")
		ada := box.Node(node).Value("after_deploy_all")
		if ad != nil || ada != nil {

			// run after deploy hook (blocking)
			if _, err := util.ExecHook("after_deploy", "web1", map[string]interface{}{"after_deploy": ad, "after_deploy_all": ada}); err != nil {
				util.LogInfo("ERROR %v\n", err)
				// j.updateStatus("errored")
				return
			}
		}
	}

	j.updateStatus("complete")
}

// todo: this is used in other places, this should be moved to util
// updateStatus
func (j *Deploy) updateStatus(status string) {
	config.Mist.Publish([]string{"job", "deploy"}, fmt.Sprintf(`{"model":"Deploy", "action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, j.ID, status))
}
