package filemanager

import (
	"context"
	"os"
	"testing"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/stretchr/testify/require"
)

func TestGCSManagerOpts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := config.New()

	conf := map[string]interface{}{
		"bucketName":  "test-bucket",
		"prefix":      "test-prefix",
		"credentials": c.GetString("GCS_TEST_CREDENTIALS", "<credentials>"),
		"disableSSL":  true,
		"jsonReads":   true,
	}
	settings := &Settings{
		Provider:            "GCS",
		Config:              conf,
		Logger:              logger.NOP,
		Conf:                config.New(),
		GCSUploadIfNotExist: true,
	}

	m, err := New(settings)
	require.NoError(t, err)

	tempDir := t.TempDir()
	f, err := os.Create(tempDir + "/testFile")
	require.NoError(t, err)
	uploadedFile, err := m.Upload(ctx, f)
	require.NoError(t, err)
	require.Equal(t, "testFile", uploadedFile.ObjectName)

	uploadedFile, err = m.Upload(ctx, f)
	require.Equal(t, UploadedFile{}, uploadedFile)
	require.ErrorIs(t, err, ErrPreConditionFailed)
}
