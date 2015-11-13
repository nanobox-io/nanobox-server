package main

import (
	"os"
	"net"
	"net/http"
	"testing"
	"time"
	"io/ioutil"
	// "runtime"
	"fmt"
	"encoding/json"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/nanobox-logtap/drain"
	"github.com/nanobox-io/nanobox-logtap/collector"

	"github.com/nanopack/mist/core"
	"github.com/nanobox-io/nanobox-server/api"
	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util/docker"
)

var apiClient = api.Init()

func TestMain(m *testing.M) {
	config.Log = lumber.NewConsoleLogger(lumber.ERROR)

	curDir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	config.MountFolder = curDir + "/test/"
	config.DockerMount = curDir + "/test/"
	config.App, _ = config.AppName()

	config.Logtap.AddDrain("console", drain.AdaptLogger(config.Log))
	config.Logtap.AddDrain("mist", drain.AdaptPublisher(config.Mist))	
	// define logtap collectors/drains; we don't need to defer Close() anything here,
	// because these want to live as long as the server
	if _, err := collector.SyslogUDPStart("app", config.Ports["logtap"]+"1", config.Logtap); err != nil {
		panic(err)
	}

	//
	if _, err := collector.SyslogTCPStart("app", config.Ports["logtap"]+"1", config.Logtap); err != nil {
		panic(err)
	}

	// we will be adding a 0 to the end of the logtap port because we cant have 2 tcp listeneres
	// on the same port
	if _, err := collector.StartHttpCollector("deploy", config.Ports["logtap"]+"0", config.Logtap); err != nil {
		panic(err)
	}

	go func() {
		// start nanobox
		if err := apiClient.Start(config.Ports["api"]); err != nil {
			os.Exit(1)
		}
	}()
	<-time.After(time.Second)
	rtn := m.Run()

	// Remove all containers
	containers, _ := docker.ListContainers()
	for _, container := range containers {
		docker.RemoveContainer(container.ID)
	}

	os.Exit(rtn)
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

func TestDeploy(t *testing.T) {
	r, err := http.Post("http://localhost:1757/deploys?run=true", "json", nil)
	if err != nil || r.StatusCode != 200 {
		fmt.Println(r, err)
		t.Errorf("unable to deploy")
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	deploy := map[string]string{}
	err = json.Unmarshal(bytes, &deploy)
	if err != nil {
		t.Errorf("unable to unmarshal body %s", bytes)
	}

	id := deploy["id"]

	mistClient := mist.NewLocalClient(config.Mist, 1)
	mistClient.Subscribe([]string{"job", "deploy"})

	message := <- mistClient.Messages()

	data := map[string]interface{}{}

	err = json.Unmarshal([]byte(message.Data), &data)
	if err != nil {
		t.Errorf("unable to unmarshal data %s\nerr: %s", message.Data, err.Error())
	}

	if data["document"].(map[string]interface{})["id"] != id {
		t.Errorf("the message is not for my deploy: %+v", data["document"])
	}
	if data["document"].(map[string]interface{})["status"] != "complete" {
		t.Errorf("I recieved a bad status: %+v", data["document"])
	}
	if list, err := docker.ListContainers(); err != nil || len(list) == 0 {
		t.Errorf("I should have atleast one docker container")
	}

	if c, err := docker.GetContainer("build1"); err != nil || c.Name == "" {
		t.Errorf("There should be a build1 container")
	}
}

func TestBuild(t *testing.T) {
	r, err := http.Post("http://localhost:1757/builds", "json", nil)
	if err != nil || r.StatusCode != 200 {
		fmt.Println(r, err)
		t.Errorf("unable to build")
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	build := map[string]string{}
	err = json.Unmarshal(bytes, &build)
	if err != nil {
		t.Errorf("unable to unmarshal body %s", bytes)
	}

	id := build["id"]

	mistClient := mist.NewLocalClient(config.Mist, 1)
	mistClient.Subscribe([]string{"job", "build"})

	message := <- mistClient.Messages()

	data := map[string]interface{}{}

	err = json.Unmarshal([]byte(message.Data), &data)
	if err != nil {
		t.Errorf("unable to unmarshal data %s\nerr: %s", message.Data, err.Error())
	}

	if data["document"].(map[string]interface{})["id"] != id {
		t.Errorf("the message is not for my build: %+v", data["document"])
	}
	if data["document"].(map[string]interface{})["status"] != "complete" {
		t.Errorf("I recieved a bad status: %+v", data["document"])
	}
	if c, err := docker.GetContainer("build1"); err != nil || c.Name == "" {
		t.Errorf("There should be a build1 container")
	}
}

func TestDevelop(t *testing.T) {
	conn, err := net.Dial("tcp4", "localhost:1757")	
	if err != nil {
		t.Errorf("unable to establish connection")
	}

	fmt.Fprintf(conn, "POST /develop? HTTP/1.1\r\n\r\n")

	// give the server time to start the dev
	<-time.After(1 * time.Second)

	if c, err := docker.GetContainer("dev1"); err != nil || c.Name == "" {
		t.Errorf("There should be a dev1 container")
	}
	conn.Close()

	// give the server time to start the dev
	<-time.After(1 * time.Second)
	if _, err := docker.GetContainer("dev1"); err == nil {
		t.Errorf("There should not be a dev1 container")
	}	
}

func TestRoutes(t *testing.T) {
	r, err := http.Get("http://localhost:1757/routes")
	if err != nil || r.StatusCode != 200 {
		fmt.Println(r, err)
		t.Errorf("unable to get routes")
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	routes := []map[string]interface{}{}
	err = json.Unmarshal(bytes, &routes)
	if err != nil {
		t.Errorf("unable to unmarshal body %s", bytes)
	}
	if len(routes) < 1 {
		t.Errorf("I should have one route")
		return
	}
	if routes[0]["Name"] != "app.dev" {
		t.Errorf("the Route I have is not app.dev")
	}
}
