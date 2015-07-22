// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package util

//
import (
	"errors"
	"fmt"
	"os/exec"
	// "strconv"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/samalba/dockerclient"
)

// CreateBuildContainer
func CreateBuildContainer(name string) (*dockerclient.ContainerInfo, error) {
	cConfig := &dockerclient.ContainerConfig{
		Tty:             true,
		Labels:          map[string]string{"build": "true", "uid": name},
		NetworkDisabled: false,
		Image:           "nanobox/build",
		Cmd:             []string{"/bin/sleep", "365d"},
		HostConfig: dockerclient.HostConfig{
			Binds: []string{
				"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
				"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",
				"/mnt/sda/var/nanobox/build/:/mnt/build/",

				"/vagrant/code/" + config.App + "/:/share/code/:ro",
				"/vagrant/engines/:/share/engines/:ro",
			},
			Privileged: true,
		},
	}

	return createContainer(cConfig)
}

// CreateCodeContainer
func CreateCodeContainer(name string) (*dockerclient.ContainerInfo, error) {
	cConfig := &dockerclient.ContainerConfig{
		Tty:             true,
		Labels:          map[string]string{"code": "true", "uid": name},
		NetworkDisabled: false,
		Image:           "nanobox/code",
		HostConfig: dockerclient.HostConfig{
			Binds: []string{
				"/mnt/sda/var/nanobox/deploy/:/data/",
				"/mnt/sda/var/nanobox/build/:/code/:ro",
			},
			Privileged: true,
		},
	}

	return createContainer(cConfig)
}

// CreateServiceContainer
func CreateServiceContainer(name, image string) (*dockerclient.ContainerInfo, error) {
	cConfig := &dockerclient.ContainerConfig{
		Tty:             true,
		Labels:          map[string]string{"service": "true", "uid": name},
		NetworkDisabled: false,
		Image:           image,
		HostConfig: dockerclient.HostConfig{
			Binds:      []string{},
			Privileged: true,
		},
	}

	return createContainer(cConfig)
}

// createContainer
func createContainer(cConfig *dockerclient.ContainerConfig) (*dockerclient.ContainerInfo, error) {

	// LogInfo("CREATE CONTAINER! %#v", cConfig)

	//
	if !ImageExists(cConfig.Image) {
		if err := dockerClient().PullImage(cConfig.Image, nil); err != nil {
			return nil, err
		}
	}

	// create container
	id, err := dockerClient().CreateContainer(cConfig, cConfig.Labels["uid"])
	if err != nil {
		LogInfo("NO WORKEY!! %v", err)
		return nil, err
	}

	if err := StartContainer(id); err != nil {
		return nil, err
	}

	return dockerClient().InspectContainer(id)
}

// Start
func StartContainer(id string) error {
	return dockerClient().StartContainer(id, &dockerclient.HostConfig{
		Privileged: true,
	})
}

// RemoveContainer
func RemoveContainer(id string) error {
	if _, err := dockerClient().InspectContainer(id); err != nil {
		return err
	}

	if err := dockerClient().StopContainer(id, 0); err != nil {
		return err
	}

	return dockerClient().RemoveContainer(id, false, true)
}

// InspectContainer
func InspectContainer(id string) (*dockerclient.ContainerInfo, error) {
	return dockerClient().InspectContainer(id)
}

// GetContainer
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

// ListContainers
func ListContainers(labels ...string) ([]dockerclient.Container, error) {
	containers, err := dockerClient().ListContainers(true, false, "")
	if len(labels) == 0 || err != nil {
		return containers, err
	}

	rtn := []dockerclient.Container{}

	for _, container := range containers {
		for _, label := range labels {
			if container.Labels[label] == "true" {
				rtn = append(rtn, container)
			}
		}
	}

	return rtn, nil
}

// Exec
func ExecInContainer(container string, args ...string) ([]byte, error) {

	// build the initial command, and then iterate over any additional arguments
	// that are passed in as commands adding them the the final command
	cmd := []string{"exec", container}
	for _, a := range args {
		cmd = append(cmd, a)
	}

	return exec.Command("docker", cmd...).Output()
}

// Run
func RunInContainer(container, img string, args ...string) ([]byte, error) {

	// build the initial command, and then iterate over any additional arguments
	// that are passed in as commands adding them the the final command
	cmd := []string{"run", "--rm", container, img}
	for _, a := range args {
		cmd = append(cmd, a)
	}

	return exec.Command("docker", cmd...).Output()
}

// ImageExists
func ImageExists(name string) bool {
	images, err := dockerClient().ListImages(true)
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

func InstallImage(image string) error {
	if err := dockerClient().PullImage(image, nil); err != nil {
		return err
	}

	return nil
}

// PullImage
func UpdateAllImages() error {
	images, err := dockerClient().ListImages(true)
	if err != nil {
		return err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			err := UpdateImage(tag)
			if err != nil {
				return err
			}
		}
	}
	config.Mist.Publish([]string{"update"}, `{"model":"Update", "action":"update", "document":"{\"id\":\"1\", \"status\":\"complete\"}"}`)
	return nil
}

func UpdateImage(image string) error {
	config.Mist.Publish([]string{"update"}, fmt.Sprintf(`{"model":"Update", "action":"update", "document":"{\"id\":\"1\", \"status\":\"pulling image for %s\"}"}`, image))
	if err := dockerClient().PullImage(image, nil); err != nil {
		return err
	}

	return nil
}

// dockerClient
func dockerClient() *dockerclient.DockerClient {
	d, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		config.Log.Error(err.Error())
	}
	return d
}
