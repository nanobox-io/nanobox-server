package jobs

import (
	"strings"
	"fmt"

	"github.com/pagodabox/nanobox-boxfile"
	"github.com/pagodabox/nanobox-router"

	"github.com/pagodabox/nanobox-server/config"
	"github.com/pagodabox/nanobox-server/util"
)

// grab the original boxfile and loop through the webs
// find all routes and regsiter the routes with the router
// 
func configureRoutes(box boxfile.Boxfile) error {
	newRoutes := []router.Route{}
	webs := box.Nodes("web")
	for _, web := range webs {
		b := box.Node(web)
		container, err := util.GetContainer(web)
		if err != nil {
			// if the container doesnt exist just continue and dont
			// add routes for that node
			continue
		}

		ip := container.NetworkSettings.IPAddress
		for _, route := range routes(b) {
			if len(ports(b)) == 0 {
				route.URLs = []string{"http://"+ip+":8080"}
				newRoutes = append(newRoutes, route)
			}
			for _, to := range ports(b) {
				route.URLs = []string{"http://"+ip+":"+to}
				newRoutes = append(newRoutes, route)
			}
		}
	}

	// add the default route if we dont have one
	defaulted := false
	for _, route := range newRoutes {
		if route.Name == config.App+".nano.dev" && route.Path == "/" {
			defaulted = true
			break
		}
	}
	if !defaulted {
		if web1, err := util.GetContainer("web1"); err == nil {
			newRoutes = append(newRoutes, router.Route{Name: config.App+".nano.dev", Path: "/", URLs: []string{"http://"+web1.NetworkSettings.IPAddress+":8080"}})
		}
	}
	fmt.Println("newRoutes:", newRoutes)
	router.UpdateRoutes(newRoutes)
	router.ErrorHandler = nil
	return nil
}

func configurePorts(box boxfile.Boxfile) error {
	vips, err := util.ListVips()
	if err != nil {
		return err
	}

	// remove all old forwards
	for _, vip := range vips {
		if vip.Port != 80 && vip.Port != 443 {
			for _, server := range vip.Servers {
				util.RemoveForward(server.Host)
			}
		}
	}

	// loop through the boxfile container nodes
	// and add in any new port maps
	nodes := box.Nodes("container")
	for _, node := range nodes {
		b := box.Node(node)
		container, err := util.GetContainer(node)
		if err != nil {
			// if the container doesnt exist just continue and dont
			// add routes for that node
			continue
		}
		ip := container.NetworkSettings.IPAddress
		for from, to := range ports(b) {
			util.AddForward(from, ip, to)
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
			rtn = append(rtn, router.Route{Name: config.App+".nano.dev", Path: routeParts[0]})
		case 2:
			subDomain := strings.Trim(routeParts[0], ".")
			rtn = append(rtn, router.Route{Name: subDomain+"."+config.App+".nano.dev", Path: routeParts[0]})
		}

	}

	return
}

func ports(box boxfile.Boxfile) (map[string]string) {
	rtn := map[string]string{}
	ports, ok := box.Value("ports").([]string)
	if !ok {
		tmps, ok := box.Value("ports").([]interface{})
		if !ok {
			return rtn
		}
		for _, tmp := range tmps {
			if str, ok := tmp.(string); ok {
				ports = append(ports, str)
			}
		}
	}

	for _, port := range ports {
		portParts := strings.Split(port, ":")
		switch len(portParts) {
		case 1:
			rtn[portParts[0]] = portParts[0]
		case 2:
			rtn[portParts[0]] = portParts[1]
		}
	}
	return rtn
}
