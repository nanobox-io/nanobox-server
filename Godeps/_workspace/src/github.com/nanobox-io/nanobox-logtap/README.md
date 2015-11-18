# Logtap

Logtap is an embeddable and configurable log aggregation, storage, and publishing service.

## Memory Usage

## logtap.Drain

A `logtap.Drain` is a simple endpoint that accepts logs that are sent through logtap. Multiple drains can be created and added to logvac. A drain can represent logs that are streamed to stdout, a file, a tcp socket, or anything that can be wrapped to accept `logtap.Message` structs. There are 3 adapters stored at `logtap/drain.Adapt*` that can be used to adapt common interfaces to the logvac.Drain interface.

## logtap.Archive

An Archive is an interface for retreiving a slice of logs from the opaque storage medium. Currently there exists one storage option: BoltDB.

## Example

This example will create a logtap that accepts udp syslog packets and stores them on disk in two files, `fatal.log` and `info.log`. The file `fatal.log` will only contain fatal errors and higher, while `info.log` will contain all message that are info and more severe.

```go

package main

import (
  "github.com/nanobox-io/nanobox-logtap"
  "github.com/nanobox-io/nanobox-logtap/drain"
  "github.com/nanobox-io/nanobox-logtap/collector"
  "os"
  "os/signal"
)

func main(){
  logTap := logtap.New(nil)
  defer logTap.Close()
  
  fatal, err := os.Create("/tmp/fatal.log")
  if err != nil {
    panic(err)
  }
  defer fatal.Close()
  info, err := os.Create("/tmp/info.log")
  if err != nil {
    panic(err)
  }
  defer info.Close()

  logTap.AddDrain("info", drain.Filter(drain.AdaptWriter(info)), 1)
  logTap.AddDrain("fatal", drain.Filter(drain.AdaptWriter(fatal)), 5)

  udpCollector, err := collector.SyslogUDPStart("app-logs", ":514" ,logTap)
  if err != nil {
    panic(err)
  }
  defer udpCollector.Close()

  logTap.Publish("logtap", 1, "listening on udp port 514")

  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, os.Kill)

  // wait for a signal to arrive
  <-c
}
```


### Notes: