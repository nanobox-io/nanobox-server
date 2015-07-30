package logtap

import "fmt"
import "github.com/pagodabox/golang-hatchet"

type ConsoleDrain struct {
	log hatchet.Logger
}

// NewConcoleDrain creates a new drain and uses a devnull logger
func NewConsoleDrain() *ConsoleDrain {
	return &ConsoleDrain{}
}

// SetLogger really allows the logtap main struct
// to assign its own logger to the concole drain
func (c *ConsoleDrain) SetLogger(l hatchet.Logger) {
	c.log = l
}

// Write formats the message given and prints it to stdout
func (c *ConsoleDrain) Write(msg Message) {
	c.log.Debug("[LOGTAP][concole][write] message:" + msg.Content)
	fmt.Printf("[%s][%s] <%d> %s", msg.Type, msg.Time, msg.Priority, msg.Content)
}
