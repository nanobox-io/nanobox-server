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
	"reflect"

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
	if err := j.RemoveOldContainers(); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	if err := j.SetupFS(); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	// parse the boxfile
	util.LogDebug(stylish.Bullet("Parsing Boxfile"))
	box := UserBoxfile(true)

	if err := j.CreateBuildContainer(box.Node("build")); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	// define the deploy payload
	j.payload = map[string]interface{}{
		"platform":    "local",
		"app":         config.App(),
		"dns":         []string{config.App() + ".dev"},
		"port":        "8080",
		"boxfile":     box.Node("build").Parsed,
		"logtap_host": config.LogtapHost,
	}

	j.payload["env"] = DefaultEVars(*box)

	if err := j.SetupBuild(); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	// get a cached copy of the combined boxfile
	oldCombinedBox := CombinedBoxfile(false)

	// make sure to grab the new engine boxfile
	EngineBoxfile(true)
	// grab a new boxfile
	box = CombinedBoxfile(true)
	// add the missing storage nodes to the boxfile
	box.AddStorageNode()

	j.payload["boxfile"] = box.Node("build").Parsed

	// remove any containers no longer in the boxfile
	// this will also remove any services where the boxfile has been modified since last deploy
	util.LogDebug(stylish.Bullet("Removing old containers..."))
	serviceContainers, _ := docker.ListContainers("service")
	for _, container := range serviceContainers {
		if !box.Node(container.Config.Labels["uid"]).Valid {
			util.LogDebug(stylish.SubBullet("- removing " +container.Config.Labels["uid"] ))
			util.RemoveForward(container.NetworkSettings.IPAddress)
			docker.RemoveContainer(container.ID)
		}
		if !reflect.DeepEqual(box.Node(container.Config.Labels["uid"]), oldCombinedBox.Node(container.Config.Labels["uid"])) {
			util.LogDebug(stylish.SubBullet("- replacing " +container.Config.Labels["uid"] ))
			util.RemoveForward(container.NetworkSettings.IPAddress)
			docker.RemoveContainer(container.ID)
		}
	}

	worker := worker.New()
	worker.Blocking = true

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

	// we dont want service starts to be concurrent here for messaging
	worker.Concurrent = false
	worker.Process()
	// make the worker concurrent from here on
	worker.Concurrent = true

	failedStart := false
	// ensure all services started correctly before continuing
	for _, starts := range serviceStarts {
		if !starts.Success {
			util.HandleError(stylish.ErrorHead("Failed to start %v", starts.UID))
			util.HandleError(stylish.ErrorBody(""))
			failedStart = true
		}
	}
	if failedStart {
		util.UpdateStatus(j, "errored")
		return
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
		for _, serviceStart := range serviceStarts {
			if serviceStart.UID == s.UID {
				s.FirstTime = true
			}
		}
		serviceEnvs = append(serviceEnvs, &s)

		worker.Queue(&s)
	}

	worker.Process()

	failedEnv := false
	for _, env := range serviceEnvs {
		if !env.Success {
			util.HandleError(stylish.ErrorHead("Failed to configure %v's environment variables", env.UID))
			util.HandleError(stylish.ErrorBody(""))
			failedEnv = true
			continue
		}

		for key, val := range env.EVars {
			evars[strings.ToUpper(env.UID+"_"+key)] = val
		}
	}
	if failedEnv {
		util.UpdateStatus(j, "errored")

		return
	}

	j.payload["env"] = evars

	if err := j.RunBuild(); err != nil {
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

	if err := j.RunDeployScripts("before", *box); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	// configure the port forwards per service
	if err := configurePorts(*box); err != nil {
		util.HandleError(stylish.Error("Failed to configure Ports", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	// configure the routing mesh for any web services
	if err := configureRoutes(*box); err != nil {
		util.HandleError(stylish.Error("Failed to configure Routes", err.Error()))
		util.UpdateStatus(j, "errored")
		return
	}

	//
	util.LogDebug(stylish.Bullet("Running after deploy hooks..."))

	if err := j.RunDeployScripts("after", *box); err != nil {
		util.UpdateStatus(j, "errored")
		return
	}

	util.UpdateStatus(j, "complete")
}

func (j *Deploy) RemoveOldContainers() error {
	// might as well remove bootstraps and execs too
	containers, _ := docker.ListContainers("code", "build", "bootstrap", "dev", "tcp", "udp")
	for _, container := range containers {
		util.RemoveForward(container.NetworkSettings.IPAddress)
		if err := docker.RemoveContainer(container.ID); err != nil {
			util.HandleError(stylish.Error("Failed to remove old containers", err.Error()))
			return err
		}
	}
	return nil
}

func (j *Deploy) SetupFS() error {
	// Make sure we have the directories
	if err := fs.CreateDirs(); err != nil {
		util.HandleError(stylish.Error("Failed to create dirs", err.Error()))
		return err
	}

	// wipe the previous deploy data if reset == true
	if j.Reset {
		util.LogInfo(stylish.Bullet("Emptying cache"))
		if err := fs.Clean(); err != nil {
			util.HandleError(stylish.Warning("Failed to reset cache and code directories:\n%v", err.Error()))
			return err
		}
	}
	return nil
}

func (j *Deploy) CreateBuildContainer(box boxfile.Boxfile) error {
	image := "nanobox/build"

	if stab := box.StringValue("stability"); stab != "" {
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
		return err
	}
	return nil
}

func (j *Deploy) SetupBuild() error {
	// run the default-user hook to get ssh keys setup
	if _, err := script.Exec("default-user", "build1", fs.UserPayload()); err != nil {
		return err
	}

	if _, err := script.Exec("default-configure", "build1", j.payload); err != nil {
		return err
	}

	if _, err := script.Exec("default-detect", "build1", j.payload); err != nil {
		return err
	}

	if _, err := script.Exec("default-sync", "build1", j.payload); err != nil {
		return err
	}

	if _, err := script.Exec("default-setup", "build1", j.payload); err != nil {
		return err
	}
	return nil
}

func (j *Deploy) RunBuild() error {
	// run prepare script (blocking)
	if _, err := script.Exec("default-prepare", "build1", j.payload); err != nil {
		return err
	}

	if j.Run {
		// run build script (blocking)
		if _, err := script.Exec("default-build", "build1", j.payload); err != nil {
			return err
		}

		// run publish script (blocking)
		if _, err := script.Exec("default-publish", "build1", j.payload); err != nil {
			return err
		}
	}

	// run cleanup script (blocking)
	if _, err := script.Exec("default-cleanup", "build1", j.payload); err != nil {
		return err
	}
	return nil
}

func (j *Deploy) RunDeployScripts(stage string, box boxfile.Boxfile) error {
	// run before deploy scripts
	for _, node := range box.Nodes() {
		bd := box.Node(node).Value(stage + "_deploy")
		bda := box.Node(node).Value(stage + "_deploy_all")
		if bd != nil || bda != nil {

			// run before deploy script (blocking)
			if _, err := script.Exec(fmt.Sprintf("default-%s_deploy", stage), node, map[string]interface{}{stage + "_deploy": bd, stage + "_deploy_all": bda}); err != nil {
				return err
			}
		}
	}
	return nil
}
