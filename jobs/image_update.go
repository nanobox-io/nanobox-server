// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"strings"

	"github.com/nanobox-io/nanobox-golang-stylish"
	"github.com/nanobox-io/nanobox-server/util"
	"github.com/nanobox-io/nanobox-server/util/docker"
)

type ImageUpdate struct{}

//
func (j *ImageUpdate) Process() {

	//
	images, err := docker.ListImages()
	if err != nil {
		util.HandleError("Unable to pull images:" + err.Error())
		util.UpdateStatus(j, "errored")
		return
	}

	//
	if len(images) == 0 {
		util.LogInfo(stylish.SubBullet("- No images available for update..."))
	}

	//
	for _, image := range images {
		for _, tag := range image.RepoTags {

			//
			if strings.HasPrefix(tag, "nanobox") {
				util.LogInfo(stylish.SubBullet("- Updating image: %s", tag))
				if err := docker.UpdateImage(tag); err != nil {
					util.HandleError("Unable to update image:" + err.Error())
					util.UpdateStatus(j, "errored")
					return
				}
			}
		}
	}

	util.LogInfo(stylish.SubBullet("- Update complete"))
	util.UpdateStatus(j, "complete")
}
