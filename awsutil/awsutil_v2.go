package awsutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// CreateAWSConfig creates an AWS config using the provided SessionConfig.
// It supports both static credentials and role-based authentication.
func CreateAWSConfig(ctx context.Context, config *SessionConfig) (aws.Config, error) {
	var (
		zero           aws.Config
		err            error
		awsCredentials aws.CredentialsProvider
	)

	if config == nil {
		return zero, errors.New("awsutil: SessionConfig cannot be nil")
	}

	// Handle credentials - either role-based or static
	if config.RoleBasedAuth {
		awsCredentials, err = createV2CredentialsForRole(ctx, config)
		if err != nil {
			return zero, fmt.Errorf("awsutil: failed to create credentials for role: %w", err)
		}
	} else if config.AccessKeyID != "" && config.AccessKey != "" {
		awsCredentials = credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKey, config.SessionToken)
	}

	// Configure HTTP client with timeout and connection settings
	httpClient := configureHTTPClient(config)

	// Load default config with options
	optFuncs := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(config.Region),
		awsconfig.WithHTTPClient(httpClient),
		awsconfig.WithCredentialsProvider(awsCredentials),
	}

	// Add shared config profile if specified
	if config.SharedConfigProfile != "" {
		optFuncs = append(optFuncs, awsconfig.WithSharedConfigProfile(config.SharedConfigProfile))
	}

	// Load the AWS configuration
	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFuncs...)
	if err != nil {
		return zero, fmt.Errorf("awsutil: failed to load AWS config: %w", err)
	}

	return cfg, nil
}

// configureHTTPClient creates an HTTP client with the configuration settings
func configureHTTPClient(config *SessionConfig) *http.Client {
	// Create transport with proper connection settings
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
	}

	// Adjust MaxIdleConns if necessary
	if config.MaxIdleConnsPerHost > 0 {
		if transport.MaxIdleConns < config.MaxIdleConnsPerHost {
			transport.MaxIdleConns = config.MaxIdleConnsPerHost
		}
	}

	// Create client with transport and timeout
	client := &http.Client{
		Transport: transport,
	}

	// Set timeout if provided
	if config.Timeout != nil {
		client.Timeout = *config.Timeout
	}

	return client
}

// createV2CredentialsForRole creates a credentials provider for assuming an IAM role.
// This is the v2 equivalent of createCredentialsForRole in v1.
func createV2CredentialsForRole(ctx context.Context, config *SessionConfig) (aws.CredentialsProvider, error) {
	if config.ExternalID == "" {
		return nil, errors.New("awsutil: externalID is required for IAM role")
	}

	// Configure HTTP client for the base session
	httpClient := configureHTTPClient(config)

	// Create base config for STS operations
	optFuncs := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(config.Region),
		awsconfig.WithHTTPClient(httpClient),
	}

	// Load configuration for STS client
	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFuncs...)
	if err != nil {
		return nil, fmt.Errorf("awsutil: failed to load default config for role: %w", err)
	}

	// Create STS client for assuming role
	client := sts.NewFromConfig(cfg)

	// Create role options
	roleOptions := func(o *stscreds.AssumeRoleOptions) {
		o.ExternalID = aws.String(config.ExternalID)
		o.RoleSessionName = createRoleSessionName(config.Service)
	}

	// Return role credentials provider
	return stscreds.NewAssumeRoleProvider(client, config.IAMRoleARN, roleOptions), nil
}
