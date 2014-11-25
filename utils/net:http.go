package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// WriteResponse
func WriteResponse(v interface{}, rw http.ResponseWriter, status int) {
	b := ToJSON(v)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(b)

	//
	rw.(http.Flusher).Flush()
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
