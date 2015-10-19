package docker

import (
	dc "github.com/fsouza/go-dockerclient"
)

func (d DockerUtil) InstallImage(image string) error {
	return Client.PullImage(dc.PullImageOptions{Repository: image}, dc.AuthConfiguration{})
}

func (d DockerUtil) ListImages() ([]dc.APIImages, error) {
	return Client.ListImages(dc.ListImagesOptions{})
}

func (d DockerUtil) ImageExists(name string) bool {
	images, err := ListImages()
	if err != nil {
		return false
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == name+":latest" || tag == name {
				return true
			}
		}
	}

	return false
}
