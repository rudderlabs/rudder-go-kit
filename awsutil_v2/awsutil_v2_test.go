package awsutil_v2

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRegionFromBucket(t *testing.T) {
	envBucket := os.Getenv("AWS_BUCKET_NAME")
	region, err := GetRegionFromBucket(context.Background(), envBucket, "us-east-1")
	if err != nil {
		t.Fatalf("failed to get region from bucket: %v", err)
	}
	require.Equal(t, "us-east-1", region)
}
