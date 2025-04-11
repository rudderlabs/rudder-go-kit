package transformer_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/rudderlabs/rudder-go-kit/httputil"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/transformer"
)

func TestSetup(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	t.Run("test get endpoints", func(t *testing.T) {
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

	t.Run("test custom transformation", func(t *testing.T) {
		transformations := map[string]string{
			"2Nazu8t4ujUC0Dzc4pBFbjmOijx": `export function transformEvent(event, metadata) {
													event.transformed=true
													return event;
												}`,
		}
		transformerContainer, err := transformer.Setup(pool, t, transformer.WithUserTransformations(transformations, t))
		require.NoError(t, err)

		transformerURL, err := url.JoinPath(transformerContainer.TransformerURL, "customTransform")
		require.NoError(t, err)

		rawReq := []byte(`[{"message":{
				"userId": "identified_user_id",
				"anonymousId":"anonymousId_1",
				"messageId":"messageId_1",
				"type": "track",
				"event": "Product Reviewed",
				"properties": {
				  "review_id": "12345",
				  "product_id" : "123",
				  "rating" : 3.5,
				  "review_body" : "Average product, expected much more."
				}
			},"metadata":{"sourceId":"xxxyyyzzEaEurW247ad9WYZLUyk","workspaceId":"fyJvxaglFrCFxsBcLiSPBCmgpWK",
			"messageId":"messageId_1"},"destination":{"Transformations":[{"VersionID":"2Nazu8t4ujUC0Dzc4pBFbjmOijx","ID":""}]}}]`)
		req, reqErr := http.NewRequest(http.MethodPost, transformerURL, bytes.NewBuffer(rawReq))
		require.NoError(t, reqErr)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("X-Feature-Gzip-Support", "?1")
		// Header to let transformer know that the client understands event filter code
		req.Header.Set("X-Feature-Filter-Code", "?1")

		resp, err := http.DefaultClient.Do(req)
		defer func() { httputil.CloseResponse(resp) }()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		respData, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.True(t, gjson.Get(string(respData), "0.output.transformed").Bool())
	})
	t.Run("test image pull behavior based on alwaysPull flag", func(t *testing.T) {
		tests := []struct {
			name            string
			tag             string
			urlPath         string
			repository      string
			alwaysPull      bool
			expectPullError bool
		}{
			{
				name:            "alwaysPull=true: should return image pull error",
				tag:             "feat.salesforce.cache.support.search",
				urlPath:         "health",
				repository:      "rudderstack/develop-rudder-transformer",
				alwaysPull:      true,
				expectPullError: true,
			},
			{
				name:            "alwaysPull=false: should ignore image pull error",
				tag:             "feat.salesforce.cache.support.search",
				urlPath:         "health",
				repository:      "rudderstack/develop-rudder-transformer",
				alwaysPull:      false,
				expectPullError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := transformer.Setup(pool, t,
					transformer.WithRepository(tt.repository),
					transformer.WithDockerImageTag(tt.tag),
					transformer.WithAlwaysPull(tt.alwaysPull),
				)

				require.Error(t, err)

				if tt.expectPullError {
					require.Contains(t, err.Error(), "failed to pull image: ")
				} else {
					require.NotContains(t, err.Error(), "failed to pull image: ")
				}
			})
		}
	})
}
