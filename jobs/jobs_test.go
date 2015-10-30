package jobs_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nanobox-io/nanobox-boxfile"

	"github.com/nanobox-io/nanobox-server/jobs"
	"github.com/nanobox-io/nanobox-server/util/docker/mock_docker"
	"github.com/nanobox-io/nanobox-server/util/docker"
	dc "github.com/fsouza/go-dockerclient"

	"github.com/nanobox-io/nanobox-server/util/fs/mock_fs"
	"github.com/nanobox-io/nanobox-server/util/fs"

	"github.com/nanobox-io/nanobox-server/util/script"
)

func TestDeployRemoveOldContainers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mDocker := mock_docker.NewMockDockerDefault(ctrl)
	docker.Default = mDocker

	mDocker.EXPECT().ListContainers("code", "build", "bootstrap", "dev", "tcp", "udp").Return([]*dc.Container{&dc.Container{ID:"1234", NetworkSettings: &dc.NetworkSettings{IPAddress: "1.2.3.4"}}}, nil)
	mDocker.EXPECT().RemoveContainer("1234")

	deploy := jobs.Deploy{}
	deploy.RemoveOldContainers()

}

func TestDeploySetupFs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mFs := mock_fs.NewMockFsUtil(ctrl)
	fs.FsDefault = mFs
	
	mFs.EXPECT().CreateDirs()
	mFs.EXPECT().Clean()

	deploy := jobs.Deploy{Reset: true}
	deploy.SetupFS()

}

func TestCreateBuildContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mDocker := mock_docker.NewMockDockerDefault(ctrl)
	docker.Default = mDocker

	gomock.InOrder(
		mDocker.EXPECT().ImageExists("nanobox/build").Return(false),
		mDocker.EXPECT().InstallImage("nanobox/build"),
		mDocker.EXPECT().CreateContainer(docker.CreateConfig{Image: "nanobox/build", Category: "build", UID: "build1"}),
	)
	
	deploy := jobs.Deploy{}
	deploy.CreateBuildContainer(boxfile.Boxfile{})

}

func TestSetupBuild(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mFs := mock_fs.NewMockFsUtil(ctrl)
	fs.FsDefault = mFs
	
	mFs.EXPECT().UserPayload()	

	names := []string{}
	script.Exec = func(name, container string, payload map[string]interface{}) ([]byte, error)  {
		names = append(names, name)
		return []byte{}, nil
	}
	deploy := jobs.Deploy{}
	deploy.SetupBuild()
	expectedNames := []string{
		"default-user", 
		"default-configure", 
		"default-detect", 
		"default-sync", 
		"default-setup",
	}
	if len(names) != len(expectedNames) {
		t.Errorf("calls dont match the expected list of calls (%+v)", names)
		return
	}
	for i, name := range expectedNames {
		if names[i] != name {
			t.Errorf("I was expecting %s but got %s", name, names[i])
		}
	}
}

func TestRunBuild(t *testing.T) {
	names := []string{}
	script.Exec = func(name, container string, payload map[string]interface{}) ([]byte, error)  {
		names = append(names, name)
		return []byte{}, nil
	}
	deploy := jobs.Deploy{Run: true}
	deploy.RunBuild()
	expectedNames := []string{
		"default-prepare",
		"default-build",
		"default-publish",
		"default-cleanup",
	}
	if len(names) != len(expectedNames) {
		t.Errorf("calls dont match the expected list of calls (%+v)", names)
		return
	}
	for i, name := range expectedNames {
		if names[i] != name {
			t.Errorf("I was expecting %s but got %s", name, names[i])
		}
	}
}

func TestRunDeployScripts(t *testing.T) {
	names := []string{}
	script.Exec = func(name, container string, payload map[string]interface{}) ([]byte, error)  {
		names = append(names, name)
		return []byte{}, nil
	}
	deploy := jobs.Deploy{Run: true}
	deploy.RunDeployScripts("before", boxfile.New([]byte(`---
web1:
  before_deploy:
    - "php artisan migrate"
  before_deploy_all:
    - "php scripts/clear_cache.php"
`)))
	expectedNames := []string{
		"default-before_deploy",
	}
	if len(names) != len(expectedNames) {
		t.Errorf("calls dont match the expected list of calls (%+v)", names)
		return
	}
	for i, name := range expectedNames {
		if names[i] != name {
			t.Errorf("I was expecting %s but got %s", name, names[i])
		}
	}
}

