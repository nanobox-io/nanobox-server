package docker

import (
	"fmt"
	"io"
	"os"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/nanobox-io/nanobox-boxfile"

	"github.com/nanobox-io/nanobox-server/config"
)

type ClientInterface interface {
	ListImages(opts dc.ListImagesOptions) ([]dc.APIImages, error)
	PullImage(opts dc.PullImageOptions, auth dc.AuthConfiguration) error
	CreateContainer(opts dc.CreateContainerOptions) (*dc.Container, error)
	StartContainer(id string, hostConfig *dc.HostConfig) error
	KillContainer(opts dc.KillContainerOptions) error
	ResizeContainerTTY(id string, height, width int) error
	StopContainer(id string, timeout uint) error
	RemoveContainer(opts dc.RemoveContainerOptions) error
	WaitContainer(id string) (int, error)
	InspectContainer(id string) (*dc.Container, error)
	ListContainers(opts dc.ListContainersOptions) ([]dc.APIContainers, error)
	CreateExec(opts dc.CreateExecOptions) (*dc.Exec, error)
	ResizeExecTTY(id string, height, width int) error
	StartExec(id string, opts dc.StartExecOptions) error
	InspectExec(id string) (*dc.ExecInspect, error)
}

type DockerDefault interface {
	CreateContainer(conf CreateConfig) (*dc.Container, error)
	StartContainer(id string) error
	KillContainer(id, sig string) error
	ResizeContainerTTY(id string, height, width int) error
	WaitContainer(id string) (int, error)
	RemoveContainer(id string) error
	InspectContainer(id string) (*dc.Container, error)
	GetContainer(id string) (*dc.Container, error)
	ListContainers(labels ...string) ([]*dc.Container, error)
	InstallImage(image string) error
	ListImages() ([]dc.APIImages, error)
	ImageExists(name string) bool
	ExecInContainer(container string, args ...string) ([]byte, error)
	CreateExec(id string, cmd []string, in, out, err bool) (*dc.Exec, error)
	ResizeExecTTY(id string, height, width int) error
	RunExec(exec *dc.Exec, in io.Reader, out io.Writer, err io.Writer) (*dc.ExecInspect, error)
}

type DockerUtil struct {
}

var Client ClientInterface

var Default DockerDefault

func init() {
	Client, _ = dc.NewClient("unix:///var/run/docker.sock")
	Default = DockerUtil{}
}

func InstallImage(image string) error {
	return Default.InstallImage(image)
}
func ListImages() ([]dc.APIImages, error) {
	return Default.ListImages()
}
func ImageExists(image string) bool {
	return Default.ImageExists(image)
}
func CreateContainer(conf CreateConfig) (*dc.Container, error) {
	return Default.CreateContainer(conf)
}
func StartContainer(id string) error {
	return Default.StartContainer(id)
}
func KillContainer(id, sig string) error {
	return Default.KillContainer(id, sig)
}
func ResizeContainerTTY(id string, height, width int) error {
	return Default.ResizeContainerTTY(id, height, width)
}
func WaitContainer(id string) (int, error) {
	return Default.WaitContainer(id)
}
func RemoveContainer(id string) error {
	return Default.RemoveContainer(id)
}
func InspectContainer(id string) (*dc.Container, error) {
	return Default.InspectContainer(id)
}
func GetContainer(id string) (*dc.Container, error) {
	return Default.GetContainer(id)
}
func ListContainers(labels ...string) ([]*dc.Container, error) {
	return Default.ListContainers(labels...)
}
func ExecInContainer(container string, args ...string) ([]byte, error) {
	return Default.ExecInContainer(container, args...)
}
func CreateExec(id string, cmd []string, in, out, err bool) (*dc.Exec, error) {
	return Default.CreateExec(id, cmd, in, out, err)
}
func ResizeExecTTY(id string, height, width int) error {
	return Default.ResizeExecTTY(id, height, width)
}
func RunExec(exec *dc.Exec, in io.Reader, out io.Writer, err io.Writer) (*dc.ExecInspect, error) {
	return Default.RunExec(exec, in, out, err)
}

// These functions are bandaids. I will be removing them once I have a clear place to put them
func libDirs() (rtn []string) {
	box := combinedBox()
	libDirs, ok := box.Node("build").Value("lib_dirs").([]interface{})
	if ok && !box.Node("console").BoolValue("ignore_lib_dirs") {
		for _, libDir := range libDirs {
			strDir, ok := libDir.(string)
			if ok && isDir("/mnt/sda/var/nanobox/cache/lib_dirs/"+strDir) {
				rtn = append(rtn, fmt.Sprintf("/mnt/sda/var/nanobox/cache/lib_dirs/%s/:/code/%s/", strDir, strDir))
			}
		}
	}
	return
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func combinedBox() boxfile.Boxfile {
	box := boxfile.NewFromPath("/vagrant/code/" + config.App + "/Boxfile")
	// run boxfile script (blocking)
	if !box.Node("build").BoolValue("disable_engine_boxfile") {
		out, err := ExecInContainer("build1", "/opt/bin/default-boxfile", "{}")
		if err == nil {
			eBox := boxfile.New([]byte(out))
			box.Merge(eBox)
		}
	}
	return box
}
