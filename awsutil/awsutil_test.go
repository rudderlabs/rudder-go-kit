package awsutil

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRegionFromBucket(t *testing.T) {
	envBucket := os.Getenv("AWS_BUCKET_NAME")
	region, err := GetRegionFromBucket(context.Background(), envBucket, "us-east-1")
	require.NoError(t, err)
	require.Equal(t, "us-east-1", region)
}

func TestRoleSessionName(t *testing.T) {
	require.Equal(t, "rudderstack-aws-s3-access", roleSessionName(&SessionConfig{Service: "S3"}))
	require.Equal(t, "30bK6N9S6Ca7C0SGITpgVsmRlIs", roleSessionName(&SessionConfig{Service: "S3", RoleSessionName: "30bK6N9S6Ca7C0SGITpgVsmRlIs"}))
}
