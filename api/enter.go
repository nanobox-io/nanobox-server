// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
  "net/http"
  "strings"
  "fmt"
  "strconv"

  "github.com/pagodabox/nanobox-server/util"
)

func (api *API) Enter(rw http.ResponseWriter, req *http.Request) {
  util.RemoveContainer("enter1")
  conn, _, err := rw.(http.Hijacker).Hijack()
  if err != nil {
    rw.WriteHeader(http.StatusInternalServerError)
    rw.Write([]byte(err.Error()))
    return
  }
  defer conn.Close()

  container, err := util.CreateEnterContainer("enter1", []string{"/bin/bash"})
  if err != nil {
    conn.Write([]byte(err.Error()))
    return
  }

  // Flush the options to make sure the client sets the raw mode
  conn.Write([]byte{})

  util.AttachToContainer(container.ID, conn, conn, conn)
  util.RemoveContainer(container.ID)
}

func (api *API) KillEnter(rw http.ResponseWriter, req *http.Request) {
  fmt.Printf("signal recieved: %s\n", req.FormValue("signal"))
  err := util.KillContainer("enter1", req.FormValue("signal"))
  fmt.Println(err)
}

func (api *API) ResizeEnter(rw http.ResponseWriter, req *http.Request) {
  h, _ := strconv.Atoi(req.FormValue("h"))
  w, _ := strconv.Atoi(req.FormValue("w"))
  if h == 0 || w == 0 {
    return
  }
  err := util.ResizeContainerTTY("enter1", h, w)
  fmt.Println(err)
}

func (api *API) Run(rw http.ResponseWriter, req *http.Request) {
  util.RemoveContainer("run1")
  conn, _, err := rw.(http.Hijacker).Hijack()
  if err != nil {
    rw.WriteHeader(http.StatusInternalServerError)
    rw.Write([]byte(err.Error()))
    return
  }
  defer conn.Close()

  cmd := req.FormValue("cmd")
  container, err := util.CreateEnterContainer("enter1", strings.Split(" ", cmd))
  if err != nil {
    conn.Write([]byte(err.Error()))
    return
  }

  // Flush the options to make sure the client sets the raw mode
  conn.Write([]byte{})

  util.AttachToContainer(container.ID, conn, conn, conn)
  util.RemoveContainer(container.ID)
}

func (api *API) KillRun(rw http.ResponseWriter, req *http.Request) {
  util.KillContainer("run1", req.FormValue("signal"))
}

func (api *API) ResizeRun(rw http.ResponseWriter, req *http.Request) {
  h, _ := strconv.Atoi(req.FormValue("h"))
  w, _ := strconv.Atoi(req.FormValue("w"))
  if h == 0 || w == 0 {
    return
  }
  err := util.ResizeContainerTTY("run1", h, w)
  fmt.Println(err)
}
