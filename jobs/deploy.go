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
	"strings"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-golang-stylish"
	// "github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

//
type Deploy struct {
	ID      string
	Reset   bool
	Sandbox bool

	payload map[string]interface{}
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Deploy) Process() {

	// clear the deploy log
	// config.Logtap.Drains["history"].(*logtap.HistoricalDrain).ClearDeploy()

	// set routing to watch logs
	util.LogDebug(stylish.Bullet("Watching logs at /deploys..."))
	config.Router.Handler = router.DeployInProgress{}

	// remove all code containers
	util.LogInfo(stylish.Bullet("Removing containers from previous deploy..."))

	// might as well remove bootstraps and execs too
	containers, _ := util.ListContainers("code", "build", "bootstrap", "exec")
	for _, container := range containers {
		if err := util.RemoveContainer(container.ID); err != nil {
			util.HandleError(stylish.Error("Failed to remove old containers", err.Error()))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	// Make sure we have the directories
	util.LogDebug(stylish.Bullet("Ensure directories exist on host..."))
	if err := util.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// wipe the previous deploy data if reset == true
	if j.Reset {
		util.LogInfo(stylish.Bullet("Resetting cache and code directories"))
		if err := util.Clean(); err != nil {
			util.LogInfo(stylish.Warning(fmt.Sprintf("Failed to reset cache and code directories:\n%v", err.Error())))
		}
	}

	// parse the boxfile
	util.LogDebug(stylish.Bullet("Parsing Boxfile..."))
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	image := "nanobox/build"
	if stab := box.Node("build").StringValue("stability"); stab != "" {
		image = image + ":" + stab
	}
	// if the build image doesn't exist it needs to be downloaded
	if !util.ImageExists(image) {
		util.LogInfo(stylish.Bullet("Pulling the latest build image (this will take awhile)... "))
		util.InstallImage(image)
	}

	// create a build container
	util.LogInfo(stylish.Bullet("Creating build container..."))

	_, err := util.CreateContainer(util.CreateConfig{Image: image, Category: "build", Name: "build1"})
	if err != nil {
		util.HandleError(stylish.Error("Failed to create build container", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}


	// define the deploy payload
	j.payload = map[string]interface{}{
		"app":        config.App,
		"dns":        []string{config.App + ".nano.dev"},
		"port":       "8080",
		"boxfile":    box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	evar := map[string]string{}
	if box.Node("env").Valid {
		for key, val := range box.Node("env").Parsed {
			if str, ok := val.(string); ok {
				evar[key] = str
			}
		}
	}

	evar["APP_NAME"] = config.App
	j.payload["env"] = evar

	// run configure hook (blocking)
	if _, err := util.ExecHook("configure", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run configure hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run detect hook (blocking)
	if _, err := util.ExecHook("detect", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run detect hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run sync hook (blocking)
	if _, err := util.ExecHook("sync", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run sync hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run setup hook (blocking)
	if _, err := util.ExecHook("setup", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run setup hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run boxfile hook (blocking)
	if !box.Node("build").BoolValue("disable_engine_boxfile") {
		if out, err := util.ExecHook("boxfile", "build1", j.payload); err != nil {
			util.HandleError(stylish.Error("Failed to run boxfile hook", err.Error()))
			util.UpdateStatus(j, "errored")
			return

			// if the hook runs succesfully merge the boxfiles
		} else {
			util.LogDebug(stylish.Bullet("Merging Boxfiles..."))
			util.LogDebug("BOXFILE STUFF! %v\n", string(out))
			box.Merge(boxfile.New([]byte(out)))
		}
	}

	// add the missing storage nodes to the boxfile
	box.AddStorageNode()

	// remove any containers no longer in the boxfile
	util.LogDebug(stylish.Bullet("Removing old containers..."))
	serviceContainers, _ := util.ListContainers("service")
	for _, container := range serviceContainers {
		if !box.Node(container.Config.Labels["uid"]).Valid {
			config.Router.RemoveForward(container.Config.Labels["uid"])
			util.RemoveContainer(container.ID)
		}
	}

	worker := util.NewWorker()
	worker.Blocking = true
	worker.Concurrent = true

	//
	serviceStarts := []*ServiceStart{}

	// build service containers according to boxfile
	util.LogInfo(stylish.Bullet("Creating new services..."))
	for _, node := range box.Nodes("service") {
		if _, err := util.GetContainer(node); err != nil {
			// container doesn't exist so we need to create it
			s := ServiceStart{
				Boxfile: box.Node(node),
				UID:     node,
				EVars:   map[string]string{},
			}

			serviceStarts = append(serviceStarts, &s)

			worker.Queue(&s)
		}
	}

	util.LogInfo(stylish.Bullet("Starting services"))
	util.LogDebug("SERVICES! %#v\n", serviceStarts)
	worker.Process()

	// ensure all services started correctly before continuing
	for _, starts := range serviceStarts {
		if !starts.Success {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to start %v", starts.UID), ""))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	// grab the environment data from all service containers
	evars := j.payload["env"].(map[string]string)

	//
	serviceEnvs := []*ServiceEnv{}

	util.LogDebug(stylish.Bullet("Creating new services..."))
	serviceContainers, _ = util.ListContainers("service")
	for _, container := range serviceContainers {

		s := ServiceEnv{UID: container.Config.Labels["uid"]}

		serviceEnvs = append(serviceEnvs, &s)

		worker.Queue(&s)
	}

	worker.Process()

	for _, env := range serviceEnvs {
		if !env.Success {
			util.HandleError(stylish.Error(fmt.Sprintf("Failed to configure %v's environment variables", env.UID), ""))
			util.UpdateStatus(j, "errored")
			return
		}

		for key, val := range env.EVars {
			evars[strings.ToUpper(env.UID+"_"+key)] = val
		}
	}

	j.payload["env"] = evars

	// run prepare hook (blocking)
	if _, err := util.ExecHook("prepare", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run prepare hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run build hook (blocking)
	if _, err := util.ExecHook("build", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run build hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run publish hook (blocking)
	if _, err := util.ExecHook("publish", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run publish hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run cleanup hook (blocking)
	if _, err := util.ExecHook("cleanup", "build1", j.payload); err != nil {
		util.HandleError(stylish.Error("Failed to run cleanup hook", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// we will only create new code nodes if we are not
	// in a sandbox environment
	if !j.Sandbox {
		// build new code containers
		codeServices := []*ServiceStart{}
		for _, node := range box.Nodes("code") {
			if _, err := util.GetContainer(node); err != nil {
				// container doesn't exist so we need to create it
				s := ServiceStart{
					Boxfile: box.Node(node),
					UID:     node,
					EVars:   evars,
				}

				codeServices = append(codeServices, &s)

				worker.Queue(&s)
			}
		}

		worker.Process()

		for _, serv := range codeServices {
			if !serv.Success {
				util.HandleError("A Service was not started correctly (" + serv.UID + ")")
				util.UpdateStatus(j, "errored")
				return
			}
		}
	}

	util.LogDebug(stylish.Bullet("Running before deploy hooks..."))

	// run before deploy hooks
	for _, node := range box.Nodes() {
		bd := box.Node(node).Value("before_deploy")
		bda := box.Node(node).Value("before_deploy_all")
		if bd != nil || bda != nil {

			// run before deploy hook (blocking)
			if _, err := util.ExecHook("before_deploy", node, map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda}); err != nil {
				util.HandleError(stylish.Error("Failed to run before_deploy hook", err.Error()))
				util.UpdateStatus(j, "errored")
				return
			}
		}
	}

	// set routing to web components
	if container, err := util.GetContainer("web1"); err == nil {
		util.LogDebug(stylish.Bullet("Configure routing..."))

		config.Router.AddTarget("/", "http://"+container.NetworkSettings.IPAddress+":8080")
		config.Router.Handler = nil
	} else {
		config.Router.Handler = router.NoDeploy{}
	}

	//
	util.LogDebug(stylish.Bullet("Running after deploy hooks..."))

	// after deploy hooks
	for _, node := range box.Nodes() {
		ad := box.Node(node).Value("after_deploy")
		ada := box.Node(node).Value("after_deploy_all")
		if ad != nil || ada != nil {

			// run after deploy hook (blocking)
			if _, err := util.ExecHook("after_deploy", node, map[string]interface{}{"after_deploy": ad, "after_deploy_all": ada}); err != nil {
				util.HandleError(stylish.Error("Failed to run after_deploy hook", err.Error()))
				util.UpdateStatus(j, "errored")
				return
			}
		}
	}

	util.UpdateStatus(j, "complete")
}
