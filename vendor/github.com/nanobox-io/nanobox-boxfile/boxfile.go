package boxfile

import (
  "launchpad.net/goyaml"
  "io/ioutil"
  "regexp"
  "strconv"
)

type Boxfile struct {
  raw []byte
  Parsed map[string]interface{}
  Valid bool
}

// NewFromPath creates a new boxfile from a file instead of raw bytes
func NewFromPath(path string) Boxfile {
  raw, _ := ioutil.ReadFile(path)
  return New(raw)
}

// New returns a boxfile object from raw data
func New(raw []byte) Boxfile {
  box := Boxfile{
    raw: raw,
    Parsed: make(map[string]interface{}),
  }
  box.parse()
  return box
}

func (self Boxfile) SaveToPath(path string) error {
  return ioutil.WriteFile(path, self.raw, 0755)
}

// Node returns just a specific node from the boxfile
// if the object is a sub hash it returns a boxfile object 
// this allows Node to be chained if you know the data
func (self Boxfile) Node(name string) (box Boxfile) {
  switch self.Parsed[name].(type) {
  case map[string]interface{}:
    box.Parsed = self.Parsed[name].(map[string]interface{})
    box.fillRaw()
    box.Valid = true
  case map[interface{}]interface{}:
    box.Parsed = make(map[string]interface{})
    for key, val := range self.Parsed[name].(map[interface{}]interface{}) {
      switch key.(type) {
      case string:
        box.Parsed[key.(string)] = val
      }
    }
    box.Valid = true
  default:
    box.Parsed = make(map[string]interface{})
    box.Valid = false
  }
  return
}

func (b Boxfile) Value(name string) interface{} {
  return b.Parsed[name]
}

func (b Boxfile) StringValue(name string) string {
  switch b.Parsed[name].(type) {
  default:
    return ""
  case string:
    return b.Parsed[name].(string)
  case bool:
    return strconv.FormatBool(b.Parsed[name].(bool))
  case int:
    return strconv.Itoa(b.Parsed[name].(int))
  case float32:
    return strconv.FormatFloat(b.Parsed[name].(float64), 'f', -1, 32)
  case float64:
    return strconv.FormatFloat(b.Parsed[name].(float64), 'f', -1, 64)
  }
}

func (b Boxfile) VersionValue(name string) string {
  switch b.Parsed[name].(type) {
  default:
    return ""
  case string:
    return b.Parsed[name].(string)
  case int:
    return strconv.Itoa(b.Parsed[name].(int)) + ".0"
  case float32:
    return strconv.FormatFloat(b.Parsed[name].(float64), 'f', 1, 32)
  case float64:
    return strconv.FormatFloat(b.Parsed[name].(float64), 'f', 1, 64)
  }
}

func (b Boxfile) IntValue(name string) int {
  switch b.Parsed[name].(type) {
  default:
    return 0
  case string:
    i, _ := strconv.Atoi(b.Parsed[name].(string))
    return i
  case bool:
    if b.Parsed[name].(bool) == true {
      return 1
    }
    return 0
  case int:
    return b.Parsed[name].(int)
  }
}

func (b Boxfile) BoolValue(name string) bool {
  switch b.Parsed[name].(type) {
  default:
    return false
  case string:
    boo, _ :=strconv.ParseBool(b.Parsed[name].(string))
    return boo
  case bool:
    return b.Parsed[name].(bool)
  case int:
    return (b.Parsed[name].(int) != 0)
  }
}

// list nodes 
// allow the user to specify which types of nodes your interested in
func (b Boxfile) Nodes(types ...string) (rtn []string) {
  if len(types) == 0 {
    for key, _ := range b.Parsed {
      rtn = append(rtn, key)
    }
    return
  }

  for _, t := range types {
    for key, _ := range b.Parsed {
      name := regexp.MustCompile(`\d+`).ReplaceAllString(key, "")
      switch t {
      case "container":
        if key != "nanobox" &&
          key != "console" &&
          key != "env" &&
          key != "build" {
          rtn = append(rtn, key)
        }
      case "service":
        if key != "nanobox" &&
          key != "console" &&
          key != "env" &&
          key != "build" &&
          name != "web" &&
          name != "worker" &&
          name != "tcp" &&
          name != "udp" {

          rtn = append(rtn, key)
        }
      case "code":
        if name == "web" || name == "worker" || name == "tcp" || name == "udp" {
          rtn = append(rtn, key)
        }
      default: 
        if name == t {
          rtn = append(rtn, key)
        }
      }
    }
  }
  return
}

// Merge puts a new boxfile data ontop of your existing boxfile
func (self *Boxfile) Merge(box Boxfile) {
  for key, val := range box.Parsed {
    switch self.Parsed[key].(type) {
    case map[string]interface{}, map[interface{}]interface{}:
      sub := self.Node(key)
      sub.Merge(box.Node(key))
      self.Parsed[key] = sub.Parsed
    default:
      self.Parsed[key] = val
    }
  }
}

// MergeProc drops a procfile into the existing boxfile
func (self *Boxfile) MergeProc(box Boxfile) {
  for key, val := range box.Parsed {
    self.Parsed[key] = map[string]interface{}{"exec":val}
  }
}

// Adds any missing storage nodes that are implied in the web => network_dirs but not 
// explicitly placed inside the root as a nfs node
func (self *Boxfile) AddStorageNode() {
  for _, node := range self.Nodes() {
    name := regexp.MustCompile(`\d+`).ReplaceAllString(node, "")
    if (name == "web" || name == "worker") && self.Node(node).Value("network_dirs") != nil {
      found := false
      for _, storage := range self.Node(node).Node("network_dirs").Nodes() {
        found = true
        if !self.Node(storage).Valid {
          self.Parsed[storage] = map[string]interface{}{"found": true}
        }
      }

      // if i dont find anything but they did have a network_dirs.. just try adding a new one
      if !found {
        if !self.Node("nfs1").Valid {
          self.Parsed["nfs1"] = map[string]interface{}{"found": true}
        }
      }
    }
  }  
}


// fillRaw is used when a boxfile is create from an existing boxfile and we want to 
// see what the raw would look like
func (b *Boxfile) fillRaw() {
  b.raw , _ = goyaml.Marshal(b.Parsed)
}

// parse takes raw data and converts it to a map structure
func (b *Boxfile) parse() {
  err := goyaml.Unmarshal(b.raw, &b.Parsed)
  if err != nil {
    b.Valid = false
  } else {
    b.Valid = true
  }
  b.ensureValid()
}

func (b *Boxfile) ensureValid() {
  if b.Valid {
    for _, node := range b.Nodes() {
      box := b.Node(node)
      if box.Valid {
        box.ensureValid()
        b.Parsed[node] = box.Parsed
      }
    }
  }
}

