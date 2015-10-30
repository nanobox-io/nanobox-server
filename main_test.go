package main

import (
	"os"
	"net/http"
	"testing"
	"time"
	"io/ioutil"
	"runtime"
	"os/exec"
	"strings"
	"fmt"

	"github.com/nanobox-io/nanobox-server/api"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util/docker"
)

var apiClient = api.Init()

func TestMain(m *testing.M) {
	curDir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	config.MountFolder = curDir + "test/"
	config.App, _ = config.AppName()

	if runtime.GOOS != "linux" {
		cmd := exec.Command("docker-machine", "url", "default")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("failed to find docker-machine on non linux system")
			os.Exit(1)
		}
		url := strings.TrimSpace(string(out))
		config.DockerEndPoint = url
		docker.ResetDefaults()
	}
	
	go func() {
		// start nanobox
		if err := apiClient.Start(config.Ports["api"]); err != nil {
			os.Exit(1)
		}
	}()
	<-time.After(time.Second)
	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	r, err := http.Get("http://localhost:1757/ping")
	if err != nil || r.StatusCode != 200 {
		t.Errorf("unable to ping")
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	body := string(bytes)
	if body != "pong" {
		t.Errorf("expected pong but got %s", body)
	}
}

// func TestDeploy(t *testing.T) {
// 	fmt.Println(docker.ListContainers())
// 	r, err := http.Post("http://localhost:1757/deploy?run=true", "json", nil)
// 	if err != nil || r.StatusCode != 200 {
// 		t.Errorf("unable to deploy")
// 	}

// }

