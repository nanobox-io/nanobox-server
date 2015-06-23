// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package tasks

import (
	"os/exec"
	"os"
)

func Clean() error {
	err := os.MkdirAll("/var/nanobox/cache/", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll("/var/nanobox/deploy/", 0755)
	if err != nil {
		return err
	}
	if _, err := exec.Command("rm", "-rf", "/var/nanobox/deploy/*").Output(); err != nil {

		return err
	}
	return nil
}

// func Copy() error {
// 	if err := copyFolder("/vagrant/code/"+config.App+"/", "/var/nanobox/code/"); err != nil {
// 		return err
// 	} 
// 	return nil
// }



// func copyFolder(source string, dest string) (err error) {
 
// 	sourceinfo, err := os.Stat(source)
// 	if err != nil {
// 		return err
// 	}
 
// 	err = os.MkdirAll(dest, sourceinfo.Mode())
// 	if err != nil {
// 		return err
// 	}
 
// 	directory, _ := os.Open(source)
 
// 	objects, err := directory.Readdir(-1)
 
// 	for _, obj := range objects {
 
// 		sourcefilepointer := source + "/" + obj.Name()
 
// 		destinationfilepointer := dest + "/" + obj.Name()
 
// 		if obj.IsDir() {
// 			err = copyFolder(sourcefilepointer, destinationfilepointer)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			err = copyFile(sourcefilepointer, destinationfilepointer)
// 			if err != nil {
// 				return err
// 			}
// 		}
 
// 	}
// 	return
// }
 
// func copyFile(source string, dest string) (err error) {
// 	sourcefile, err := os.Open(source)
// 	if err != nil {
// 		return err
// 	}
 
// 	defer sourcefile.Close()
 
// 	destfile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
 
// 	defer destfile.Close()
 
// 	_, err = io.Copy(destfile, sourcefile)
// 	if err == nil {
// 		sourceinfo, err := os.Stat(source)
// 		if err != nil {
// 			err = os.Chmod(dest, sourceinfo.Mode())
// 		}
 
// 	}
 
// 	return
// }