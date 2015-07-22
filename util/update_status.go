package util

import (
  "reflect"
  "strings"
  "fmt"
  "github.com/pagodabox/nanobox-server/config"
)

func UpdateStatus(v interface{}, status string) {
	t := reflect.TypeOf(v)
	value := reflect.ValueOf(v)
	id := "1"
	if value.FieldByName("ID").Kind() == reflect.String {
		id = value.FieldByName("ID").String()
	}
	config.Mist.Publish([]string{"job", strings.ToLower(t.Name())}, fmt.Sprintf(`{"model":"%s", "action":"update", "document":"{\"id\":\"%s\", \"status\":\"%s\"}"}`, t.Name(), id, status))
}