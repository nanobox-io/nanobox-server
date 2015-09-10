package util

import (
	"fmt"
	"strconv"

	"github.com/pagodabox/nanobox-lvs"
	"github.com/pagodabox/nanobox-server/config"
)

// make sure the router is being forwarded
func init() {
	AddForward("80", config.IP, config.Ports["router"])
}

// add a server into the lvs system
func AddForward(fromPort, toIp, toPort string) error {
	fromInt, err := strconv.Atoi(fromPort)
	if err != nil {
		return err
	}
	_, err = lvs.AddVip(config.IP, fromInt)
	if err != nil {
		config.Log.Error(fmt.Sprintf("error on: lvs.AddVip(\"%s\", %d)\n", config.IP, fromInt))
		return err
	}
	toInt, _ := strconv.Atoi(toPort)
	_, err = lvs.AddServer(fmt.Sprintf("%s:%d", config.IP, fromInt), toIp, toInt)
	if err != nil {
		config.Log.Error(fmt.Sprintf("error on: lvs.AddServer(\"%s:%d\", \"%s\", %d)\n", config.IP, fromInt, config.IP, toInt))
		return err
	}
	return nil
}

func RemoveForward(ip string) error {
	vips, err := lvs.ListVips()
	if err != nil {
		return err
	}

	errorString := ""

	for _, vip := range vips {
		for _, server := range vip.Servers {
			if server.Host == ip {
				err := lvs.DeleteVip(fmt.Sprintf("%s:%d", vip.Host, vip.Port))
				if err != nil {
					errorString = fmt.Sprintf("%s%v\n", errorString, err.Error())
				}
				break
			}
		}
	}

	if errorString != "" {
		return fmt.Errorf(errorString)
	}
	return nil
}
