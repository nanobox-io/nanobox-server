// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pagodabox/nanobox-server/config"
)

var dirs []string = []string{"cache", "deploy", "build"}

func CreateDirs() error {
	for _, dir := range dirs {
		err := os.MkdirAll("/mnt/sda/var/nanobox/"+dir+"/", 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func Clean() error {
	for _, dir := range dirs {
		err := os.RemoveAll("/mnt/sda/var/nanobox/" + dir + "/")
		if err != nil {
			return err
		}
	}
	return CreateDirs()
}

func Touch(file string) {
	file = "/vagrant/code/" + config.App + file
	exec.Command("touch", "-c", file).Output()
}

func LibDirs() (rtn []string) {
	files, err := ioutil.ReadDir("/mnt/sda/var/nanobox/cache/lib_dirs/")
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			rtn = append(rtn, file.Name())
		}
	}
	return rtn
	
}


func libDirs() (rtn []string) {
	files, err := ioutil.ReadDir("/mnt/sda/var/nanobox/cache/lib_dirs/")
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			rtn = append(rtn, fmt.Sprintf("/mnt/sda/var/nanobox/cache/lib_dirs/%s/:/code/%s/", file.Name(), file.Name()))
		}
	}
	return rtn
}
