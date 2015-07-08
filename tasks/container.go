// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package tasks

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/samalba/dockerclient"
)

func PullImages() error {
	images, err := dockerClient().ListImages()
	if err != nil {
		return err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			config.Mist.Publish([]string{"update"}, fmt.Sprintf(`{"model":"Update", "action":"update", "document":"{\"id\":\"1\", \"status\":\"pulling image for %s\"}"}`, tag))
			err := dockerClient().PullImage(tag, nil)
			if err != nil {
				return err
			}
		}
	}
	config.Mist.Publish([]string{"update"}, `{"model":"Update", "action":"update", "document":"{\"id\":\"1\", \"status\":\"complete\"}"}`)
	return nil
}

func CreateContainer(image string, labels map[string]string) (*dockerclient.ContainerInfo, error) {
	config.Log.Debug("[Task::Container] Create %s, %#v\n", image, labels)
	if !ImageExists(image) {
		err := dockerClient().PullImage(image, nil)
		if err != nil {
			return nil, err
		}
	}

	containerConfig := &dockerclient.ContainerConfig{
		// Cmd: []string{"bash"},
		Tty:             true,
		Labels:          labels,
		NetworkDisabled: false,
		Image:           image,
	}

	val := labels["uid"]
	if val == "" {
		containers, _ := ListContainers()
		val = "unknown" + strconv.Itoa(len(containers)+1)
	}

	containerId, err := dockerClient().CreateContainer(containerConfig, val)
	if err != nil {
		config.Log.Error("[Task::Container] create %s, \n", err.Error())
		return nil, err
	}

	config.Log.Debug("[Task::Container] containerid %#v\n", containerId)
	// Start the container
	hostConfig := &dockerclient.HostConfig{
		Privileged: true,
	}
	if labels["build"] == "true" {
		hostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
			"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",

			"/vagrant/code/" + config.App + "/:/share/code/:ro",
			"/vagrant/engines/:/share/engines/:ro",
		}
	}

	if labels["code"] == "true" {
		hostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/deploy/:/data/",
		}
	}
	err = dockerClient().StartContainer(containerId, hostConfig)
	if err != nil {
		config.Log.Error("[Task::Container] start %s, \n", err.Error())
		return nil, err
	}
	container, err := dockerClient().InspectContainer(containerId)
	config.Router.AddForward(container.NetworkSettings.IPAddress + ":22")
	return container, err
}

func StartContainer(id string) error {
	c, err := GetContainer(id)
	if err != nil {
		return err
	}
	hostConfig := &dockerclient.HostConfig{
		Privileged: true,
	}
	if c.Labels["build"] == "true" {
		hostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
			"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",

			"/vagrant/code/" + config.App + "/:/share/code/:ro",
			"/vagrant/engines/:/share/engines/:ro",
		}
	}

	if c.Labels["code"] == "true" {
		hostConfig.Binds = []string{
			"/mnt/sda/var/nanobox/deploy/:/data/:ro",
		}
	}
	return dockerClient().StartContainer(id, hostConfig)
}

func RemoveContainer(id string) error {
	container, err := dockerClient().InspectContainer(id)
	if err != nil {
		return err
	}
	config.Router.RemoveForward(container.NetworkSettings.IPAddress + ":22")

	err = dockerClient().StopContainer(id, 0)
	if err != nil {
		// return err
	}

	return dockerClient().RemoveContainer(id, false, true)
}

func GetDetailedContainer(id string) (*dockerclient.ContainerInfo, error) {
	return dockerClient().InspectContainer(id)
}

func GetContainer(name string) (dockerclient.Container, error) {
	containers, err := ListContainers()
	if err != nil {
		return dockerclient.Container{}, err
	}

	for _, container := range containers {
		for _, n := range container.Names {
			if n == name || n == ("/"+name) {
				return container, nil
			}
		}
		if container.Id == name {
			return container, nil
		}
	}
	return dockerclient.Container{}, errors.New("not found")
}

func ListContainers(labels ...string) ([]dockerclient.Container, error) {
	containers, err := dockerClient().ListContainers(true, false, "")
	if len(labels) == 0 || err != nil {
		return containers, err
	}

	rtn := []dockerclient.Container{}

	for _, label := range labels {
		for _, container := range containers {
			if container.Labels[label] == "true" {
				rtn = append(rtn, container)
			}
		}
	}
	return rtn, nil

}

func dockerClient() *dockerclient.DockerClient {
	d, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		config.Log.Error(err.Error())
	}
	return d
}

func ImageExists(name string) bool {
	images, err := dockerClient().ListImages()
	if err != nil {
		return false
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == name+":latest" {
				return true
			}
		}
	}
	return false
}
