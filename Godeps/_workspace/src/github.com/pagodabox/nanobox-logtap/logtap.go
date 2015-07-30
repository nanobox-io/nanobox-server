package logtap

// Overview

import (
	"github.com/pagodabox/golang-hatchet"
	"reflect"
	"time"
)

const (
	FATAL = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type Collector interface {
	CollectChan() chan Message
	SetLogger(hatchet.Logger)
}

type Drain interface {
	Write(Message)
	SetLogger(hatchet.Logger)
}

type Message struct {
	Type     string
	Time     time.Time
	Priority int
	Content  string
}

type Logtap struct {
	log        hatchet.Logger
	Collectors map[string]Collector
	Drains     map[string]Drain
}

// Establishes a new logtap object
// and makes sure it has the some logger
func New(log hatchet.Logger) *Logtap {
	if log == nil {
		log = hatchet.DevNullLogger{}
	}
	return &Logtap{
		log:        log,
		Collectors: make(map[string]Collector),
		Drains:     make(map[string]Drain),
	}
}

// AddDrain addes a drain to the listeners and sets its logger
func (l *Logtap) AddDrain(tag string, d Drain) {
	d.SetLogger(l.log)
	l.Drains[tag] = d
}

// RemoveDrain drops a drain
func (l *Logtap) RemoveDrain(tag string) {
	delete(l.Drains, tag)
}

// AddCollector adds a collector and begins listening to it
// also adds logging
func (l *Logtap) AddCollector(tag string, c Collector) {
	c.SetLogger(l.log)
	l.Collectors[tag] = c
}

// RemoveCollector remove given collector
func (l *Logtap) RemoveCollector(tag string) {
	delete(l.Collectors, tag)
}

// Start begins listening to all the collectors that are registered.
// it then broadcasts all messages to all the drains that are registered
// this is backgrounded so it can be used by a parent process without getting in the way
func (l *Logtap) Start() {
	go func() {
		for {
			cases := make([]reflect.SelectCase, len(l.Collectors))
			index := 0
			for _, col := range l.Collectors {
				cases[index] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(col.CollectChan())}
				index++
			}

			_, value, ok := reflect.Select(cases)
			// ok will be true if the channel has not been closed.
			if ok {
				l.log.Debug("[LOGTAP][start][collect] %v", value.Interface().(Message))
				l.writeMessage(value.Interface().(Message))
			}

		}

	}()
}

func (l *Logtap) Publish(category string, priority int, content string) {
	m := Message{
		Type:     category,
		Time:     time.Now(),
		Priority: priority,
		Content:  content,
	}
	l.writeMessage(m)
}

// writeMessage broadcasts to all drains in seperate go routines
func (l *Logtap) writeMessage(msg Message) {
	for _, drain := range l.Drains {
		go drain.Write(msg)
	}
}

func priorityString(priority int) string {
	switch priority {
	default:
		return "info"
	case FATAL:
		return "fatal"
	case ERROR:
		return "error"
	case WARN:
		return "warn"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	}
}

func priorityInt(priority string) int {
	switch priority {
	default:
		return INFO
	case "fatal":
		return FATAL
	case "error":
		return ERROR
	case "warn":
		return WARN
	case "info":
		return INFO
	case "debug":
		return DEBUG
	}
}
