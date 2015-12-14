package jobs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nanobox-io/nanobox-boxfile"
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanobox-io/nanobox-server/config"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
	"github.com/nanobox-io/nanobox-server/util/script"
)

var userBoxfile *boxfile.Boxfile
var engineBoxfile *boxfile.Boxfile
var combinedBoxfile *boxfile.Boxfile

func init() {
	// on start pull the cached boxfile if it is there
	box := boxfile.NewFromPath(config.CachedBox)
	if box.Valid {
		combinedBoxfile = &box
	}
}

// grab the original boxfile and loop through the webs
// find all routes and regsiter the routes with the router
//
func configureRoutes(box boxfile.Boxfile) error {
	newRoutes := []router.Route{}
	webs := box.Nodes("web")
	for _, web := range webs {
		b := box.Node(web)
		container, err := docker.GetContainer(web)
		if err != nil {
			// if the container doesnt exist just continue and dont
			// add routes for that node
			continue
		}

		ip := container.NetworkSettings.IPAddress
		for _, route := range routes(b) {
			if len(ports(b)) == 0 {
				route.URLs = []string{"http://" + ip + ":8080"}
			}
			fmt.Printf("web:ports: %+v\n", ports(b))
			for _, to := range ports(b) {
				route.URLs = append(route.URLs, "http://"+ip+":"+to)
			}
			newRoutes = append(newRoutes, route)
		}
	}

	// add the default route if we dont have one
	defaulted := false
	for _, route := range newRoutes {
		if route.Name == config.App()+".dev" && route.Path == "/" {
			defaulted = true
			break
		}
	}
	if !defaulted {
		if web1, err := docker.GetContainer("web1"); err == nil {
			ip := web1.NetworkSettings.IPAddress
			route := router.Route{Name: config.App() + ".dev", Path: "/"}
			b := box.Node("web1")
			if len(ports(b)) == 0 {
				route.URLs = []string{"http://" + ip + ":8080"}
			}
			for _, to := range ports(b) {
				route.URLs = append(route.URLs, "http://"+ip+":"+to)
			}
			newRoutes = append(newRoutes, route)
		}
	}
	router.UpdateRoutes(newRoutes)
	router.ErrorHandler = nil
	return nil
}

func clearPorts() {
	vips, err := util.ListVips()
	if err != nil {
		return
	}

	// remove all old forwards
	for _, vip := range vips {
		if vip.Port != 80 && vip.Port != 443 {
			for _, server := range vip.Servers {
				util.RemoveForward(server.Host)
			}
		}
	}
}

func configurePorts(box boxfile.Boxfile) error {

	// loop through the boxfile container nodes
	// and add in any new port maps
	nodes := box.Nodes("container")
	for _, node := range nodes {
		b := box.Node(node)
		container, err := docker.GetContainer(node)
		if err != nil {
			// if the container doesnt exist just continue and dont
			// add routes for that node
			config.Log.Debug("no container for %s", node)
			continue
		}
		ip := container.NetworkSettings.IPAddress
		for from, to := range ports(b) {
			err := util.AddForward(from, ip, to)
			if err != nil {
				config.Log.Debug("failed to add forward %+v", err)
			}
		}
	}
	return nil
}

func routes(box boxfile.Boxfile) (rtn []router.Route) {
	boxRoutes, ok := box.Value("routes").([]string)
	if !ok {
		tmps, ok := box.Value("routes").([]interface{})
		if !ok {
			return
		}
		for _, tmp := range tmps {
			if str, ok := tmp.(string); ok {
				boxRoutes = append(boxRoutes, str)
			}
		}
	}
	for _, route := range boxRoutes {
		routeParts := strings.Split(route, ":")
		switch len(routeParts) {
		case 1:
			rtn = append(rtn, router.Route{Name: config.App() + ".dev", Path: routeParts[0]})
		case 2:
			subDomain := strings.Trim(routeParts[0], ".")
			rtn = append(rtn, router.Route{Name: subDomain + "." + config.App() + ".dev", Path: routeParts[0]})
		}

	}

	return
}

func ports(box boxfile.Boxfile) map[string]string {
	rtn := map[string]string{}
	ports, ok := box.Value("ports").([]interface{})
	if !ok {
		return rtn
	}
	for _, port := range ports {
		p, ok := port.(string)
		if ok {
			portParts := strings.Split(p, ":")
			switch len(portParts) {
			case 1:
				rtn[portParts[0]] = portParts[0]
			case 2:
				rtn[portParts[0]] = portParts[1]
			}
		}
		portInt, ok := port.(int)
		if ok {
			rtn[strconv.Itoa(portInt)] = strconv.Itoa(portInt)
		}

	}
	return rtn
}

func UserBoxfile(refresh bool) *boxfile.Boxfile {
	// clear the cached boxfile if we need to
	if refresh == true {
		userBoxfile = nil
	}

	// return the cached one if we have it
	if userBoxfile != nil {
		return userBoxfile
	}

	// create a new one if we didnt have one
	box := boxfile.NewFromPath(config.MountFolder + "code/" + config.App() + "/Boxfile")
	userBoxfile = &box

	return userBoxfile
}

func EngineBoxfile(refresh bool) *boxfile.Boxfile {
	// clear the cached boxfile if we need to
	if refresh == true {
		engineBoxfile = nil
	}

	// return the cached one if we have it
	if engineBoxfile != nil {
		return engineBoxfile
	}

	// create a new one if we didnt have one
	if !UserBoxfile(false).Node("build").BoolValue("disable_engine_boxfile") {
		pload := map[string]interface{}{
			"platform":    "local",
			"boxfile":     UserBoxfile(false).Node("build").Parsed,
			"logtap_host": config.LogtapHost,
		}
		if out, err := script.Exec("default-boxfile", "build1", pload); err == nil {
			box := boxfile.New([]byte(out))
			engineBoxfile = &box
		}
	}

	return engineBoxfile
}

func CombinedBoxfile(refresh bool) *boxfile.Boxfile {
	// clear the cached boxfile if we need to
	if refresh == true {
		combinedBoxfile = nil
	}

	// return the cached one if we have it
	if combinedBoxfile != nil {
		return combinedBoxfile
	}

	box := UserBoxfile(false)
	if eBox := EngineBoxfile(false); eBox != nil {
		box.Merge(*eBox)
	}
	combinedBoxfile = box
	// save the combined boxfile to a file so can recover from crashes
	go combinedBoxfile.SaveToPath(config.CachedBox)

	return combinedBoxfile
}

func DefaultEVars(box boxfile.Boxfile) map[string]string {
	evar := map[string]string{}
	if box.Node("env").Valid {
		b := box.Node("env")
		for key, _ := range b.Parsed {
			val := b.StringValue(key)
			if val != "" {
				evar[key] = val
			}
		}
	}

	evar["APP_NAME"] = config.App()
	return evar
}

func SetLibDirs() {
	dockerLibDirs := []string{}
	box := CombinedBoxfile(false)
	libDirs, ok := box.Node("build").Value("lib_dirs").([]interface{})
	if ok && !box.Node("dev").BoolValue("ignore_lib_dirs") {
		for _, libDir := range libDirs {
			strDir, ok := libDir.(string)
			if ok && isDir("/mnt/sda/var/nanobox/cache/lib_dirs/"+strDir) {
				dockerLibDirs = append(dockerLibDirs, fmt.Sprintf("/mnt/sda/var/nanobox/cache/lib_dirs/%s/:/code/%s/", strDir, strDir))
			}
		}
	}
	docker.LibDirs = dockerLibDirs
}

func SetWorkingDir() {
	docker.WorkingDir = CombinedBoxfile(false).Node("dev").StringValue("working_dir")
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
