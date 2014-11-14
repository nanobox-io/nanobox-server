package utils

import (
	"encoding/JSON"
	"io/ioutil"
	"net/http"
)

// WriteResponse
func WriteResponse(v interface{}, w http.ResponseWriter, status int) {
	b := ToJSON(v)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

// ParseBody
func ParseBody(r *http.Request, v interface{}) error {

	//
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	//
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}
