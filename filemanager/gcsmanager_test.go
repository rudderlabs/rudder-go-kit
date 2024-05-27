package filemanager

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/testhelper"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestGCSManagerOpts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// testcases:
	tcs := []struct {
		Name                string
		GCSUploadIfNotExist bool
	}{
		{
			Name:                "without UploadIfNotExist",
			GCSUploadIfNotExist: true,
		},
		{
			Name:                "with UploadIfNotExist",
			GCSUploadIfNotExist: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			port, err := testhelper.GetFreePort()
			require.NoError(t, err)

			server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
				InitialObjects: []fakestorage.Object{
					{
						ObjectAttrs: fakestorage.ObjectAttrs{
							BucketName: "test-bucket",
							Name:       "test-prefix/testFile",
						},
						Content: []byte("inside the file"),
					},
				},
				Scheme: "http",
				Host:   "127.0.0.1",
				Port:   uint16(port),
			})
			require.NoError(t, err)
			defer server.Stop()

			gcsURL := fmt.Sprintf("%s/storage/v1/", server.URL())
			t.Log("GCS URL:", gcsURL)

			conf := map[string]interface{}{
				"bucketName": "test-bucket",
				"prefix":     "test-prefix",
				"endPoint":   gcsURL,
				"disableSSL": true,
				"jsonReads":  true,
			}
			m, err := New(&Settings{
				Provider:            "GCS",
				Config:              conf,
				Logger:              logger.NOP,
				Conf:                config.New(),
				GCSUploadIfNotExist: tc.GCSUploadIfNotExist,
			})
			require.NoError(t, err)

			tempDir := t.TempDir()
			f, err := os.Create(tempDir + "/testFile")
			require.NoError(t, err)

			t.Log("pre-existing file")
			uploadedFile, err := m.Upload(ctx, f)
			if tc.GCSUploadIfNotExist {
				require.Equal(t, UploadedFile{}, uploadedFile)
				require.ErrorIs(t, err, ErrPreConditionFailed)
			} else {
				require.NoError(t, err)
			}

			t.Run("new file", func(t *testing.T) {
				tempDir := t.TempDir()
				f, err := os.Create(tempDir + "/testFile-new")
				require.NoError(t, err)

				_, err = m.Upload(ctx, f)
				require.NoError(t, err)
			})
		})
	}
}
