package docker_test

import (
	"fmt"
	"os"
	"testing"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/nanobox-io/nanobox-server/util/docker"
)

type TestClient struct {
	CallLog []string
}

func (t *TestClient) ListImages(opts dc.ListImagesOptions) ([]dc.APIImages, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ListImages(%#v)", opts))
	return []dc.APIImages{}, nil
}
func (t *TestClient) PullImage(opts dc.PullImageOptions, auth dc.AuthConfiguration) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("PullImage(%#v, %#v)", opts, auth))
	return nil
}
func (t *TestClient) CreateContainer(opts dc.CreateContainerOptions) (*dc.Container, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("CreateContainer(%#v)", opts))
	return &dc.Container{ID:"1234"}, nil
}
func (t *TestClient) StartContainer(id string, hostConfig *dc.HostConfig) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StartContainer(%s, %#v)", id, hostConfig))
	return nil
}
func (t *TestClient) KillContainer(opts dc.KillContainerOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("KillContainer(%#v)", opts))
	return nil
}
func (t *TestClient) ResizeContainerTTY(id string, height, width int) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ResizeContainerTTY(%s, %d, %d)", id, height, width))
	return nil
}
func (t *TestClient) StopContainer(id string, timeout uint) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StopContainer(%s, %d)", id, timeout))
	return nil
}
func (t *TestClient) RemoveContainer(opts dc.RemoveContainerOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("RemoveContainer(%#v)", opts))
	return nil
}
func (t *TestClient) WaitContainer(id string) (int, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("WaitContainer(%s)", id))
	return 0, nil
}
func (t *TestClient) InspectContainer(id string) (*dc.Container, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("InspectContainer(%s)", id))
	return &dc.Container{}, nil
}
func (t *TestClient) ListContainers(opts dc.ListContainersOptions) ([]dc.APIContainers, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ListContainers(%#v)", opts))
	return []dc.APIContainers{}, nil
}
func (t *TestClient) CreateExec(opts dc.CreateExecOptions) (*dc.Exec, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("CreateExec(%#v)", opts))
	return &dc.Exec{"1234"}, nil
}
func (t *TestClient) ResizeExecTTY(id string, height, width int) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ResizeExecTTY(%s, %d, %d)", id, height, width))
	return nil
}
func (t *TestClient) StartExec(id string, opts dc.StartExecOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StartExec(%s, %#v)", id, opts))
	return nil
}
func (t *TestClient) InspectExec(id string) (*dc.ExecInspect, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("InspectExec(%s)", id))
	return &dc.ExecInspect{}, nil
}



func TestMain(m *testing.M) {
	docker.Client = &TestClient{[]string{}}
	os.Exit(m.Run())
}

func TestCreatContainer(t *testing.T) {
	cc := docker.CreateConfig{
		Category: "exec",
		UID:      "exec1",
		Name:     "exec",
		Cmd:      []string{"ls"},
		Image:    "nanobox/build",
	}
	docker.CreateContainer(cc)

}