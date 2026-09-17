package awsutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/stretchr/testify/require"
)

func TestGetRegionFromBucket(t *testing.T) {
	envBucket := os.Getenv("AWS_BUCKET_NAME")
	region, err := GetRegionFromBucket(context.Background(), envBucket, "us-east-1")
	require.NoError(t, err)
	require.Equal(t, "us-east-1", region)
}

func TestRoleSessionName(t *testing.T) {
	require.Equal(t, "rudderstack-aws-s3-access", roleSessionName(&SessionConfig{Service: "S3"}))
	require.Equal(t, "30bK6N9S6Ca7C0SGITpgVsmRlIs", roleSessionName(&SessionConfig{Service: "S3", RoleSessionName: "30bK6N9S6Ca7C0SGITpgVsmRlIs"}))
}

// Without a role or static keys the default chain resolves IRSA web identity from the environment, and
// RoleSessionName must become its session name.
func TestCreateAWSConfigWebIdentitySessionName(t *testing.T) {
	var gotSessionName, gotRoleARN string
	sts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		gotSessionName, gotRoleARN = r.Form.Get("RoleSessionName"), r.Form.Get("RoleArn")
		w.Header().Set("Content-Type", "text/xml")
		_, err := w.Write([]byte(`<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<AssumeRoleWithWebIdentityResult><Credentials><AccessKeyId>ASIAKEY</AccessKeyId><SecretAccessKey>secret</SecretAccessKey>
<SessionToken>token</SessionToken><Expiration>2099-01-01T00:00:00Z</Expiration></Credentials></AssumeRoleWithWebIdentityResult>
</AssumeRoleWithWebIdentityResponse>`))
		require.NoError(t, err)
	}))
	t.Cleanup(sts.Close)

	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("irsa-jwt"), 0o600))
	for k, v := range map[string]string{
		"AWS_ROLE_ARN":                "arn:aws:iam::422074288268:role/data-plane-service-account",
		"AWS_WEB_IDENTITY_TOKEN_FILE": tokenFile,
		"AWS_ENDPOINT_URL_STS":        sts.URL,
		"AWS_ROLE_SESSION_NAME":       "from-env", // must be overridden
		"AWS_ACCESS_KEY_ID":           "",
		"AWS_SECRET_ACCESS_KEY":       "",
		"AWS_CONFIG_FILE":             filepath.Join(t.TempDir(), "none"),
		"AWS_SHARED_CREDENTIALS_FILE": filepath.Join(t.TempDir(), "none"),
	} {
		t.Setenv(k, v)
	}

	cfg, err := CreateAWSConfig(t.Context(), &SessionConfig{Region: "us-east-1", RoleSessionName: "30bK6N9S6Ca7C0SGITpgVsmRlIs"})
	require.NoError(t, err)
	creds, err := cfg.Credentials.Retrieve(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ASIAKEY", creds.AccessKeyID)
	require.Equal(t, stscreds.WebIdentityProviderName, creds.Source)
	require.Equal(t, "30bK6N9S6Ca7C0SGITpgVsmRlIs", gotSessionName)
	require.Equal(t, "arn:aws:iam::422074288268:role/data-plane-service-account", gotRoleARN)
}
