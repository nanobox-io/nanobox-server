package data

import (
	// "time"
	"strings"
	"regexp"
	"encoding/json"

	"github.com/hookyd/go-client"
	"github.com/pagodabox/nanobox-router"
	"github.com/pagodabox/nanobox-logtap"
	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/tasks"
	"github.com/pagodabox/nanobox-server/worker"
)

//
type Sync struct {
}

// this method syncronies your docker containers
// with the boxfile specification
func (s *Sync) Process() {

	config.Log.Debug("[NANOBOX :: SYNC] Started\n")
	ch := make(chan string)
	defer close(ch)

	go func() {
		for data := range ch {
			config.Logtap.Publish("deploy", data)
		}
	}()

	// clear the deploy log
	config.Logtap.Drains["history"].(*logtap.HistoricalDrain).ClearDeploy()


	// set routing to watch logs
	config.Log.Debug("[NANOBOX :: SYNC] setting routes\n")
	config.Router.Handler = router.DeployInProgress{}

	// remove all code containers
	config.Log.Debug("[NANOBOX :: SYNC] clearing old containers")

	containers, _ := tasks.ListContainers("code", "build")
	config.Log.Debug("[NANOBOX :: SYNC] containers (%#v)", containers)
	for _, container := range containers {
		config.Log.Debug("[NANOBOX :: SYNC] clean container %#v\n", container)
		err := tasks.RemoveContainer(container.Id)
		if err != nil {
			config.Log.Error("[NANOBOX :: SYNC] clean error %s\n", err.Error())
		}
	}

	// wipe the data dir /var/nanobox/deploy/
	if err := tasks.Clean(ch); err != nil {
		ch <- "could not clean directories"
		config.Log.Error("could not clean directories(%s)", err.Error())
		config.Router.Handler = router.FailedDeploy{}
		return
	}

	config.Log.Debug("[NANOBOX :: SYNC] copying code from /vagrant/code/%s/* to /var/nanobox/deploy/", config.App)
	if err := tasks.Copy(ch); err != nil {
		ch <- "could not copy build image container"
		config.Log.Error("could not copy build stuff(%s)", err.Error())
		config.Router.Handler = router.FailedDeploy{}
		return
	}

	// create a build container
	config.Log.Debug("[NANOBOX :: SYNC] creating container")
	con, err := tasks.CreateContainer("nanobox/base", map[string]string{"build": "true", "uid": "build"})
	if err != nil {
		ch <- "could not create build image container"
		config.Log.Error("could not create build image container(%s)", err.Error())
		config.Router.Handler = router.FailedDeploy{}
		return
	}

	config.Log.Debug("[NANOBOX :: SYNC] container created %#v", con)

	addr := con.NetworkSettings.IPAddress

	h := hooky.Hooky{
		Host: addr,
		Port: "1234", // dont know the port
	}

	box := boxfile.NewFromPath("/vagrant/code/"+config.App+"/Boxfile")

	payload := map[string]interface{}{
		"boxfile": box.Node("build").Parsed,
		"logtap_uri": config.LogtapURI,
	}

	config.Log.Debug("[NANOBOX :: SYNC] running sniff hook")

	cPayload, _ := json.Marshal(payload)

	if response, err := h.Run("configure", cPayload, "0"); err != nil {
		config.Log.Debug("[NANOBOX :: SYNC] hook response(%#v) err(%#v)", response, err)
		config.Router.Handler = router.FailedDeploy{}
		// return
	}

	response, err := h.Run("boxfile", "{}", "1")
	if err != nil {
		config.Log.Debug("[NANOBOX :: SYNC] hook response(%#v) err(%#v)", response, err)
		config.Router.Handler = router.FailedDeploy{}
		// return
	} else {
		// combine boxfiles
		hookBox := boxfile.New([]byte(response.Out))
		box.Merge(hookBox)
	}

	serviceWorker := worker.New()
	serviceWorker.Blocking = true
	serviceWorker.Concurrent = true

	serviceStarts := []*ServiceStart{}

	// build containers according to boxfile
	// run hooks in new containers (services)
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node.(string), "")
		config.Log.Debug("[NANOBOX :: SYNC] looking at node (%s)", node)
		if node != "nanobox" &&
			node != "global" &&
			node != "build" &&
			name != "web" &&
			name != "worker" {
				config.Log.Debug("[NANOBOX :: SYNC] looks good making sure we dont have one already")
			if _, err := tasks.GetContainer(node.(string)); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node.(string),
				}
				config.Log.Debug("[NANOBOX :: SYNC] sure didnt (%#v)", s)
				serviceStarts = append(serviceStarts, &s)

				serviceWorker.Queue(&s)
			}
		}
	}

	serviceWorker.Process()

	evars := map[string]string{}

	for _, s := range serviceStarts {
		if !s.Success {
			ch <- "A Service was not started correctly ("+s.Uid+")"
			config.Log.Error("[NANOBOX :: SYNC] A service failed to start correctly (%s)", s.Uid)
			config.Router.Handler = router.FailedDeploy{}
			return
		}

		for key, val := range s.EnvVars {
			evars[strings.ToUpper(s.Uid+"_"+key)] = val
		}
	}

	payload["env_vars"] = evars

	pload, _ := json.Marshal(payload)
	response, err = h.Run("build", string(pload), "2")
	if err != nil {
		config.Log.Debug("[NANOBOX :: SYNC] hook response(%#v) err(%#v)", response, err)
		config.Router.Handler = router.FailedDeploy{}
		// return
	}

	// remove build
	config.Log.Debug("[NANOBOX :: SYNC] remove build container")
	tasks.RemoveContainer(con.Id)

	// run hooks in new containers (code)
	codeWorker := worker.New()
	codeWorker.Blocking = true
	codeWorker.Concurrent = true

	codeServices := []*ServiceStart{}
	for _, node := range box.Nodes() {
		name := regexp.MustCompile(`\d+`).ReplaceAllString(node.(string), "")
		if name == "web" || name == "worker" {
			if _, err := tasks.GetContainer(node.(string)); err != nil {
				s := ServiceStart{
					Boxfile: box.Node(node),
					Uid:     node.(string),
				}
				codeServices = append(codeServices, &s)

				codeWorker.Queue(&s)
			}
		}
	}

	codeWorker.Process()

	for _, s := range codeServices {
		if !s.Success {
			ch <- "A Service was not started correctly ("+s.Uid+")"
			config.Log.Error("[NANOBOX :: SYNC] A service failed to start correctly (%s)", s.Uid)
			config.Router.Handler = router.FailedDeploy{}
			return
		}
	}


	// before deploy hooks
	for _, node := range box.Nodes() {
		n := node.(string)
		bd  := box.Node(n).Value("before_deploy")
		bda := box.Node(n).Value("before_deploy_all")
		if bd != nil || bda != nil {
			if container, err := tasks.GetContainer(n); err == nil {
				dc, _ := tasks.GetDetailedContainer(container.Id)
				addr := dc.NetworkSettings.IPAddress

				h := hooky.Hooky{
					Host: addr,
					Port: "1234", // dont know the port
				}

				pload, _ := json.Marshal(map[string]interface{}{"before_deploy": bd, "before_deploy_all": bda})

				if response, err := h.Run("before_deploy", pload, "0"); err != nil {
					config.Log.Debug("[NANOBOX :: SYNC] hook response(%#v) err(%#v)", response, err)
					config.Router.Handler = router.FailedDeploy{}
					// return
				}

			}
		}
	}


	// set routing to web components
	config.Log.Debug("[NANOBOX :: SYNC] set routing")
	if container, err := tasks.GetContainer("web1"); err == nil {
		dc, _ := tasks.GetDetailedContainer(container.Id)

		config.Router.AddTarget("/", dc.NetworkSettings.IPAddress)
		config.Router.Handler = nil
	}
	
	// after deploy hooks
	for _, node := range box.Nodes() {
		n := node.(string)
		ad  := box.Node(n).Value("after_deploy")
		ada := box.Node(n).Value("after_deploy_all")
		if ad != nil || ada != nil {
			if container, err := tasks.GetContainer(n); err == nil {
				dc, _ := tasks.GetDetailedContainer(container.Id)
				addr := dc.NetworkSettings.IPAddress

				h := hooky.Hooky{
					Host: addr,
					Port: "1234", // dont know the port
				}

				pload, _ := json.Marshal(map[string]interface{}{"before_deploy": ad, "before_deploy_all": ada})

				if response, err := h.Run("before_deploy", pload, "0"); err != nil {
					config.Log.Debug("[NANOBOX :: SYNC] hook response(%#v) err(%#v)", response, err)
					config.Router.Handler = router.FailedDeploy{}
					// return
				}

			}
		}
	}

	config.Log.Debug("[NANOBOX :: SYNC] sync complete")

}
