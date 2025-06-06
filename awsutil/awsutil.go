package awsutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
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
	Endpoint            *string        `mapstructure:"endpoint"`
	S3ForcePathStyle    *bool          `mapstructure:"s3ForcePathStyle"`
	DisableSSL          *bool          `mapstructure:"disableSSL"`
	Service             string         `mapstructure:"service"`
	Timeout             *time.Duration `mapstructure:"timeout"`
	SharedConfigProfile string         `mapstructure:"sharedConfigProfile"`
	MaxIdleConnsPerHost int            `mapstructure:"maxIdleConnsPerHost"`
}

// CreateSession creates a new AWS session using the provided config
func CreateSession(config *SessionConfig) (*session.Session, error) {
	var (
		awsCredentials *credentials.Credentials
		err            error
	)
	if config.RoleBasedAuth {
		awsCredentials, err = createCredentialsForRole(config)
	} else if config.AccessKey != "" && config.AccessKeyID != "" {
		awsCredentials, err = credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKey, config.SessionToken), nil
	}
	if err != nil {
		return nil, err
	}

	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			HTTPClient:                    getHttpClient(config),
			Region:                        aws.String(config.Region),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Credentials:                   awsCredentials,
			Endpoint:                      config.Endpoint,
			S3ForcePathStyle:              config.S3ForcePathStyle,
			DisableSSL:                    config.DisableSSL,
		},
		Profile: config.SharedConfigProfile,
	})
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

func getHttpClient(config *SessionConfig) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

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
	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
		transport.MaxIdleConns = max(transport.MaxIdleConns, config.MaxIdleConnsPerHost)
	}

	httpClient := &http.Client{
		Transport: transport,
	}
	if config.Timeout != nil {
		httpClient.Timeout = *config.Timeout
	}

	return httpClient
}

func createDefaultSession(config *SessionConfig) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		HTTPClient: getHttpClient(config),
		Region:     aws.String(config.Region),
	})
}

func createRoleSessionName(serviceName string) string {
	return fmt.Sprintf("rudderstack-aws-%s-access", strings.ToLower(strings.ReplaceAll(serviceName, " ", "-")))
}

func createCredentialsForRole(config *SessionConfig) (*credentials.Credentials, error) {
	if config.ExternalID == "" {
		return nil, errors.New("externalID is required for IAM role")
	}
	hostSession, err := createDefaultSession(config)
	if err != nil {
		return nil, err
	}
	return stscreds.NewCredentials(hostSession, config.IAMRoleARN,
		func(p *stscreds.AssumeRoleProvider) {
			p.ExternalID = aws.String(config.ExternalID)
			p.RoleSessionName = createRoleSessionName(config.Service)
		}), err
}

func isRoleBasedAuthFieldExist(config map[string]interface{}) bool {
	_, ok := config["roleBasedAuth"].(bool)
	return ok
}
