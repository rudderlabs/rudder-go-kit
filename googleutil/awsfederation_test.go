package googleutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/require"
)

func TestAWSFederatedTokenSource(t *testing.T) {
	cfg := AWSFederationConfig{
		ProjectNumber: "799415897419", PoolID: "wif-pool", ProviderID: "rudderstack-aws",
		TargetServiceAccount: "rudderstack-bq@acme.iam.gserviceaccount.com", WorkspaceID: "30bK6N9S6Ca7C0SGITpgVsmRlIs",
		RoleARN: "arn:aws:iam::422074288268:role/rudderstack-gcp-federation", Region: "us-east-1",
	}

	t.Run("exchanges the AWS credentials through the customer's pool", func(t *testing.T) {
		var audience, impersonated string
		var subject struct {
			URL     string              `json:"url"`
			Headers []map[string]string `json:"headers"`
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			audience = r.Form.Get("audience")
			raw, err := url.QueryUnescape(r.Form.Get("subject_token"))
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal([]byte(raw), &subject))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"access_token": "federated", "token_type": "Bearer", "expires_in": 3600,
				"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			}))
		})
		mux.HandleFunc("/impersonate/", func(w http.ResponseWriter, r *http.Request) {
			impersonated = strings.TrimPrefix(r.URL.Path, "/impersonate/")
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"accessToken": "sa-token", "expireTime": "2030-01-01T00:00:00Z"}))
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)

		orig := awsFederationEndpoints
		awsFederationEndpoints.stsURL, awsFederationEndpoints.impersonationURL = srv.URL+"/token", srv.URL+"/impersonate/%s"
		t.Cleanup(func() { awsFederationEndpoints = orig })

		creds := credentials.NewStaticCredentialsProvider("ASIAKEY", "secret", "session-token")
		ts, err := awsFederatedTokenSource(t.Context(), cfg, []string{"scope"}, creds)
		require.NoError(t, err)
		tok, err := ts.Token()
		require.NoError(t, err)
		require.Equal(t, "sa-token", tok.AccessToken)

		require.Contains(t, subject.URL, "sts.us-east-1.amazonaws.com")
		headers := map[string]string{}
		for _, h := range subject.Headers {
			headers[strings.ToLower(h["key"])] = h["value"]
		}
		require.Contains(t, headers["authorization"], "Credential=ASIAKEY/")
		require.Equal(t, "session-token", headers["x-amz-security-token"])
		require.Equal(t, "//iam.googleapis.com/projects/799415897419/locations/global/workloadIdentityPools/wif-pool/providers/rudderstack-aws", audience)
		require.Equal(t, cfg.TargetServiceAccount, impersonated)
	})

	t.Run("reports a missing field before calling AWS", func(t *testing.T) {
		missing := cfg
		missing.WorkspaceID = ""
		_, err := AWSFederatedTokenSource(t.Context(), missing, []string{"scope"})
		require.EqualError(t, err, "workload identity federation: workspace id is required")
	})

	t.Run("without a dedicated role, requires IRSA", func(t *testing.T) {
		t.Setenv("AWS_ROLE_ARN", "")
		t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", "")
		noRole := cfg
		noRole.RoleARN = ""
		_, err := AWSFederatedTokenSource(t.Context(), noRole, []string{"scope"})
		require.EqualError(t, err, "workload identity federation: AWS role ARN is required, or IRSA (AWS_ROLE_ARN, AWS_WEB_IDENTITY_TOKEN_FILE)")
	})
}

type fakeAssumeRole struct {
	calls int
	input *sts.AssumeRoleWithWebIdentityInput
}

func (f *fakeAssumeRole) AssumeRoleWithWebIdentity(_ context.Context, in *sts.AssumeRoleWithWebIdentityInput, _ ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	f.calls++
	f.input = in
	return &sts.AssumeRoleWithWebIdentityOutput{Credentials: &ststypes.Credentials{
		AccessKeyId: aws.String("ASIAKEY"), SecretAccessKey: aws.String("secret"),
		SessionToken: aws.String("session-token"), Expiration: aws.Time(time.Now().Add(time.Hour)),
	}}, nil
}

func TestWebIdentityCredentials(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("irsa-jwt"), 0o600))
	fake := &fakeAssumeRole{}
	creds := webIdentityCredentials(fake, "arn:aws:iam::422074288268:role/data-plane-service-account", tokenFile, "30bK6N9S6Ca7C0SGITpgVsmRlIs")

	for range 3 {
		c, err := creds.Retrieve(t.Context())
		require.NoError(t, err)
		require.Equal(t, "ASIAKEY", c.AccessKeyID)
	}
	require.Equal(t, 1, fake.calls, "credentials are cached until expiry")
	require.Equal(t, "30bK6N9S6Ca7C0SGITpgVsmRlIs", aws.ToString(fake.input.RoleSessionName))
	require.Equal(t, "arn:aws:iam::422074288268:role/data-plane-service-account", aws.ToString(fake.input.RoleArn))
	require.Equal(t, "irsa-jwt", aws.ToString(fake.input.WebIdentityToken))
}
