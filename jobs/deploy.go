// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

//
package jobs

//
import (
	"strings"

	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-golang-stylish"
	// "github.com/nanobox-io/nanobox-logtap"
	"github.com/nanobox-io/nanobox-router"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/fs"
	"github.com/nanobox-io/nanobox-server/util/script"
	"github.com/nanobox-io/nanobox-server/util/worker"
)

//
type Deploy struct {
	ID    string
	Reset bool
	Run   bool

	payload map[string]interface{}
}

// Proccess syncronies your docker containers with the boxfile specification
func (j *Deploy) Process() {
	// add a lock so the service wont go down whil im running
	util.Lock()
	defer util.Unlock()

	// set routing to watch logs
	router.ErrorHandler = router.DeployInProgress{}

	// remove all code containers
	util.LogInfo(stylish.Bullet("Cleaning containers"))

	// might as well remove bootstraps and execs too
	containers, _ := docker.ListContainers("code", "build", "bootstrap", "exec", "tcp", "udp")
	for _, container := range containers {
		util.RemoveForward(container.NetworkSettings.IPAddress)
		if err := docker.RemoveContainer(container.ID); err != nil {
			util.HandleError(stylish.Error("Failed to remove old containers", err.Error()))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	// Make sure we have the directories
	if err := fs.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// wipe the previous deploy data if reset == true
	if j.Reset {
		util.LogInfo(stylish.Bullet("Emptying cache"))
		if err := fs.Clean(); err != nil {
			util.HandleError(stylish.Warning("Failed to reset cache and code directories:\n%v", err.Error()))
		}
	}

	// parse the boxfile
	util.LogDebug(stylish.Bullet("Parsing Boxfile"))
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")

	image := "nanobox/build"

	if stab := box.Node("build").StringValue("stability"); stab != "" {
		image = image + ":" + stab
	}

	// if the build image doesn't exist it needs to be downloaded
	if !docker.ImageExists(image) {
		util.LogInfo(stylish.Bullet("Pulling the latest build image (this may take awhile)... "))
		docker.InstallImage(image)
	}

	util.LogDebug(stylish.Bullet("image name: %v", image))

	// create a build container
	util.LogInfo(stylish.Bullet("Creating build container"))

	_, err := docker.CreateContainer(docker.CreateConfig{Image: image, Category: "build", UID: "build1"})
	if err != nil {
		util.HandleError(stylish.Error("Failed to create build container", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// define the deploy payload
	j.payload = map[string]interface{}{
		"platform":    "local",
		"app":         config.App,
		"dns":         []string{config.App + ".dev"},
		"port":        "8080",
		"boxfile":     box.Node("build").Parsed,
		"logtap_host": config.LogtapHost,
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

	// run the default-user hook to get ssh keys setup
	if out, err := script.Exec("default-user", "build1", fs.UserPayload()); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run user script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run configure hook (blocking)
	if out, err := script.Exec("default-configure", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run configure script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run detect script (blocking)
	if out, err := script.Exec("default-detect", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run detect script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run sync script (blocking)
	if out, err := script.Exec("default-sync", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run sync script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run setup script (blocking)
	if out, err := script.Exec("default-setup", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run setup script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// run boxfile script (blocking)
	if !box.Node("build").BoolValue("disable_engine_boxfile") {
		if out, err := script.Exec("default-boxfile", "build1", j.payload); err != nil {
			util.LogDebug("Failed script output: \n %s", out)
			util.HandleError(stylish.Error("Failed to run boxfile script", err.Error()))
			util.UpdateStatus(j, "errored")
			return

			// if the script runs succesfully merge the boxfiles
		} else {
			util.LogDebug(stylish.Bullet("Merging Boxfiles..."))
			box.Merge(boxfile.New([]byte(out)))
		}
	}

	// add the missing storage nodes to the boxfile
	box.AddStorageNode()
	j.payload["boxfile"] = box.Node("build").Parsed

	// remove any containers no longer in the boxfile
	util.LogDebug(stylish.Bullet("Removing old containers..."))
	serviceContainers, _ := docker.ListContainers("service")
	for _, container := range serviceContainers {
		if !box.Node(container.Config.Labels["uid"]).Valid {
			util.RemoveForward(container.NetworkSettings.IPAddress)
			docker.RemoveContainer(container.ID)
		}
	}

	worker := worker.New()
	worker.Blocking = true
	worker.Concurrent = true

	//
	serviceStarts := []*ServiceStart{}

	// build service containers according to boxfile
	for _, node := range box.Nodes("service") {
		if _, err := docker.GetContainer(node); err != nil {
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

	if worker.Count() > 0 {
		util.LogInfo(stylish.Bullet("Launching data services"))
	}

	worker.Process()

	// ensure all services started correctly before continuing
	for _, starts := range serviceStarts {
		if !starts.Success {
			util.HandleError(stylish.ErrorHead("Failed to start %v", starts.UID))
			util.HandleError(stylish.ErrorBody(""))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	// grab the environment data from all service containers
	evars := j.payload["env"].(map[string]string)

	// clear out the old ports from the previous deploy
	clearPorts()

	//
	serviceEnvs := []*ServiceEnv{}

	serviceContainers, _ = docker.ListContainers("service")
	for _, container := range serviceContainers {

		s := ServiceEnv{UID: container.Config.Labels["uid"]}
		serviceEnvs = append(serviceEnvs, &s)

		worker.Queue(&s)
	}

	worker.Process()

	for _, env := range serviceEnvs {
		if !env.Success {
			util.HandleError(stylish.ErrorHead("Failed to configure %v's environment variables", env.UID))
			util.HandleError(stylish.ErrorBody(""))
			util.UpdateStatus(j, "errored")
			return
		}

		for key, val := range env.EVars {
			evars[strings.ToUpper(env.UID+"_"+key)] = val
		}
	}

	j.payload["env"] = evars

	// run prepare script (blocking)
	if out, err := script.Exec("default-prepare", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run prepare script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	if j.Run {
		// run build script (blocking)
		if out, err := script.Exec("default-build", "build1", j.payload); err != nil {
			util.LogDebug("Failed script output: \n %s", out)
			util.HandleError(stylish.Error("Failed to run build script", err.Error()))
			util.UpdateStatus(j, "errored")
			return
		}

		// run publish script (blocking)
		if out, err := script.Exec("default-publish", "build1", j.payload); err != nil {
			util.LogDebug("Failed script output: \n %s", out)
			util.HandleError(stylish.Error("Failed to run publish script", err.Error()))
			util.UpdateStatus(j, "errored")
			return
		}
	}

	// run cleanup script (blocking)
	if out, err := script.Exec("default-cleanup", "build1", j.payload); err != nil {
		util.LogDebug("Failed script output: \n %s", out)
		util.HandleError(stylish.Error("Failed to run cleanup script", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// we will only create new code nodes if we are
	// supposed to be running
	if j.Run {

		// build new code containers
		codeServices := []*ServiceStart{}
		for _, node := range box.Nodes("code") {
			if _, err := docker.GetContainer(node); err != nil {
				// container doesn't exist so we need to create it
				s := ServiceStart{
					Boxfile: box.Node(node),
					UID:     node,
					EVars:   evars,
				}

				codeServices = append(codeServices, &s)

				worker.Queue(&s)
			}
			if worker.Count() > 0 {
				util.LogInfo(stylish.Bullet("Launching Code services"))
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

	util.LogDebug(stylish.Bullet("Running before deploy scripts..."))

	// run before deploy scripts
	for _, node := range box.Nodes() {
		bd := box.Node(node).Value("before_deploy")
		bda := box.Node(node).Value("before_deploy_all")
		if bd != nil || bda != nil {

			// run before deploy script (blocking)
			if out, err := script.Exec("default-before_deploy", node, map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda}); err != nil {
				util.LogDebug("Failed script output: \n %s", out)
				util.HandleError(stylish.Error("Failed to run before_deploy script", err.Error()))
				util.UpdateStatus(j, "errored")
				return
			}
		}
	}

	// configure the port forwards per service
	err = configurePorts(box)
	if err != nil {
		util.HandleError(stylish.Error("Failed to configure Ports", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// configure the routing mesh for any web services
	err = configureRoutes(box)
	if err != nil {
		util.HandleError(stylish.Error("Failed to configure Routes", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	//
	util.LogDebug(stylish.Bullet("Running after deploy hooks..."))

	// after deploy hooks
	for _, node := range box.Nodes() {
		ad := box.Node(node).Value("after_deploy")
		ada := box.Node(node).Value("after_deploy_all")
		if ad != nil || ada != nil {

			// run after deploy hook (blocking)
			if out, err := script.Exec("default-after_deploy", node, map[string]interface{}{"after_deploy": ad, "after_deploy_all": ada}); err != nil {
				util.LogDebug("Failed script output: \n %s", out)
				util.HandleError(stylish.Error("Failed to run after_deploy script", err.Error()))
				util.UpdateStatus(j, "errored")
				return
			}
		}
	}

	util.UpdateStatus(j, "complete")
}
