package googleutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
)

// AuthMethodWorkloadIdentityFederation is the destination authMethod value for authenticating
// through AWSFederatedTokenSource instead of a service account key.
const AuthMethodWorkloadIdentityFederation = "workloadIdentityFederation"

// AWSFederationConfig identifies a customer's workload identity pool, the service account
// RudderStack impersonates through it, and the RudderStack AWS role used to authenticate.
type AWSFederationConfig struct {
	ProjectNumber        string // numeric id of the GCP project owning the pool
	PoolID               string
	ProviderID           string // AWS provider in the pool
	TargetServiceAccount string
	// WorkspaceID is the AWS role session name, so the ARN Google sees is
	// arn:aws:sts::<account>:assumed-role/<role>/<workspaceID>. Customers bind their service
	// account to it, scoping the grant to one workspace like an AWS External ID.
	WorkspaceID string
	RoleARN     string // RudderStack's federation role, assumed from the workload's own identity
	Region      string
}

// awsSubjectType identifies the subject as a signed AWS GetCallerIdentity request.
const awsSubjectType = "urn:ietf:params:aws:token-type:aws4_request"

var awsFederationEndpoints = struct{ stsURL, impersonationURL string }{
	stsURL:           "https://sts.googleapis.com/v1/token",
	impersonationURL: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken",
}

// AWSFederatedTokenSource returns tokens for cfg.TargetServiceAccount by assuming cfg.RoleARN with
// the workspace ID as session name and exchanging those credentials through the customer's pool.
// No credential material is stored in the destination config.
func AWSFederatedTokenSource(ctx context.Context, cfg AWSFederationConfig, scopes []string) (oauth2.TokenSource, error) {
	for _, f := range [][2]string{
		{"project number", cfg.ProjectNumber},
		{"pool id", cfg.PoolID},
		{"provider id", cfg.ProviderID},
		{"target service account", cfg.TargetServiceAccount},
		{"workspace id", cfg.WorkspaceID},
		{"AWS role ARN", cfg.RoleARN},
		{"AWS region", cfg.Region},
	} {
		if f[1] == "" {
			return nil, fmt.Errorf("workload identity federation: %s is required", f[0])
		}
	}
	awsCfg, err := awsutil.CreateAWSConfig(ctx, &awsutil.SessionConfig{
		Region:          cfg.Region,
		RoleBasedAuth:   true,
		IAMRoleARN:      cfg.RoleARN,
		ExternalID:      cfg.WorkspaceID,
		RoleSessionName: cfg.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating AWS config for workload identity federation: %w", err)
	}
	return awsFederatedTokenSource(ctx, cfg, scopes, awsCfg.Credentials) // already cached by the AWS config
}

func awsFederatedTokenSource(ctx context.Context, cfg AWSFederationConfig, scopes []string, creds aws.CredentialsProvider) (oauth2.TokenSource, error) {
	ts, err := externalaccount.NewTokenSource(ctx, externalaccount.Config{
		Audience: fmt.Sprintf("//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s",
			cfg.ProjectNumber, cfg.PoolID, cfg.ProviderID),
		SubjectTokenType:               awsSubjectType,
		TokenURL:                       awsFederationEndpoints.stsURL,
		ServiceAccountImpersonationURL: fmt.Sprintf(awsFederationEndpoints.impersonationURL, cfg.TargetServiceAccount),
		Scopes:                         scopes,
		AwsSecurityCredentialsSupplier: &awsCredentialsSupplier{region: cfg.Region, creds: creds},
	})
	if err != nil {
		return nil, fmt.Errorf("building workload identity federation credentials: %w", err)
	}
	return ts, nil
}

type awsCredentialsSupplier struct {
	region string
	creds  aws.CredentialsProvider
}

func (s *awsCredentialsSupplier) AwsRegion(context.Context, externalaccount.SupplierOptions) (string, error) {
	return s.region, nil
}

func (s *awsCredentialsSupplier) AwsSecurityCredentials(ctx context.Context, _ externalaccount.SupplierOptions) (*externalaccount.AwsSecurityCredentials, error) {
	c, err := s.creds.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("assuming AWS role: %w", err)
	}
	return &externalaccount.AwsSecurityCredentials{AccessKeyID: c.AccessKeyID, SecretAccessKey: c.SecretAccessKey, SessionToken: c.SessionToken}, nil
}
