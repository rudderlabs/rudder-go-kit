package rudderauth_test

import (
	"encoding/json"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/jsonparser"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/rudderauth"
)

func TestRudderAuth(t *testing.T) {
	runScenario := func(t *testing.T, accountDefinition bool) {
		pool, err := dockertest.NewPool("")
		require.NoError(t, err)
		fixture := rudderauth.AccountFixture{
			ID:                   "account-id",
			Type:                 "test",
			Category:             "destination",
			WorkspaceID:          "workspace-id",
			DestinationID:        "destination-id",
			HasAccountDefinition: accountDefinition,
			Secret: json.RawMessage(`{
			"access_token": "abcdefghijklmnop",
			"refresh_token": "qrstuvwxyz"
		}`),
			AuthClientSecrets: map[string]string{
				"CLIENT_ID":     "client-id",
				"CLIENT_SECRET": "client-secret",
			},
		}

		rudderAuthResource, err := rudderauth.Setup(pool, fixture, t)
		require.NoError(t, err, "it should be able to setup rudder auth resource")

		token, err := rudderAuthResource.FetchToken()
		require.NoError(t, err, "it should be able to fetch the token")
		require.NotEmpty(t, token, "token should not be empty")
		require.NotEmpty(t, jsonparser.GetValueOrEmpty(token, "secret"), "token.secret should not be empty")

		err = rudderAuthResource.ToggleStatus("inactive")
		require.NoError(t, err, "it should be able to toggle the status to inactive")
		err = rudderAuthResource.ToggleStatus("active")
		require.NoError(t, err, "it should be able to toggle the status back to active")
	}

	t.Parallel()
	t.Run("without account definition", func(t *testing.T) {
		runScenario(t, false)
	})
	t.Run("with account definition", func(t *testing.T) {
		runScenario(t, true)
	})
}
