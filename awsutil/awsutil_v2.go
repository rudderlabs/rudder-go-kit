package awsutil

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func CreateAWSConfig(ctx context.Context, config *SessionConfig) (aws.Config, error) {
	var (
		zero           aws.Config
		err            error
		awsCredentials aws.CredentialsProvider
	)

	if config.RoleBasedAuth {
		awsCredentials, err = createV2CredentialsForRole(ctx, config)
		if err != nil {
			return zero, err
		}
	} else if config.AccessKey != "" && config.AccessKeyID != "" {
		awsCredentials = credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKey, "")
	}

	return aws.Config{
		Region:      config.Region,
		Credentials: awsCredentials,
	}, nil
}

func createDefaultConfig(ctx context.Context, config *SessionConfig) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.Region),
		// awsconfig.WithTimeout(config.Timeout),
	)
}

func createV2CredentialsForRole(ctx context.Context, config *SessionConfig) (aws.CredentialsProvider, error) {
	var zero aws.CredentialsProvider

	if config.ExternalID == "" {
		return zero, errors.New("externalID is required for IAM role")
	}
	cfg, err := createDefaultConfig(ctx, config)
	if err != nil {
		return zero, err
	}

	client := sts.NewFromConfig(cfg)

	appCreds := stscreds.NewAssumeRoleProvider(client, config.IAMRoleARN, func(aro *stscreds.AssumeRoleOptions) {
		aro.ExternalID = aws.String(config.ExternalID)
		aro.RoleSessionName = createRoleSessionName(config.Service)
	})
	return appCreds, nil
}
