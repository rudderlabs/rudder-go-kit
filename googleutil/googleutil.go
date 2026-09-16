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
	var f struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(jsonKey, &f); err != nil {
		return fmt.Errorf("invalid credentials JSON: %w", err)
	}
	if f.Type != "service_account" {
		return fmt.Errorf("unsupported credential type %q: only service account credentials are supported", f.Type)
	}
	return nil
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
