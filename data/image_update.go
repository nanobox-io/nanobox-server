// Copyright (c) 2014 Pagoda Box Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package data

import (
	"github.com/pagodabox/nanobox-server/tasks"
)

type ImageUpdate struct {
}

func (s *ImageUpdate) Process() {
	err := tasks.PullImages()
	if err != nil {
		handleError("[UPDATE] unable to pull images", err)
		return
	}
}
