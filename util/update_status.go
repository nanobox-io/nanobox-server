package util

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/nanobox-io/nanobox-server/config"
)

// due to the way this uses the reflect library there are certain assumptions made
// about the interface the function recieves:
// 1. It is of type Struct{}
// 2. Its Kind is one of: Array, Chan, Map, Ptr, or Slice
// 3. The ID (if any) will always be of type String
func UpdateStatus(v interface{}, status string) {

	name := reflect.TypeOf(v).Elem().Name()
	id := reflect.ValueOf(v).Elem().FieldByName("ID").String()

	if id == "" {
		id = "1"
	}

	// allow any messages that were waiting to be sent before me
	runtime.Gosched()
	config.Mist.Publish([]string{"job", strings.ToLower(name)}, fmt.Sprintf(`{"model":"%s", "action":"update", "document":{"id":"%s", "status":"%s"}}`, name, id, status))
}
