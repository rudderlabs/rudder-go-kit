package googleutil

import (
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/google"

	"github.com/rudderlabs/rudder-go-kit/config"
)

const (
	EMPTY_CREDS   = "{}"
	WI_CONFIG_KEY = "workloadIdentity"
)

// CompatibleGoogleCredentialsJSON returns an error if jsonKey is a Google
// developers console client_credentials.json file, which is not supported.
func CompatibleGoogleCredentialsJSON(jsonKey []byte) error {
	// google.ConfigFromJSON checks if jsonKey is a valid console client_credentials.json
	// which we won't support so "err == nil" means it is bad for us.
	if _, err := google.ConfigFromJSON(jsonKey); err == nil {
		return fmt.Errorf("google developers console client_credentials.json file is not supported")
	}
	return nil
}

// CompatibleServiceAccountJSON validates that jsonKey is a Google service account
// credentials JSON. It returns an error if the JSON is a Google developers console
// client_credentials.json file, is malformed, or has a credential type other than
// "service_account".
func CompatibleServiceAccountJSON(jsonKey []byte) error {
	if err := CompatibleGoogleCredentialsJSON(jsonKey); err != nil {
		return err
	}
	credType, err := CredentialType(jsonKey)
	if err != nil {
		return err
	}
	if credType != "service_account" {
		return fmt.Errorf("unsupported credential type %q: only service account credentials are supported", credType)
	}
	return nil
}

// federatedCredentialTypes are the credential JSON "type" values accepted by
// CompatibleFederatedCredentialsJSON, in addition to a static service account key:
// Workload Identity Federation credential configurations, as produced by
// `gcloud iam workload-identity-pools create-cred-config`. Such a configuration can
// itself carry a `service_account_impersonation_url`, letting a workload authenticate
// through an external identity provider (AWS, Azure, OIDC, SAML) and impersonate a
// GCP service account for short-lived tokens, without ever holding a static private key.
var federatedCredentialTypes = map[string]bool{
	"service_account":                  true,
	"external_account":                 true,
	"external_account_authorized_user": true,
}

// CompatibleFederatedCredentialsJSON validates that jsonKey is either a static service
// account key or a Workload Identity Federation credential configuration
// ("external_account" / "external_account_authorized_user"). It returns an error if the
// JSON is a Google developers console client_credentials.json file, is malformed, or has
// any other credential type.
//
// Important: unlike a service account key, a Workload Identity Federation configuration
// is not itself a secret, but it can reference an external `credential_source` (a file
// path or URL) that this process will read from at authentication time. Callers accepting
// such a configuration from an untrusted source must validate it first - see
// https://cloud.google.com/docs/authentication/external/externally-sourced-credentials.
func CompatibleFederatedCredentialsJSON(jsonKey []byte) error {
	if err := CompatibleGoogleCredentialsJSON(jsonKey); err != nil {
		return err
	}
	credType, err := CredentialType(jsonKey)
	if err != nil {
		return err
	}
	if !federatedCredentialTypes[credType] {
		return fmt.Errorf("unsupported credential type %q: only service account or workload identity federation credentials are supported", credType)
	}
	return nil
}

// CredentialType returns the `type` field of a Google credentials JSON (e.g.
// "service_account", "external_account"). It returns an error if jsonKey is malformed.
func CredentialType(jsonKey []byte) (string, error) {
	var f struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(jsonKey, &f); err != nil {
		return "", fmt.Errorf("invalid credentials JSON: %w", err)
	}
	return f.Type, nil
}

func ShouldSkipCredentialsInit(credentials string) bool {
	return isGKEEnabledWorkload() && isCredentialsStringEmpty(credentials)
}

/*
IsCredentialsStringEmpty checks for empty credentials.
The credentials are deemed to be empty when either the field credentials is
sent as empty string or when the field is set with "{}"

Note: This is true only for workload identity enabled rudderstack data-plane deployments
*/
func isCredentialsStringEmpty(credentials string) bool {
	return (credentials == "" || credentials == EMPTY_CREDS)
}

/*
IsGKEEnabledWorkload  checks against rudder-server configuration to find if workload identity for google destinations is enabled
*/
func isGKEEnabledWorkload() bool {
	workloadType := config.GetStringVar("", fmt.Sprintf("%s.type", WI_CONFIG_KEY))
	return workloadType == "GKE"
}
