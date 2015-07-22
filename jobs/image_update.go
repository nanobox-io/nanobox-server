// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package jobs

import (
	"github.com/pagodabox/nanobox-server/util"
)

type ImageUpdate struct {
}

func (j *ImageUpdate) Process() {
	err := util.UpdateAllImages()
	if err != nil {
		util.HandleError("Unable to pull images", err.Error())
		return
	}
}
