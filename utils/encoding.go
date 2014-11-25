package utils

import (
	"encoding/json"
)

// ToJSON converts an interface (v) into JSON bytecode
func ToJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

// ToJSONIndent converts an interface (v) into formatted JSON bytecode
func ToJSONIndent(v interface{}) []byte {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}

	return b
}

// FromJSON converts JSON into the provided interface (v)
func FromJSON(body []byte, v interface{}) error {
	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}

	return nil
}
