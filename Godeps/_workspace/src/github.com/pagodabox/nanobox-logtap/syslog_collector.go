package logtap

import (
	"github.com/jeromer/syslogparser/rfc3164"
	"github.com/jeromer/syslogparser/rfc5424"
	"github.com/pagodabox/golang-hatchet"
	"net"
	"strconv"
	"time"
)

type SyslogCollector struct {
	wChan chan Message
	log   hatchet.Logger

	Port string
}

// NewSyslogCollector creates a new syslog collector
func NewSyslogCollector(port string) *SyslogCollector {
	return &SyslogCollector{
		Port:  port,
		wChan: make(chan Message),
	}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the syslog collector
func (s *SyslogCollector) SetLogger(l hatchet.Logger) {
	s.log = l
}

// CollectChan grats access to the collection channel
// any logs collected from syslog will be translated into a message
// and dropped on this channel for processing
func (s *SyslogCollector) CollectChan() chan Message {
	return s.wChan
}

// Start begins listening to the syslog port transfers all
// syslog messages on the wChan
func (s *SyslogCollector) Start() {
	go func() {

		udpAddress, err := net.ResolveUDPAddr("udp4", ("0.0.0.0:" + s.Port))
		if err != nil {
			s.log.Error("[LOGTAP][syslog]resolving UDP address on ", s.Port)
			s.log.Error("[LOGTAP][syslog]" + err.Error())
			return
		}

		conn, err := net.ListenUDP("udp", udpAddress)
		if err != nil {
			s.log.Error("[LOGTAP][syslog]listening on UDP port ", s.Port)
			s.log.Error("[LOGTAP][syslog]" + err.Error())
			return
		}
		defer conn.Close()

		var buf []byte = make([]byte, 1024)
		for {
			s.log.Debug("[LOGTAP][syslog][start] listen")
			n, address, err := conn.ReadFromUDP(buf)
			if err != nil {
				s.log.Error("[LOGTAP][syslog]reading data from connection")
				s.log.Error("[LOGTAP][syslog]" + err.Error())
			}
			if address != nil {
				s.log.Debug("[LOGTAP][syslog][start] got message from " + address.String() + " with n = " + strconv.Itoa(n))
				if n > 0 {
					msg := s.parseMessage(buf[0:n])
					s.log.Debug("[LOGTAP][syslog][start] msg content: " + msg.Content)
					s.wChan <- msg
				}
			}
		}
	}()
}

// parseMessage parses the syslog message and returns a msg
// if the msg is not parsable or a standard formatted syslog message
// it will drop the whole message into the content and make up a timestamp
// and a priority
func (s *SyslogCollector) parseMessage(b []byte) (msg Message) {
	msg.Type = "syslog"
	p := rfc3164.NewParser(b)
	err := p.Parse()
	if err == nil {
		parsedData := p.Dump()
		// fmt.Printf("%#v\n",parsedData)
		msg.Time = parsedData["timestamp"].(time.Time)
		msg.Priority = adjustInt(parsedData["priority"].(int) % 8)
		msg.Content = parsedData["tag"].(string) + " " + parsedData["content"].(string)
	} else {
		p := rfc5424.NewParser(b)
		err := p.Parse()
		if err == nil {
			parsedData := p.Dump()
			// fmt.Printf("%#v\n",parsedData)
			msg.Time = parsedData["timestamp"].(time.Time)
			msg.Priority = adjustInt(parsedData["priority"].(int) % 8)
			msg.Content = parsedData["tag"].(string) + " " + parsedData["content"].(string)
		} else {
			s.log.Error("[LOGTAP]Unable to parse data: " + string(b))
			msg.Time = time.Now()
			msg.Priority = 1
			msg.Content = string(b)
		}
	}
	return
}

// I need to adjust the possible prioritys from rfc3164 and rfc5424
// to the 5 priority options.
func adjustInt(in int) int {
	if in < 3 {
		return 0
	}
	if in < 5 {
		return in - 2
	}
	return in - 3
}
