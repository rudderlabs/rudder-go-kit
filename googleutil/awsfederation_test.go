package googleutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestAWSFederatedTokenSource(t *testing.T) {
	cfg := AWSFederationConfig{
		ProjectNumber: "799415897419", PoolID: "wif-pool", ProviderID: "rudderstack-aws",
		TargetServiceAccount: "rudderstack-bq@acme.iam.gserviceaccount.com", WorkspaceID: "30bK6N9S6Ca7C0SGITpgVsmRlIs",
		RoleARN: "arn:aws:iam::422074288268:role/rudderstack-gcp-federation", Region: "us-east-1",
	}

	// fakeGoogle serves the STS token exchange and service account impersonation endpoints, recording
	// the requests they receive. The returned context routes Google API calls to it.
	type exchange struct {
		audience     string
		impersonated string
		subject      struct {
			URL     string              `json:"url"`
			Headers []map[string]string `json:"headers"`
		}
	}
	fakeGoogle := func(t *testing.T) (context.Context, *exchange) {
		var ex exchange
		mux := http.NewServeMux()
		mux.HandleFunc("sts.googleapis.com/v1/token", func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			ex.audience = r.Form.Get("audience")
			raw, err := url.QueryUnescape(r.Form.Get("subject_token"))
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal([]byte(raw), &ex.subject))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"access_token": "federated", "token_type": "Bearer", "expires_in": 3600,
				"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			}))
		})
		mux.HandleFunc("iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/{account}", func(w http.ResponseWriter, r *http.Request) {
			ex.impersonated = strings.TrimSuffix(r.PathValue("account"), ":generateAccessToken")
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"accessToken": "sa-token", "expireTime": "2030-01-01T00:00:00Z"}))
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)

		client := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r = r.Clone(r.Context())
			r.Host = r.URL.Host // keeps the Google host for routing
			r.URL.Scheme, r.URL.Host = "http", srv.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(r)
		})}
		return context.WithValue(t.Context(), oauth2.HTTPClient, client), &ex
	}
	creds := credentials.NewStaticCredentialsProvider("ASIAKEY", "secret", "session-token")

	t.Run("exchanges the AWS credentials through the customer's pool and impersonates the service account", func(t *testing.T) {
		ctx, ex := fakeGoogle(t)
		ts, err := awsFederatedTokenSource(ctx, cfg, []string{"scope"}, creds)
		require.NoError(t, err)
		tok, err := ts.Token()
		require.NoError(t, err)
		require.Equal(t, "sa-token", tok.AccessToken)

		require.Contains(t, ex.subject.URL, "sts.us-east-1.amazonaws.com")
		headers := map[string]string{}
		for _, h := range ex.subject.Headers {
			headers[strings.ToLower(h["key"])] = h["value"]
		}
		require.Contains(t, headers["authorization"], "Credential=ASIAKEY/")
		require.Equal(t, "session-token", headers["x-amz-security-token"])
		require.Equal(t, "//iam.googleapis.com/projects/799415897419/locations/global/workloadIdentityPools/wif-pool/providers/rudderstack-aws", ex.audience)
		require.Equal(t, cfg.TargetServiceAccount, ex.impersonated)
	})

	t.Run("uses the federated token directly without a target service account", func(t *testing.T) {
		ctx, ex := fakeGoogle(t)
		direct := cfg
		direct.TargetServiceAccount = ""
		ts, err := awsFederatedTokenSource(ctx, direct, []string{"scope"}, creds)
		require.NoError(t, err)
		tok, err := ts.Token()
		require.NoError(t, err)
		require.Equal(t, "federated", tok.AccessToken)
		require.Empty(t, ex.impersonated)
	})

	t.Run("reports a missing field before calling AWS", func(t *testing.T) {
		missing := cfg
		missing.WorkspaceID = ""
		_, err := AWSFederatedTokenSource(t.Context(), missing, []string{"scope"})
		require.EqualError(t, err, "workload identity federation: workspace id is required")
	})
}

func TestWebIdentityCredentials(t *testing.T) {
	provider := func(source string) webIdentityCredentials {
		return webIdentityCredentials{aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "ASIAKEY", SecretAccessKey: "secret", Source: source}, nil
		})}
	}

	t.Run("accepts IRSA web identity credentials", func(t *testing.T) {
		c, err := provider(stscreds.WebIdentityProviderName).Retrieve(t.Context())
		require.NoError(t, err)
		require.Equal(t, "ASIAKEY", c.AccessKeyID)
	})

	t.Run("rejects credentials that cannot carry the workspace session name", func(t *testing.T) {
		for _, source := range []string{"EnvConfigCredentials", "EC2RoleProvider"} {
			_, err := provider(source).Retrieve(t.Context())
			require.ErrorContains(t, err, "expected IRSA web identity credentials", source)
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
