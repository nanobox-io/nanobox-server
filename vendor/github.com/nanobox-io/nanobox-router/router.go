// Copyright (C) Pagoda Box, Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential
package router

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

// A route object from the api
type Route struct {
	SubDomain string `json:"subdomain"`
	Domain    string `json:"subdomain"`
	Path string   `json:"path"`
	Targets []string `json:"targets"`
	Page string   `json:"page"`

	proxies []*proxy
}

// Simple Ip storage for creating Reverse Proxies
type proxy struct {
	URL          string
	reverseProxy *httputil.ReverseProxy
}

var routes = []Route{}
var mutex = sync.Mutex{}

// replace my routes with a new set
func UpdateRoutes(newRoutes []Route) {
	for _, route := range newRoutes {
		for _, target := range route.Targets {
			prox := &proxy{URL: target}
			err := prox.initProxy()
			if err == nil {
				route.proxies = append(route.proxies, prox)
			}
		}
	}
	mutex.Lock()
	routes = newRoutes
	mutex.Unlock()
}

// Show my cached copy of the routes.
// This is used for syncronization.
func Routes() []Route {
	return routes
}

// Establish the ReverseProxy
func (self *proxy) initProxy() error {
	if self.reverseProxy == nil {
		url, err := url.Parse(self.URL)
		if err != nil {
			return err
		}
		self.reverseProxy = httputil.NewSingleHostReverseProxy(url)
	}
	return nil
}

// Start both http and tls servers
func Start(httpAddress, tlsAddress string) error {
	err := StartHTTP(httpAddress)
	if err != nil {
		return err
	}
	return StartTLS(tlsAddress)
}
