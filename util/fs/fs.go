// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package fs

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/nanobox-io/nanobox-server/config"
)

var dirs = []string{"cache", "deploy", "build"}

type FsUtil interface {
	CreateDirs() error
	Clean() error
	Touch(file string)
	LibDirs() (rtn []string)
	UserPayload() map[string]interface{}
}

var FsDefault FsUtil

type Fs struct {
}

func init() {
	FsDefault = Fs{}
}

func CreateDirs() error {
	return FsDefault.CreateDirs()
}
func Clean() error {
	return FsDefault.Clean()
}
func Touch(file string) {
	FsDefault.Touch(file)
}
func LibDirs() (rtn []string) {
	return FsDefault.LibDirs()
}
func UserPayload() map[string]interface{} {
	return FsDefault.UserPayload()
}

func (f Fs) CreateDirs() error {
	for _, dir := range dirs {
		err := os.MkdirAll(config.DockerMount+"sda/var/nanobox/"+dir+"/", 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f Fs) Clean() error {
	for _, dir := range dirs {
		err := os.RemoveAll(config.DockerMount + "sda/var/nanobox/" + dir + "/")
		if err != nil {
			return err
		}
	}
	return CreateDirs()
}

func (f Fs) Touch(file string) {
	file = config.MountFolder + "code/" + config.App + file
	exec.Command("touch", "-c", file).Output()
}

func (f Fs) LibDirs() (rtn []string) {
	files, err := ioutil.ReadDir(config.DockerMount + "sda/var/nanobox/cache/lib_dirs/")
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

func (f Fs) UserPayload() map[string]interface{} {
	sshFiles, err := ioutil.ReadDir(config.DockerMount + "ssh/")
	if err != nil {
		return map[string]interface{}{"ssh_files": map[string]string{}}
	}
	files := map[string]string{}
	for _, file := range sshFiles {
		if !file.IsDir() && file.Name() != "authorized_keys" && file.Name() != "config" && file.Name() != "known_hosts" {
			content, err := ioutil.ReadFile(config.DockerMount + "ssh/" + file.Name())
			if err == nil {
				files[file.Name()] = string(content)
			}
		}
	}
	return map[string]interface{}{"ssh_files": files}
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
