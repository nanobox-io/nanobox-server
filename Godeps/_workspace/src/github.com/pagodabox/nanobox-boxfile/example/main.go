package main

import "github.com/pagodabox/nanobox-boxfile"
import "fmt"

func main() {
  box := boxfile.New([]byte("a: Easy!\nb:\n  c: 2\n  d: [3, 4]\n"))
  fmt.Println(box)
  if !box.Valid { fmt.Println("why isnt it true") }

  box2 := boxfile.NewFromPath("Boxfile")
  fmt.Println(box2)
  fmt.Println(box2.Node("web1").Node("php_extensions"))
  box2.Merge(box)
  fmt.Printf("%#v\n", box2)
}