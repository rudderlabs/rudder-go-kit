package filemanager

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/minio/minio-go/v7"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	miniotest "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/minio"
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

func TestEmptyRegion(t *testing.T) {
	const (
		prefix      = "some-prefix"
		objectName  = "minio.object"
		fileContent = "This is test content for download functionality"
		// Expected error message when region is missing and bucket is wrong
		expectedRegionError = "failed to download from S3: operation error S3: GetObject, failed to resolve service endpoint, endpoint rule error, A region must be set when sending requests to S3."
	)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	minioResource, err := miniotest.Setup(pool, t)
	require.NoError(t, err)

	// Upload test file using minio client
	_, err = minioResource.Client.PutObject(context.Background(),
		minioResource.BucketName, prefix+"/"+objectName,
		bytes.NewReader([]byte(fileContent)), int64(len(fileContent)),
		minio.PutObjectOptions{},
	)
	require.NoError(t, err)

	createFileManager := func(configModifications map[string]interface{}) (FileManager, error) {
		c := minioResource.ToFileManagerConfig(prefix)
		c["region"] = ""
		c["endPoint"] = "https://" + c["endPoint"].(string)

		// Apply any config modifications
		for key, value := range configModifications {
			c[key] = value
		}

		config := config.New()
		config.Set("FileManager.useAwsSdkV2", true)

		return New(&Settings{
			Provider: "S3",
			Config:   c,
			Conf:     config,
		})
	}

	t.Run("should create a session and download file even when region is empty", func(t *testing.T) {
		fm, err := createFileManager(nil)
		require.NoError(t, err)

		// Create temporary file for download
		tempFile, err := os.CreateTemp(t.TempDir(), "downloaded-*.txt")
		require.NoError(t, err)
		defer tempFile.Close()

		// Download the file
		err = fm.Download(context.Background(), tempFile, prefix+"/"+objectName)
		require.NoError(t, err)

		// Read downloaded content
		_, err = tempFile.Seek(0, 0) // Reset file pointer to beginning
		require.NoError(t, err)
		downloadedContent, err := io.ReadAll(tempFile)
		require.NoError(t, err)

		// Verify content matches
		require.Equal(t, fileContent, string(downloadedContent))
	})

	t.Run("should try to download even if fetching region fails", func(t *testing.T) {
		fm, err := createFileManager(map[string]interface{}{
			"bucketName": "wrong-bucket-name",
		})
		require.NoError(t, err)

		// Wrong bucket name will trigger region fetch failure
		err = fm.Download(context.Background(), nil, prefix+"/"+objectName)
		require.Error(t, err)
		require.Equal(t, expectedRegionError, err.Error())
	})
}
