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
	Name string   `json:"name"`
	Path string   `json:"path"`
	URLs []string `json:"urls"`
	Page string   `json:"page"`
}

var domainLock = sync.Mutex{}

var domains = []Domain{}

// A Domain representation
// used for matching routes to web requests.
// It also knows how to forward requests to the appropriate servers.
type Domain struct {
	Name    string
	Path    string
	proxies []*Proxy
	Page    []byte
}

// Simple Ip storage for creating Reverse Proxies
type Proxy struct {
	URL          string
	reverseProxy *httputil.ReverseProxy
}

// Establish the ReverseProxy
func (self *Proxy) initProxy() error {
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
