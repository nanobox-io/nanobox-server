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