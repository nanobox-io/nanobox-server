package db

import (
  "encoding/JSON"
  "fmt"
  "io/ioutil"
  "os"
)

//
type (

  // Driver represents
  Driver struct {
    dir string
    channels map[string]chan int
  }

  // Transaction represents
  Transaction struct {
    Action string
    Collection string
    Resource string
    Container interface{}
  }
)

// Init
func (d *Driver) Init(opts map[string]string) int {
  d.dir = opts["db_dir"]

  d.channels = make(map[string]chan int)

  // make a ping channel
  ping := make(chan int, 1)
  d.channels["ping"] = ping

  //
  if err := mkDir(d.dir); err != nil {
    fmt.Printf("Unable to create dir '%s': %s", d.dir, err)
    return 1
  }

  //
  return 0
}

// Transact
func (d *Driver) Transact(trans Transaction) {

  //
  c := d.getChan(trans.Collection)

  //
  switch trans.Action {
  case "write":
    go d.write(trans, c)
  case "read":
    go d.read(trans, c)
  case "readall":
    go d.readAll(trans, c)
  case "delete":
    go d.delete(trans, c)
  default:
    fmt.Println("Unsupported action ", trans.Action)
  }

  // wait...
  <-c
}

// getChan
func (d *Driver) getChan(channel string) chan int {

  c, ok := d.channels[channel]

  // if the chan doesn't exist make it
  if !ok {
    d.channels[channel] = make(chan int, 1)
    return d.channels[channel]
  }

  return c
}


// private

// write
func (d *Driver) write(trans Transaction, c chan<- int) {

  //
  dir := d.dir+"/"+trans.Collection

  if err := mkDir(dir); err != nil {
    fmt.Println("Unable to create dir '%s': %s", dir, err)
    os.Exit(1)
  }

  //
  file, err := os.Create(dir + "/" + trans.Resource)
  if err != nil {
    fmt.Printf("Unable to create file %s/%s: %s", trans.Collection, trans.Resource, err)
    os.Exit(1)
  }

  defer file.Close()

  //
  b := toJSONIndent(trans.Container)

  _, err = file.WriteString(string(b))
  if err != nil {
    fmt.Printf("Unable to write to file %s: %s", trans.Resource, err)
    os.Exit(1)
  }

  c <- 0
}

// read
func (d *Driver) read(trans Transaction, c chan<- int) interface{} {

  dir := d.dir+"/"+trans.Collection

  b, err := ioutil.ReadFile(dir + "/" + trans.Resource)
  if err != nil {
    fmt.Printf("Unable to read file %s/%s: %s", trans.Collection, trans.Resource, err)
    os.Exit(1)
  }

  if err := fromJSON(b, trans.Container); err != nil {
    panic(err)
  }

  c <- 0

  return trans.Container
}

// readAll
func (d *Driver) readAll(trans Transaction, c chan<- int) {

  dir := d.dir+"/"+trans.Collection

  //
  files, err := ioutil.ReadDir(dir)

  // if there is an error here it just means there are no evars so dont do anything
  if err != nil { }

  var f []string

  for _, file := range files {
    b, err := ioutil.ReadFile(dir + "/" + file.Name())
    if err != nil {
      panic(err)
    }

    f = append(f, string(b))

  }

  b := toJSON(f)

  if err := json.Unmarshal(b, &trans.Container); err != nil {
    panic(err)
  }

  c <- 0
}

// delete
func (d *Driver) delete(trans Transaction, c chan<- int) {

  dir := d.dir+"/"+trans.Collection

  err := os.Remove(dir + "/" + trans.Resource)
  if err != nil {
    fmt.Printf("Unable to delete file %s/%s: %s", trans.Collection, trans.Resource, err)
    os.Exit(1)
  }

  c <- 0
}

// helpers

// mkDir
func mkDir(d string) error {

  //
  dir, _ := os.Stat(d)

  if dir == nil {
    err := os.MkdirAll(d, 0755)
    if err != nil {
      return err
    }
  }

  return nil
}

// toJSON converts an interface (v) into JSON bytecode
func toJSON(v interface{}) []byte {
  b, err := json.Marshal(v)
  if err != nil {
    panic(err)
  }

  return b
}

// toJSONIndent
func toJSONIndent(v interface{}) []byte {
  b, err := json.MarshalIndent(v, "", "\t")
  if err != nil {
    panic(err)
  }

  return b
}

// fromJSON converts an interface (v) into JSON bytecode
func fromJSON(body []byte, v interface{}) error {
  if err := json.Unmarshal(body, &v); err != nil {
    return err
  }

  return nil
}
