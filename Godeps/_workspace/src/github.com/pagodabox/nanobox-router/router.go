package router

import (
	"github.com/pagodabox/golang-hatchet"
	"net"
	"net/http"
)

// Router is the device by which you create routing rules
type Router struct {
	Handler  http.Handler
	log      hatchet.Logger
	Targets  map[string]string
	Port     string
	Forwards map[string]*net.TCPListener
}

// New creates a new router sets its logger and returns a pointer to the Router
// object
func New(port string, logger hatchet.Logger) *Router {

	if logger == nil {
		logger = hatchet.DevNullLogger{}
	}

	r := &Router{
		log:      logger,
		Port:     port,
		Targets:  make(map[string]string),
		Forwards: make(map[string]*net.TCPListener),
	}

	r.start()

	return r
}

// handleError handle errors and log data It does not stop execution
func (r *Router) handleError(err error) {
	if err != nil {
		r.log.Error(err.Error())
	}
}
