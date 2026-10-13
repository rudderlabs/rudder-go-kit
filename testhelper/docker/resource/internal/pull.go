package internal

import (
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

// PullImage pulls repository:tag unless it is already present. A tag of the form "tag@sha256:..." is pulled by digest,
// because dockertest sends the whole value as the tag and the Docker API rejects it.
func PullImage(pool *dockertest.Pool, repository, tag string) error {
	if _, err := pool.Client.InspectImage(repository + ":" + tag); err == nil {
		return nil
	}
	if _, digest, pinned := strings.Cut(tag, "@"); pinned {
		return pool.Client.PullImage(docker.PullImageOptions{Repository: repository + "@" + digest}, registry.AuthConfiguration())
	}
	return pool.Client.PullImage(docker.PullImageOptions{Repository: repository, Tag: tag}, registry.AuthConfiguration())
}
