// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//
package lvs

import (
// "encoding/json"
// "fmt"
// "strings"
)

type (
	ToJson interface {
		ToJson() ([]byte, error)
	}

	FromJson interface {
		FromJson([]byte) error
	}
)
