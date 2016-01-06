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
	domain := bestMatch(host, req.URL.Path)
	if domain != nil {
		if domain.Page != nil {
			rw.Write(domain.Page)
			return
		}
		if len(domain.proxies) == 0 {
			NoRoutes{}.ServeHTTP(rw, req)
			return
		}
		proxy := domain.proxies[atomic.AddUint32(&robiner, 1)%uint32(len(domain.proxies))]
		proxy.reverseProxy.ServeHTTP(rw, req)
		return
	}
	// if i get to here i have no idea where to send your request
}

// route and subdomain matching system.
// This first makes sure the domain matches in a recursive manor
// example: sub.domain.com is requested and we recursively strip subdomains
// until we find a match. Then score the path match and confirm it is a match.
func bestMatch(host, path string) *Domain {
	var dom *Domain
	matchScore := 0
	for i := 0; i < len(domains); i++ {
		if domains[i].Name == host && strings.HasPrefix(path, domains[i].Path) && matchScore < len(domains[i].Path) {
			dom = &domains[i]
			matchScore = len(domains[i].Path)
		}
	}
	if dom == nil {
		hostParts := strings.Split(host, ".")
		if len(hostParts) <= 2 {
			return nil
		}
		return bestMatch(strings.Join(hostParts[1:], "."), path)
	}
	return dom
}
