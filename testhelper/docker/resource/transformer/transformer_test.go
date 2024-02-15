package transformer_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/transformer"
)

func TestSetup(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	t.Run("check get endpoints", func(t *testing.T) {
		tests := []struct {
			name    string
			tag     string
			urlPath string
		}{
			{
				name:    "transformer - health",
				tag:     "latest",
				urlPath: "health",
			},
			{
				name:    "user transformer - health",
				tag:     "ut-latest",
				urlPath: "health",
			},
			{
				name:    "transformer - features",
				tag:     "latest",
				urlPath: "features",
			},
			{
				name:    "user transformer - features",
				tag:     "ut-latest",
				urlPath: "features",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				transformerContainer, err := transformer.Setup(pool, t,
					transformer.WithConfigBackendURL("random-url"),
					transformer.WithDockerImageTag(tt.tag))
				require.NoError(t, err)
				endpoint, err := url.JoinPath(transformerContainer.TransformerURL, tt.urlPath)
				require.NoError(t, err)
				resp, err := http.Get(endpoint)
				require.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				require.Equal(t, resp.StatusCode, http.StatusOK)
			})
		}
	})
}
