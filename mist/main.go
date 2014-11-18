package mist

import(
  "fmt"
)

type(
  Adapter struct{
    Subscriptions map[string]string
  }
)

//
func (a *Adapter) Init(opts map[string]string) int {
  fmt.Println("Initializing 'Mist'")

  return 0
}

//
func (a *Adapter) Subscribe() {

}

//
func (a *Adapter) Unsubscribe() {

}

//
func (a *Adapter) forward() {

}
