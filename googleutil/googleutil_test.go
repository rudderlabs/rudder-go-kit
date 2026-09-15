package googleutil_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/googleutil"
)

func TestCompatibleGoogleCredentialsJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonKey     string
		wantErrFrag string
	}{
		{
			name:    "service account is accepted",
			jsonKey: `{"type": "service_account", "project_id": "my-project"}`,
		},
		{
			name:    "authorized_user is accepted",
			jsonKey: `{"type": "authorized_user"}`,
		},
		{
			name:    "external_account is accepted",
			jsonKey: `{"type": "external_account"}`,
		},
		{
			name:        "console client credentials are rejected",
			jsonKey:     `{"installed": {"client_id": "x", "client_secret": "y", "redirect_uris": ["http://localhost"]}}`,
			wantErrFrag: "google developers console client_credentials.json file is not supported",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := googleutil.CompatibleGoogleCredentialsJSON([]byte(tc.jsonKey))
			if tc.wantErrFrag != "" {
				require.ErrorContains(t, err, tc.wantErrFrag)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompatibleServiceAccountJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonKey     string
		wantErrFrag string
	}{
		{
			name:    "valid service account",
			jsonKey: `{"type": "service_account", "project_id": "my-project"}`,
		},
		{
			name:        "authorized_user is rejected",
			jsonKey:     `{"type": "authorized_user"}`,
			wantErrFrag: `unsupported credential type "authorized_user"`,
		},
		{
			name:        "external_account is rejected",
			jsonKey:     `{"type": "external_account"}`,
			wantErrFrag: `unsupported credential type "external_account"`,
		},
		{
			name:        "impersonated_service_account is rejected",
			jsonKey:     `{"type": "impersonated_service_account"}`,
			wantErrFrag: `unsupported credential type "impersonated_service_account"`,
		},
		{
			name:        "console client credentials are rejected",
			jsonKey:     `{"installed": {"client_id": "x", "client_secret": "y", "redirect_uris": ["http://localhost"]}}`,
			wantErrFrag: "google developers console client_credentials.json file is not supported",
		},
		{
			name:        "missing type field is rejected",
			jsonKey:     `{"project_id": "my-project"}`,
			wantErrFrag: `unsupported credential type ""`,
		},
		{
			name:        "malformed JSON is rejected",
			jsonKey:     `not-json`,
			wantErrFrag: "invalid credentials JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := googleutil.CompatibleServiceAccountJSON([]byte(tc.jsonKey))
			if tc.wantErrFrag != "" {
				require.ErrorContains(t, err, tc.wantErrFrag)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompatibleFederatedCredentialsJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonKey     string
		wantErrFrag string
	}{
		{
			name:    "service account is accepted",
			jsonKey: `{"type": "service_account", "project_id": "my-project"}`,
		},
		{
			name:    "external_account is accepted",
			jsonKey: `{"type": "external_account", "audience": "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider"}`,
		},
		{
			name:    "external_account with service account impersonation is accepted",
			jsonKey: `{"type": "external_account", "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/sa@my-project.iam.gserviceaccount.com:generateAccessToken"}`,
		},
		{
			name:    "external_account_authorized_user is accepted",
			jsonKey: `{"type": "external_account_authorized_user"}`,
		},
		{
			name:        "authorized_user is rejected",
			jsonKey:     `{"type": "authorized_user"}`,
			wantErrFrag: `unsupported credential type "authorized_user"`,
		},
		{
			name:        "impersonated_service_account is rejected",
			jsonKey:     `{"type": "impersonated_service_account"}`,
			wantErrFrag: `unsupported credential type "impersonated_service_account"`,
		},
		{
			name:        "console client credentials are rejected",
			jsonKey:     `{"installed": {"client_id": "x", "client_secret": "y", "redirect_uris": ["http://localhost"]}}`,
			wantErrFrag: "google developers console client_credentials.json file is not supported",
		},
		{
			name:        "missing type field is rejected",
			jsonKey:     `{"project_id": "my-project"}`,
			wantErrFrag: `unsupported credential type ""`,
		},
		{
			name:        "malformed JSON is rejected",
			jsonKey:     `not-json`,
			wantErrFrag: "invalid credentials JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := googleutil.CompatibleFederatedCredentialsJSON([]byte(tc.jsonKey))
			if tc.wantErrFrag != "" {
				require.ErrorContains(t, err, tc.wantErrFrag)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCredentialType(t *testing.T) {
	t.Run("returns the type field", func(t *testing.T) {
		credType, err := googleutil.CredentialType([]byte(`{"type": "external_account"}`))
		require.NoError(t, err)
		require.Equal(t, "external_account", credType)
	})

	t.Run("returns empty string when type is absent", func(t *testing.T) {
		credType, err := googleutil.CredentialType([]byte(`{"project_id": "my-project"}`))
		require.NoError(t, err)
		require.Equal(t, "", credType)
	})

	t.Run("errors on malformed JSON", func(t *testing.T) {
		_, err := googleutil.CredentialType([]byte(`not-json`))
		require.ErrorContains(t, err, "invalid credentials JSON")
	})
}
