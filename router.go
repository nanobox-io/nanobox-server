package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/pat"

	"github.com/nanobox-core/nanobox-server/controllers"
)

type (

	//
	Router struct {
		host, port, addr string
	}
)

//
func (r *Router) start() error {
	fmt.Println("Starting router...")

	//
	p := pat.New()

	//
	fmt.Println("Registering routes...")
	registerRoutes(p)

	fmt.Println("Listening at " + r.addr)

	//
	http.Handle("/", p)
	err := http.ListenAndServe(r.addr, nil)
	if err != nil {
		return err
	}

	return nil
}

//
func registerRoutes(p *pat.Router) {

	// evars
	evars := &controllers.EVars{}

	p.Delete("/evars/{slug}", evars.Delete)
	p.Put("/evars/{slug}", evars.Update)
	p.Get("/evars/{slug}", evars.Show)
	p.Post("/evars", evars.Create)
	p.Get("/evars", evars.Index)

	// services
	// services := &controllers.Services{}

	// p.Delete("/services/{slug}", services.Delete)
	// p.Put("/services/{slug}", services.Update)
	// p.Get("/services/{slug}", services.Show)
	// p.Post("/services", services.Create)
	// p.Get("/services", services.Index)
}
