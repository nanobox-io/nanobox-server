package controllers

import "net/http"

type (

	//
	Controller interface {
		Index(w http.ResponseWriter, req *http.Request)
		Create(w http.ResponseWriter, req *http.Request)
		Show(w http.ResponseWriter, req *http.Request)
		Update(w http.ResponseWriter, req *http.Request)
		Delete(w http.ResponseWriter, req *http.Request)
	}
)
