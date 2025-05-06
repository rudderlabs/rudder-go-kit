package filemanager

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func TestNewS3ManagerWithNil(t *testing.T) {
	s3Manager, err := NewS3Manager(nil, nil, logger.NOP, func() time.Duration { return time.Minute })
	assert.EqualError(t, err, "config should not be nil")
	assert.Nil(t, s3Manager)
}

func TestNewS3ManagerWithAccessKeys(t *testing.T) {
	s3Manager, err := NewS3Manager(config.Default, map[string]interface{}{
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
	s3Manager, err := NewS3Manager(config.Default, map[string]interface{}{
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
	s3Manager, err := NewS3Manager(config.Default, map[string]interface{}{
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
	s3Manager, err := NewS3Manager(config.Default, map[string]interface{}{
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
	s3Manager := S3Manager{
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
	s3Manager := S3Manager{
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
