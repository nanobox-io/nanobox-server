package logtap

import (
	"github.com/pagodabox/golang-hatchet"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpCollector struct {
	wChan chan Message
	log   hatchet.Logger

	port string
}

// NewHttpCollector creates a new syslog collector
func NewHttpCollector(port string) *HttpCollector {
	return &HttpCollector{
		port:  port,
		wChan: make(chan Message),
	}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the syslog collector
func (h *HttpCollector) SetLogger(l hatchet.Logger) {
	h.log = l
}

// CollectChan grats access to the collection channel
// any logs collected from syslog will be translated into a message
// and dropped on this channel for processing
func (h *HttpCollector) CollectChan() chan Message {
	return h.wChan
}

// Start begins listening to the syslog port transfers all
// syslog messages on the wChan
func (h *HttpCollector) Start() {
	go func() {
		err := http.ListenAndServe(":"+h.port, h)
		if err != nil {
			h.log.Error("[LOGTAP]" + err.Error())
		}

	}()
}

// implement the http.Handler interface so i can just serve all request to this
// port to this single Handler an not worry about pathing at all
func (h *HttpCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	msg := Message{
		Type:     "deploy",
		Time:     time.Now(),
		Priority: priorityInt(r.Header.Get("X-Log-Level")),
		Content:  string(body),
	}
	h.wChan <- msg
}
