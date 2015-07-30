package util

import (
	"fmt"
	"github.com/pagodabox/nanobox-server/config"
	"reflect"
	"runtime"
	"strings"
)

func UpdateStatus(v interface{}, status string) {
	t := reflect.TypeOf(v).Elem()
	value := reflect.ValueOf(v).Elem()
	id := "1"
	if value.FieldByName("ID").Kind() == reflect.String {
		id = value.FieldByName("ID").String()
	}
	// allow any messages that were waiting to be sent before me
	runtime.Gosched()
	config.Mist.Publish([]string{"job", strings.ToLower(t.Name())}, fmt.Sprintf(`{"model":"%s", "action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, t.Name(), id, status))
}
