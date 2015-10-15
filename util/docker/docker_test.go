package docker_test

import (
	"fmt"
	dc "github.com/fsouza/go-dockerclient"
)

type TestClient struct {
	CallLog []string
}

func (t *TestClient) ListImages(opts dc.ListImagesOptions) ([]dc.APIImages, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ListImages()"))
	return []dc.APIImages{}, nil
}
func (t *TestClient) PullImage(opts dc.PullImageOptions, auth dc.AuthConfiguration) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("PullImage()"))
	return nil
}
func (t *TestClient) CreateContainer(opts dc.CreateContainerOptions) (*dc.Container, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("CreateContainer()"))
	return *dc.Container{}, nil
}
func (t *TestClient) StartContainer(id string, hostConfig *dc.HostConfig) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StartContainer()"))
	return nil
}
func (t *TestClient) KillContainer(opts dc.KillContainerOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("KillContainer()"))
	return nil
}
func (t *TestClient) ResizeContainerTTY(id string, height, width int) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ResizeContainerTTY()"))
	return nil
}
func (t *TestClient) StopContainer(id string, timeout uint) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StopContainer()"))
	return nil
}
func (t *TestClient) RemoveContainer(opts dc.RemoveContainerOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("RemoveContainer()"))
	return nil
}
func (t *TestClient) WaitContainer(id string) (int, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("WaitContainer()"))
	return 0, nil
}
func (t *TestClient) InspectContainer(id string) (*dc.Container, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("InspectContainer()"))
	return *dc.Container{}, nil
}
func (t *TestClient) ListContainers(opts dc.ListContainersOptions) ([]dc.APIContainers, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ListContainers()"))
}
func (t *TestClient) CreateExec(opts dc.CreateExecOptions) (*dc.Exec, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("CreateExec()"))
}
func (t *TestClient) ResizeExecTTY(id string, height, width int) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("ResizeExecTTY()"))
	return nil
}
func (t *TestClient) StartExec(id string, opts dc.StartExecOptions) error {
	t.CallLog = append(t.CallLog, fmt.Sprintf("StartExec()"))
	return nil
}
func (t *TestClient) InspectExec(id string) (*dc.ExecInspect, error) {
	t.CallLog = append(t.CallLog, fmt.Sprintf("InspectExec()"))
}