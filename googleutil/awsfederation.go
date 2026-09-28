package googleutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
)

// AuthMethodWorkloadIdentityFederation is the destination authMethod value for authenticating
// through AWSFederatedTokenSource instead of a service account key.
const AuthMethodWorkloadIdentityFederation = "workloadIdentityFederation"

// AWSFederationConfig identifies a customer's workload identity pool, the optional service account
// RudderStack impersonates through it, and the RudderStack AWS role used to authenticate.
type AWSFederationConfig struct {
	ProjectNumber string // numeric id of the GCP project owning the pool
	PoolID        string
	ProviderID    string // AWS provider in the pool
	// TargetServiceAccount is impersonated when set. When empty, the federated token is used directly,
	// so the customer grants roles to the pool principal instead of a service account.
	TargetServiceAccount string
	// WorkspaceID is the AWS role session name, so the ARN Google sees is
	// arn:aws:sts::<account>:assumed-role/<role>/<workspaceID>. Customers scope their grant (service
	// account impersonation or direct resource access) to it, like an AWS External ID.
	WorkspaceID string
	// RoleARN is RudderStack's dedicated federation role, assumed from the workload's own identity.
	// When empty, the workload's IRSA role (AWS_ROLE_ARN) is assumed via web identity instead.
	RoleARN string
	Region  string
}

const (
	// awsSubjectType identifies the subject as a signed AWS GetCallerIdentity request.
	awsSubjectType   = "urn:ietf:params:aws:token-type:aws4_request"
	stsURL           = "https://sts.googleapis.com/v1/token"
	impersonationURL = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken"
)

// AWSFederatedTokenSource assumes an AWS role with the workspace ID as session name and exchanges those
// credentials through the customer's pool, returning tokens for cfg.TargetServiceAccount when set or the
// federated identity itself otherwise. No credential material is stored in the destination config.
func AWSFederatedTokenSource(ctx context.Context, cfg AWSFederationConfig, scopes []string) (oauth2.TokenSource, error) {
	if field := cfg.missingField(); field != "" {
		return nil, fmt.Errorf("workload identity federation: %s is required", field)
	}
	creds, err := awsFederationCredentials(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return awsFederatedTokenSource(ctx, cfg, scopes, creds)
}

// missingField names the first required field that is empty, or returns "" when all are set.
func (cfg AWSFederationConfig) missingField() string {
	switch {
	case cfg.ProjectNumber == "":
		return "project number"
	case cfg.PoolID == "":
		return "pool id"
	case cfg.ProviderID == "":
		return "provider id"
	case cfg.WorkspaceID == "":
		return "workspace id"
	case cfg.Region == "":
		return "AWS region"
	}
	return ""
}

// awsFederationCredentials returns cached AWS credentials whose session name is the workspace ID: from
// the dedicated federation role when configured, otherwise from the pod's IRSA role via the SDK default
// credential chain.
func awsFederationCredentials(ctx context.Context, cfg AWSFederationConfig) (aws.CredentialsProvider, error) {
	roleBased := cfg.RoleARN != ""
	sessionConfig := &awsutil.SessionConfig{
		Region:          cfg.Region,
		RoleBasedAuth:   roleBased,
		IAMRoleARN:      cfg.RoleARN,
		RoleSessionName: cfg.WorkspaceID,
	}
	if roleBased {
		// awsutil requires an External ID for role-based auth; the workspace ID binds the dedicated role's
		// sessions to the workspace as well, so its trust policy must not expect a fixed value.
		sessionConfig.ExternalID = cfg.WorkspaceID
	}
	awsCfg, err := awsutil.CreateAWSConfig(ctx, sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("creating AWS config for workload identity federation: %w", err)
	}
	if roleBased {
		// AssumeRole sets the workspace ID as session name explicitly, whatever the source of its base credentials.
		return awsCfg.Credentials, nil
	}
	// The default chain may fall back to static keys or the node role, which lack the workspace ID.
	return webIdentityCredentials{awsCfg.Credentials}, nil
}

// webIdentityCredentials only accepts web identity (IRSA) credentials from the default chain. Any other
// source - static keys, instance metadata - would not carry the workspace ID as session name, so the
// customer's workspace-scoped grant must not be satisfied by it.
type webIdentityCredentials struct {
	aws.CredentialsProvider
}

func (p webIdentityCredentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	c, err := p.CredentialsProvider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}
	if c.Source != stscreds.WebIdentityProviderName {
		return aws.Credentials{}, fmt.Errorf(
			"workload identity federation: expected IRSA web identity credentials, got %q: set a federation role ARN or AWS_ROLE_ARN and AWS_WEB_IDENTITY_TOKEN_FILE", c.Source,
		)
	}
	return c, nil
}

func awsFederatedTokenSource(ctx context.Context, cfg AWSFederationConfig, scopes []string, creds aws.CredentialsProvider) (oauth2.TokenSource, error) {
	conf := externalaccount.Config{
		Audience: fmt.Sprintf("//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s",
			cfg.ProjectNumber, cfg.PoolID, cfg.ProviderID),
		SubjectTokenType:               awsSubjectType,
		TokenURL:                       stsURL,
		Scopes:                         scopes,
		AwsSecurityCredentialsSupplier: &awsCredentialsSupplier{region: cfg.Region, creds: creds},
	}
	if cfg.TargetServiceAccount != "" {
		conf.ServiceAccountImpersonationURL = fmt.Sprintf(impersonationURL, cfg.TargetServiceAccount)
	}
	// the token source refreshes long after this call, so it must not hold a request-scoped context
	ts, err := externalaccount.NewTokenSource(context.WithoutCancel(ctx), conf)
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
