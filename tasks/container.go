package tasks

import (
	// "os"
	// "os/exec"
	"github.com/samalba/dockerclient"
)

func ListContainers() []dockerclient.Container, error {
	dockerclient,ListContainers(true, false, "") (, error)
}


func dockerClient() *dockerclient.DockerClient, err{
	 dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
}