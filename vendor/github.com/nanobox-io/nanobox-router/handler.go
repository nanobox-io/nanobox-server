// Copyright (C) Pagoda Box, Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential
package router

import (
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
)

// the number used to % for round robin requests
var robiner = uint32(0)

type handler struct {
	https bool
}

// Implement the http.Handler interface. Also let clients know when I have
// no matching route listeners
func (self handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if ErrorHandler != nil {
		ErrorHandler.ServeHTTP(rw, req)
		return
	}
	if self.https {
		req.Header.Set("X-Forwarded-Proto", "https")
	}

	re := regexp.MustCompile(`:\d+`) // used to remove the port from the host
	host := string(re.ReplaceAll([]byte(req.Host), nil))
	route := bestMatch(host, req.URL.Path)
	if route != nil {
		if route.Page != "" {
			rw.Write([]byte(route.Page))
			return
		}
		if len(route.proxies) == 0 {
			NoRoutes{}.ServeHTTP(rw, req)
			return
		}
		proxy := route.proxies[atomic.AddUint32(&robiner, 1)%uint32(len(route.proxies))]
		proxy.reverseProxy.ServeHTTP(rw, req)
		return
	}
	// if i get to here i have no idea where to send your request
}

// route and subdomain matching system.
// This first makes sure the domain matches in a recursive manor
// example: sub.domain.com is requested and we recursively strip subdomains
// until we find a match. Then score the path match and confirm it is a match.
func bestMatch(host, path string) (route *Route) {
	matchScore := 0
	for _, r := range routes {
		if domainMatch(host, r) && pathMatch(path, r) && matchScore < len(r.Path) {
			route = &r
			matchScore = len(r.Path)
		}
	}

	if route == nil {
		hostParts := strings.Split(host, ".")
		if len(hostParts) <= 2 {
			return nil
		}
		return bestMatch(strings.Join(hostParts[1:], "."), path)
	}
	return route
}

// match the different parts of the domain
func domainMatch(requestHost string, r Route) bool {
	switch {
	case r.Domain != "" && r.SubDomain != "":
		return requestHost == (r.SubDomain + "." + r.Domain)
	case r.Domain != "":
		return strings.HasSuffix(requestHost, r.Domain)
	case r.SubDomain != "":
		// remove domain.com part of the host string
		hostParts := strings.Split(requestHost, ".")
		if len(hostParts) < 2 {
			return false
		}
		subHost := strings.Join(hostParts[:len(hostParts)-2], ".")
		
		// direct compare
		return subHost == r.SubDomain
	}
	return true
}

func pathMatch(requestPath string , r Route) bool {
	return strings.HasPrefix(requestPath, r.Path)
}
