package mist

import (
	"fmt"
)

type (
	Adapter struct {
		Subscriptions map[string]string
	}
)

//
func (a *Adapter) Init(opts map[string]string) int {
	fmt.Println("Initializing 'Mist'")

	return 0
}

//
func (a *Adapter) Subscribe(p string) {
	fmt.Println("SUB! ", p)
}

//
func (a *Adapter) Unsubscribe(p string) {
	fmt.Println("UNSUB! ", p)
}

//
func (a *Adapter) List(p string) {

}

//
func (a *Adapter) forward() {

}
