// Copyright (C) Pagoda Box, Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential
package router

import ()

var routes = []Route{}

// Update my domains with a new set of routes
func UpdateRoutes(newRoutes []Route) {
	newDomains := []Domain{}
	for _, route := range newRoutes {
		dom := Domain{Name: route.Name, Path: route.Path, proxies: []*Proxy{}}
		if route.Page != "" {
			// if im given a page for the route
			// do not populate the proxies.
			dom.Page = []byte(route.Page)
		}
		for _, url := range route.URLs {
			prox := &Proxy{URL: url}
			err := prox.initProxy()
			if err == nil {
				dom.proxies = append(dom.proxies, prox)
			}

		}
		newDomains = append(newDomains, dom)
	}
	domainLock.Lock()
	routes = newRoutes
	domains = newDomains
	domainLock.Unlock()
}

// Show my cached copy of the routes.
// This is used for syncronization.
func Routes() []Route {
	return routes
}
