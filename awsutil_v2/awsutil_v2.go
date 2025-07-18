package awsutil_v2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mitchellh/mapstructure"
)

// Some AWS destinations are using SecretAccessKey instead of accessKey
type SessionConfig struct {
	Region              string         `mapstructure:"region"`
	AccessKeyID         string         `mapstructure:"accessKeyID"`
	AccessKey           string         `mapstructure:"accessKey"`
	SecretAccessKey     string         `mapstructure:"secretAccessKey"`
	SessionToken        string         `mapstructure:"sessionToken"`
	RoleBasedAuth       bool           `mapstructure:"roleBasedAuth"`
	IAMRoleARN          string         `mapstructure:"iamRoleARN"`
	ExternalID          string         `mapstructure:"externalID"`
	WorkspaceID         string         `mapstructure:"workspaceID"`
	Service             string         `mapstructure:"service"`
	Timeout             *time.Duration `mapstructure:"timeout"`
	SharedConfigProfile string         `mapstructure:"sharedConfigProfile"`
	MaxIdleConnsPerHost int            `mapstructure:"maxIdleConnsPerHost"`
}

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

	httpClient := configureHTTPClient(config)
	// Handle credentials - either role-based or static
	if config.RoleBasedAuth {
		awsCredentials, err = createV2CredentialsForRole(ctx, httpClient, config)
		if err != nil {
			return zero, fmt.Errorf("awsutil: failed to create credentials for role: %w", err)
		}
	} else if config.AccessKeyID != "" && config.AccessKey != "" {
		awsCredentials = credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKey, config.SessionToken)
	}

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
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create transport with proper connection settings
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
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
func createV2CredentialsForRole(ctx context.Context, httpClient *http.Client, config *SessionConfig) (aws.CredentialsProvider, error) {
	if config.ExternalID == "" {
		return nil, errors.New("awsutil: externalID is required for IAM role")
	}

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

func createRoleSessionName(serviceName string) string {
	return fmt.Sprintf("rudderstack-aws-%s-access", strings.ToLower(strings.ReplaceAll(serviceName, " ", "-")))
}

// NewSimpleSessionConfig creates a new session config using the provided config map
func NewSimpleSessionConfig(config map[string]interface{}, serviceName string) (*SessionConfig, error) {
	if config == nil {
		return nil, errors.New("config should not be nil")
	}
	sessionConfig := SessionConfig{}
	if err := mapstructure.Decode(config, &sessionConfig); err != nil {
		return nil, fmt.Errorf("unable to populate session config using destinationConfig: %w", err)
	}

	if !isRoleBasedAuthFieldExist(config) {
		sessionConfig.RoleBasedAuth = sessionConfig.IAMRoleARN != ""
	}

	if sessionConfig.IAMRoleARN == "" {
		sessionConfig.RoleBasedAuth = false
	}

	// Some AWS destinations are using SecretAccessKey instead of accessKey
	if sessionConfig.SecretAccessKey != "" {
		sessionConfig.AccessKey = sessionConfig.SecretAccessKey
	}
	sessionConfig.Service = serviceName
	return &sessionConfig, nil
}

func GetRegionFromBucket(ctx context.Context, bucket, regionHint string) (string, error) {
	region, err := s3manager.GetBucketRegion(ctx, s3.New(s3.Options{
		Region: regionHint,
	}), bucket)
	if err != nil {
		return "", fmt.Errorf("failed to fetch AWS region for bucket: %w", err)
	}
	return region, nil
}

func isRoleBasedAuthFieldExist(config map[string]interface{}) bool {
	_, ok := config["roleBasedAuth"].(bool)
	return ok
}
