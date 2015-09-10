// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"fmt"
	"strings"

	"github.com/pagodabox/nanobox-golang-stylish"
	"github.com/pagodabox/nanobox-server/util"
)

type ImageUpdate struct{}

//
func (j *ImageUpdate) Process() {

	//
	images, err := util.ListImages()
	if err != nil {
		util.HandleError("Unable to pull images:" + err.Error())
		util.UpdateStatus(j, "errored")
		return
	}

	//
	if len(images) == 0 {
		util.LogInfo(stylish.Bullet("No images available for update..."))
	}

	//
	for _, image := range images {
		for _, tag := range image.RepoTags {

			//
			util.LogInfo(stylish.Bullet("Updating image: %s", tag))
			if strings.HasPrefix(tag, "nanobox") {
				if err := util.UpdateImage(tag); err != nil {
					util.HandleError("Unable to update image:" + err.Error())
					util.UpdateStatus(j, "errored")
					return
				}
			}
		}
	}

	util.UpdateStatus(j, "complete")
}
