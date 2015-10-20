// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package docker

//
import (
	"fmt"
	"strings"

	docksig "github.com/docker/docker/pkg/signal"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/nanobox-io/nanobox-server/config"
)

type CreateConfig struct {
	Category string
	UID      string
	Name     string
	Cmd      []string
	Image    string
}

func (d DockerUtil) CreateContainer(conf CreateConfig) (*dc.Container, error) {
	if conf.Category == "" || conf.Image == "" {
		return nil, fmt.Errorf("Cannot create a container without an image")
	}
	cConfig := dc.CreateContainerOptions{
		Name: conf.UID,
		Config: &dc.Config{
			Tty:             true,
			Labels:          map[string]string{conf.Category: "true", "uid": conf.UID, "name": conf.Name},
			NetworkDisabled: false,
			Image:           conf.Image,
			Cmd:             conf.Cmd,
		},
		HostConfig: &dc.HostConfig{
			Privileged:    true,
			RestartPolicy: dc.AlwaysRestart(),
		},
	}
	addCategoryConfig(conf.Category, &cConfig)
	return createContainer(cConfig)
}

func addCategoryConfig(category string, cConfig *dc.CreateContainerOptions) {
	switch category {
	case "dev":
		cConfig.Config.Hostname = fmt.Sprintf("%s.dev", config.App)
		cConfig.Config.OpenStdin = true
		cConfig.Config.AttachStdin = true
		cConfig.Config.AttachStdout = true
		cConfig.Config.AttachStderr = true
		cConfig.Config.WorkingDir = "/code"
		cConfig.Config.User = "gonano"
		cConfig.HostConfig.Binds = append([]string{
			"/vagrant/code/" + config.App + "/:/code/",
		}, libDirs()...)
		if container, err := GetContainer("build1"); err == nil {
			cConfig.HostConfig.Binds = append(cConfig.HostConfig.Binds, fmt.Sprintf("/mnt/sda/var/lib/docker/aufs/mnt/%s/data/:/data/", container.ID))
		}

		cConfig.HostConfig.NetworkMode = "host"
	case "build":
		cConfig.Config.Cmd = []string{"/bin/sleep", "365d"}
		cConfig.HostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
			"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",
			"/mnt/sda/var/nanobox/build/:/mnt/build/",

			"/vagrant/code/" + config.App + "/:/share/code/:ro",
			"/vagrant/engines/:/share/engines/:ro",
		}
	case "bootstrap":
		cConfig.Config.Cmd = []string{"/bin/sleep", "365d"}
		cConfig.HostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
			"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",

			"/vagrant/code/" + config.App + "/:/code/",
			"/vagrant/engines/:/share/engines/:ro",
		}
	case "code":
		cConfig.HostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/deploy/:/data/",
			"/mnt/sda/var/nanobox/build/:/code/",
		}
	case "service":
		if strings.Contains(cConfig.Name, "/") {
			cConfig.Name = strings.Replace(cConfig.Name, "/", "-", -1)
		}
		// nothing to be done here
	}
	return
}

// createContainer
func createContainer(cConfig dc.CreateContainerOptions) (*dc.Container, error) {

	// LogInfo("CREATE CONTAINER! %#v", cConfig)

	//
	if !ImageExists(cConfig.Config.Image) {
		if err := InstallImage(cConfig.Config.Image); err != nil {
			return nil, err
		}
	}

	// create container
	container, err := Client.CreateContainer(cConfig)
	if err != nil {
		return nil, err
	}

	if err := StartContainer(container.ID); err != nil {
		return nil, err
	}

	return InspectContainer(container.ID)
}

// Start
func (d DockerUtil) StartContainer(id string) error {
	return Client.StartContainer(id, nil)
}

func (d DockerUtil) KillContainer(id, sig string) error {
	return Client.KillContainer(dc.KillContainerOptions{ID: id, Signal: dc.Signal(docksig.SignalMap[sig])})
}

func (d DockerUtil) ResizeContainerTTY(id string, height, width int) error {
	return Client.ResizeContainerTTY(id, height, width)
}

func (d DockerUtil) WaitContainer(id string) (int, error) {
	return Client.WaitContainer(id)
}

// RemoveContainer
func (d DockerUtil) RemoveContainer(id string) error {
	Client.StopContainer(id, 0)
	// if it errors on stopping ignore it
	return Client.RemoveContainer(dc.RemoveContainerOptions{ID: id, RemoveVolumes: false, Force: true})
}

// InspectContainer
func (d DockerUtil) InspectContainer(id string) (*dc.Container, error) {
	return Client.InspectContainer(id)
}

// GetContainer
func (d DockerUtil) GetContainer(id string) (*dc.Container, error) {
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.Name == id || container.Name == ("/"+id) || container.ID == id {
			return container, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

// ListContainers
func (d DockerUtil) ListContainers(labels ...string) ([]*dc.Container, error) {
	rtn := []*dc.Container{}

	apiContainers, err := Client.ListContainers(dc.ListContainersOptions{All: true, Size: false})
	if len(labels) == 0 || err != nil {
		for _, apiContainer := range apiContainers {
			container, _ := InspectContainer(apiContainer.ID)
			if container != nil {
				rtn = append(rtn, container)
			}
		}
		return rtn, err
	}

	for _, apiContainer := range apiContainers {
		for _, label := range labels {
			if apiContainer.Labels[label] == "true" {
				container, _ := InspectContainer(apiContainer.ID)
				if container != nil {
					rtn = append(rtn, container)
				}
			}
		}
	}

	return rtn, nil
}
