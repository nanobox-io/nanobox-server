package router

import "net/http"

type NoDeploy struct {
}

func (self NoDeploy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("NoDeploy"))
}

type DeployInProgress struct {
}

func (self DeployInProgress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("DeployInProgress"))
}

type FailedDeploy struct {
}

func (self FailedDeploy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("FailedDeploy"))
}
