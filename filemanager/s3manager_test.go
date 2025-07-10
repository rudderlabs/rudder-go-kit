package filemanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
	awsutilV2 "github.com/rudderlabs/rudder-go-kit/awsutil_v2"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

// mockRegionFetcher implements RegionFetcher for testing
type mockRegionFetcher struct {
	region string
	err    error
}

func (m *mockRegionFetcher) GetBucketRegion(ctx context.Context, client *s3.Client, bucket string) (string, error) {
	return m.region, m.err
}

func TestNewS3ManagerWithNil(t *testing.T) {
	s3Manager, err := newS3ManagerV1(nil, nil, logger.NOP, func() time.Duration { return time.Minute })
	assert.EqualError(t, err, "config should not be nil")
	assert.Nil(t, s3Manager)
}

func TestNewS3ManagerWithAccessKeys(t *testing.T) {
	s3Manager, err := newS3ManagerV1(config.Default, map[string]interface{}{
		"bucketName":  "someBucket",
		"region":      "someRegion",
		"accessKeyID": "someAccessKeyId",
		"accessKey":   "someSecretAccessKey",
	}, logger.NOP, func() time.Duration { return time.Minute })
	assert.Nil(t, err)
	assert.NotNil(t, s3Manager)
	assert.Equal(t, "someBucket", s3Manager.config.Bucket)
	assert.Equal(t, aws.String("someRegion"), s3Manager.config.Region)
	assert.Equal(t, "someAccessKeyId", s3Manager.sessionConfig.AccessKeyID)
	assert.Equal(t, "someSecretAccessKey", s3Manager.sessionConfig.AccessKey)
	assert.Equal(t, false, s3Manager.sessionConfig.RoleBasedAuth)
}

func TestNewS3ManagerWithRole(t *testing.T) {
	s3Manager, err := newS3ManagerV1(config.Default, map[string]interface{}{
		"bucketName": "someBucket",
		"region":     "someRegion",
		"iamRoleARN": "someIAMRole",
		"externalID": "someExternalID",
	}, logger.NOP, func() time.Duration { return time.Minute })
	assert.Nil(t, err)
	assert.NotNil(t, s3Manager)
	assert.Equal(t, "someBucket", s3Manager.config.Bucket)
	assert.Equal(t, aws.String("someRegion"), s3Manager.config.Region)
	assert.Equal(t, "someIAMRole", s3Manager.sessionConfig.IAMRoleARN)
	assert.Equal(t, "someExternalID", s3Manager.sessionConfig.ExternalID)
	assert.Equal(t, true, s3Manager.sessionConfig.RoleBasedAuth)
}

func TestNewS3ManagerWithBothAccessKeysAndRole(t *testing.T) {
	s3Manager, err := newS3ManagerV1(config.Default, map[string]interface{}{
		"bucketName":  "someBucket",
		"region":      "someRegion",
		"iamRoleARN":  "someIAMRole",
		"externalID":  "someExternalID",
		"accessKeyID": "someAccessKeyId",
		"accessKey":   "someSecretAccessKey",
	}, logger.NOP, func() time.Duration { return time.Minute })
	assert.Nil(t, err)
	assert.NotNil(t, s3Manager)
	assert.Equal(t, "someBucket", s3Manager.config.Bucket)
	assert.Equal(t, aws.String("someRegion"), s3Manager.config.Region)
	assert.Equal(t, "someAccessKeyId", s3Manager.sessionConfig.AccessKeyID)
	assert.Equal(t, "someSecretAccessKey", s3Manager.sessionConfig.AccessKey)
	assert.Equal(t, "someIAMRole", s3Manager.sessionConfig.IAMRoleARN)
	assert.Equal(t, "someExternalID", s3Manager.sessionConfig.ExternalID)
	assert.Equal(t, true, s3Manager.sessionConfig.RoleBasedAuth)
}

func TestNewS3ManagerWithBothAccessKeysAndRoleButRoleBasedAuthFalse(t *testing.T) {
	s3Manager, err := newS3ManagerV1(config.Default, map[string]interface{}{
		"bucketName":    "someBucket",
		"region":        "someRegion",
		"iamRoleARN":    "someIAMRole",
		"externalID":    "someExternalID",
		"accessKeyID":   "someAccessKeyId",
		"accessKey":     "someSecretAccessKey",
		"roleBasedAuth": false,
	}, logger.NOP, func() time.Duration { return time.Minute })
	assert.Nil(t, err)
	assert.NotNil(t, s3Manager)
	assert.Equal(t, "someBucket", s3Manager.config.Bucket)
	assert.Equal(t, aws.String("someRegion"), s3Manager.config.Region)
	assert.Equal(t, "someAccessKeyId", s3Manager.sessionConfig.AccessKeyID)
	assert.Equal(t, "someSecretAccessKey", s3Manager.sessionConfig.AccessKey)
	assert.Equal(t, "someIAMRole", s3Manager.sessionConfig.IAMRoleARN)
	assert.Equal(t, "someExternalID", s3Manager.sessionConfig.ExternalID)
	assert.Equal(t, false, s3Manager.sessionConfig.RoleBasedAuth)
}

func TestGetSessionWithAccessKeys(t *testing.T) {
	s3Manager := s3ManagerV1{
		baseManager: &baseManager{
			logger: logger.NOP,
		},
		config: &S3Config{
			Bucket: "someBucket",
			Region: aws.String("someRegion"),
		},
		sessionConfig: &awsutil.SessionConfig{
			AccessKeyID: "someAccessKeyId",
			AccessKey:   "someSecretAccessKey",
			Region:      "someRegion",
		},
	}
	awsSession, err := s3Manager.GetSession(context.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, awsSession)
	assert.NotNil(t, s3Manager.session)
}

func TestGetSessionWithIAMRole(t *testing.T) {
	s3Manager := s3ManagerV1{
		baseManager: &baseManager{
			logger: logger.NOP,
		},
		config: &S3Config{
			Bucket: "someBucket",
			Region: aws.String("someRegion"),
		},
		sessionConfig: &awsutil.SessionConfig{
			IAMRoleARN: "someIAMRole",
			ExternalID: "someExternalID",
			Region:     "someRegion",
		},
	}
	awsSession, err := s3Manager.GetSession(context.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, awsSession)
	assert.NotNil(t, s3Manager.session)
}

func TestS3ManagerV2GetClientWithRegionFetchFailure(t *testing.T) {
	mockFetcher := &mockRegionFetcher{
		region: "",
		err:    errors.New("context deadline exceeded"),
	}

	s3Manager := &s3ManagerV2{
		baseManager: &baseManager{
			logger:         logger.NOP,
			defaultTimeout: func() time.Duration { return time.Minute },
		},
		config: &S3Config{
			Bucket:     "non-existent-bucket-12345",
			Region:     nil,
			RegionHint: "us-east-1",
		},
		sessionConfig: &awsutilV2.SessionConfig{
			AccessKeyID: "test-access-key",
			AccessKey:   "test-secret-key",
			Region:      "",
		},
		regionFetcher: mockFetcher,
	}

	client, err := s3Manager.getClient(context.Background())

	assert.Nil(t, err)
	assert.NotNil(t, client)
}
