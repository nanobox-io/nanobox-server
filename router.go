package main

import (
  "fmt"
  "net/http"

  "github.com/gorilla/pat"

  "github.com/nanobox-core/nanobox-server/controllers"
)

type(

  //
  Router struct{
    host, port, addr string
  }

)

//
func (r *Router) start() error {
  fmt.Println("Starting router... \n")

  //
  p := pat.New()

  //
  registerRoutes(p)

  //
  http.Handle("/", p)
  err := http.ListenAndServe(r.addr, nil)
  if err != nil {
    return err
  }

  fmt.Println("Listening... \n")

  return nil
}

//
func registerRoutes(p *pat.Router) {

  // evars
  evars := &controllers.EVars{}

  p.Delete("/evars/{slug}", evars.Delete)
  p.Put("/evars/{slug}", evars.Update)
  p.Get("/evars/{slug}", evars.Show)
  p.Get("/evars", evars.Index)
  p.Post("/evars", evars.Create)
}
