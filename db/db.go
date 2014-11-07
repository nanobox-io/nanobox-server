package db

import(
  "encoding/JSON"
  "fmt"
  "io/ioutil"
  "os"
)

// Write
func Write(collection, resourceID string, v interface{}) {

  //
  dir, _ := os.Stat("./db/"+collection)

  if dir == nil {
    err := os.Mkdir("./db/"+collection, 0755)
    if err != nil {
      fmt.Printf("Unable to create dir '%s': %s", collection, err)
      os.Exit(1)
    }
  }

  //
  file, err := os.Create("./db/"+collection+"/"+resourceID)
  if err != nil {
    fmt.Printf("Unable to create file %s/%s: %s", collection, resourceID, err)
    os.Exit(1)
  }

  defer file.Close()

  //
  b := toJSONIndent(v)

  _, err = file.WriteString(string(b))
  if err != nil {
    fmt.Printf("Unable to write to file %s: %s", resourceID, err)
    os.Exit(1)
  }
}

// Read
func Read(collection, resourceID string, v interface{}) interface{} {
  b, err := ioutil.ReadFile("./db/"+collection+"/"+resourceID)
  if err != nil {
    fmt.Printf("Unable to read file %s/%s: %s", collection, resourceID, err)
    os.Exit(1)
  }

  if err := fromJSON(b, v); err != nil {
    panic(err)
  }

  return v
}

// Delete
func Delete(collection, resourceID string) {
  err := os.Remove("./db/"+collection+"/"+resourceID)
  if err != nil {
    fmt.Printf("Unable to delete file %s/%s: %s", collection, resourceID, err)
    os.Exit(1)
  }
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

  //
  if err := json.Unmarshal(body, &v); err != nil {
    return err
  }

  return nil
}
