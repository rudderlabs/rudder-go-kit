package transformer_test

import (
	"bytes"
	"io"
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

	transformerContainer, err := transformer.Setup(pool, t, transformer.WithConfigBackendURL("https://api.dev.rudderlabs.com"))
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
	if reqErr != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Feature-Gzip-Support", "?1")
	// Header to let transformer know that the client understands event filter code
	req.Header.Set("X-Feature-Filter-Code", "?1")

	client := http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, http.StatusOK)
	respData, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(respData), "\"transformed\":true")
}
