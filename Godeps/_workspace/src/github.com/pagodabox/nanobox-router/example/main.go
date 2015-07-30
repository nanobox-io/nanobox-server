package main

import "github.com/pagodabox/nanobox-router"
import "github.com/jcelliott/lumber"
import "time"
import "net"

func main() {
	log := lumber.NewConsoleLogger(lumber.INFO)

	r := router.New("80", log)
	r.AddTarget("/", "http://drawception.com")
	r.AddTarget("/category/", "http://macmagazine.com.br")

	log.Info("adding tcpforward to google.com:80")
	port, err := r.AddForward("what", 1234, "192.168.13.164:22")
	if err != nil {
		log.Error("%#v\n", err.Error())
		errer := err.(*net.OpError)
		if errer.Err.Error() == "bind: address already in use" {
			log.Info("HAY!!!")
		}
	}
	port, err = r.AddForward("what", 1234, "192.168.13.164:22")
	if err != nil {
		log.Error("%#v\n", err.Error())
		errer := err.(*net.OpError)
		if errer.Err.Error() == "bind: address already in use" {
			log.Info("HAY!!!")
		}
	}
	log.Info("%d\n", port)

	time.Sleep(100 * time.Second)
	r.Handler = router.NoDeploy{}
	time.Sleep(20 * time.Second)
	r.Handler = nil
	log.Info("port is still : ", r.GetLocalPort("192.168.13.164:22"))
	r.RemoveForward("what")
	time.Sleep(100 * time.Second)
}
