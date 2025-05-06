package awsutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
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

	if config.RoleBasedAuth {
		awsCredentials, err = createV2CredentialsForRole(ctx, config)
		if err != nil {
			return zero, fmt.Errorf("awsutil: failed to create credentials for role: %w", err)
		}
	} else if config.AccessKeyID != "" && config.AccessKey != "" {
		awsCredentials = credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKey, "")
	}

	return aws.Config{
		Region:      config.Region,
		Credentials: awsCredentials,
	}, nil
}

// createDefaultConfig loads the default AWS config with the provided SessionConfig.
func createDefaultConfig(ctx context.Context, config *SessionConfig) (aws.Config, error) {
	customClient := awshttp.NewBuildableClient()
	if config.Timeout != nil {
		customClient = customClient.WithTimeout(*config.Timeout)
	}

	return awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.Region),
		awsconfig.WithHTTPClient(customClient),
	)
}

// createV2CredentialsForRole creates a credentials provider for assuming an IAM role.
func createV2CredentialsForRole(ctx context.Context, config *SessionConfig) (aws.CredentialsProvider, error) {
	var zero aws.CredentialsProvider

	if config.ExternalID == "" {
		return zero, errors.New("awsutil: externalID is required for IAM role")
	}
	cfg, err := createDefaultConfig(ctx, config)
	if err != nil {
		return zero, fmt.Errorf("awsutil: failed to load default config for role: %w", err)
	}

	client := sts.NewFromConfig(cfg)

	appCreds := stscreds.NewAssumeRoleProvider(client, config.IAMRoleARN, func(aro *stscreds.AssumeRoleOptions) {
		aro.ExternalID = aws.String(config.ExternalID)
		aro.RoleSessionName = createRoleSessionName(config.Service)
	})
	return appCreds, nil
}
