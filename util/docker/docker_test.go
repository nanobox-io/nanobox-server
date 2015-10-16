package docker_test

import (
	"testing"
	"github.com/golang/mock/gomock"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/docker/mock_docker"
)

type createMatcher struct {
}

func (c createMatcher) Matches(x interface{}) bool {
	createConfig, ok := x.(dc.CreateContainerOptions)
	if !ok {
		return false
	}
	binds :=[]string{
		"/mnt/sda/var/nanobox/cache/:/mnt/cache/",
		"/mnt/sda/var/nanobox/deploy/:/mnt/deploy/",
		"/mnt/sda/var/nanobox/build/:/mnt/build/",
		"/vagrant/code//:/share/code/:ro", // the app name cannot be grabbed outside the environment
		"/vagrant/engines/:/share/engines/:ro",
	}
	for i, bind := range createConfig.HostConfig.Binds {
		if binds[i] != bind {
			return false
		}
	}
	return true
}

func (c createMatcher) String() string {
	return "is a CreateContainerOptions"
}

func TestCreatContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mClient := mock_docker.NewMockClientInterface(ctrl)
	docker.Client = mClient

	gomock.InOrder(
    mClient.EXPECT().ListImages(dc.ListImagesOptions{}).Return([]dc.APIImages{dc.APIImages{RepoTags: []string{"nanobox/build:latest"}}}, nil),
    mClient.EXPECT().CreateContainer(createMatcher{}).Return(&dc.Container{ID: "1234"}, nil),
    mClient.EXPECT().StartContainer("1234", nil),
    mClient.EXPECT().InspectContainer("1234"),
	)
	
	cc := docker.CreateConfig{
		Category: "build",
		UID:      "build1",
		Name:     "build",
		Cmd:      []string{"ls"},
		Image:    "nanobox/build",
	}
	docker.CreateContainer(cc)

}


func TestExecInContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mClient := mock_docker.NewMockClientInterface(ctrl)
	docker.Client = mClient

	 opts := dc.CreateExecOptions{
	      AttachStdout: true,
	      AttachStderr: true,
	      Cmd:          []string{"ls", "-la"},
	      Container:    "exec1",
	      User:         "root",
	  }
	gomock.InOrder(
    mClient.EXPECT().CreateExec(opts).Return(&dc.Exec{ID:"1234"}, nil),
    mClient.EXPECT().StartExec("1234", gomock.Any()),
    mClient.EXPECT().InspectExec("1234").Return(&dc.ExecInspect{ExitCode:0}, nil),
  )
	docker.ExecInContainer("exec1", "ls", "-la")
	
}