package tasks

import (
	"github.com/pagodabox/nanobox-server/config"
	// "os/exec"
	"github.com/samalba/dockerclient"
)

func CreateContainer(image string) (*dockerClient.ContainerInfo, err) {
	err := docker.PullImage(image, nil)
	if err != nil {
		return nil, err
	}

	// Create a container
	containerConfig := &dockerclient.ContainerConfig{
		Image:       image,
		Cmd:         []string{"bash"},
		Tty:         true,
	}
	if image == "engine-something" {
		containerConfig.Volumes = map[string]string{"/src":"/src"}
	}
	containerId, err := docker.CreateContainer(containerConfig, "foobar")
	if err != nil {
		return nil, err
	}

	// Start the container
	hostConfig := &dockerclient.HostConfig{}
	err = docker.StartContainer(containerId, hostConfig)
	if err != nil {
		return nil, err
	}

	return docker.InspectContainer(id), nil
}

func RemoveContainer(id string) error {
  err := docker.StopContainer(id, 0)
  if err != nil {
  	return err
  }

  return docker.RemoveContainer(id, false, true)
}

func ListContainers() ([]dockerclient.Container, error) {
	return dockerclient.ListContainers(true, false, "")
}

func dockerClient() *dockerclient.DockerClient {
	d, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		config.Log.Error(err)
	}
	return d
}
