package minio

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/ory/dockertest/v3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/filemanager"
)

func TestMinioResource(t *testing.T) {
	const prefix = "some-prefix"
	const objectName = "minio.object"

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	minioResource, err := Setup(pool, t)
	require.NoError(t, err)

	_, err = minioResource.Client.FPutObject(context.Background(),
		minioResource.BucketName, prefix+"/"+objectName, "testdata/minio.object", minio.PutObjectOptions{},
	)
	require.NoError(t, err)
	c := minioResource.ToFileManagerConfig("some-prefix")

	t.Run("can use a minio filemanager", func(t *testing.T) {
		fm, err := filemanager.New(&filemanager.Settings{
			Provider: "MINIO",
			Config:   c,
			Conf:     config.New(),
		})
		require.NoError(t, err)

		it := fm.ListFilesWithPrefix(context.Background(), "", "some-prefix", 1)
		items, err := it.Next()
		require.NoError(t, err)
		require.Len(t, items, 1)
	})

	t.Run("can use a s3 filemanager", func(t *testing.T) {
		fm, err := filemanager.New(&filemanager.Settings{
			Provider: "S3",
			Config:   c,
			Conf:     config.New(),
		})
		require.NoError(t, err)

		it := fm.ListFilesWithPrefix(context.Background(), "", "some-prefix", 1)
		items, err := it.Next()
		require.NoError(t, err)
		require.Len(t, items, 1)
	})
}

func TestMinioContents(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	minioResource, err := Setup(pool, t)
	require.NoError(t, err)

	uploadInfo, err := minioResource.Client.PutObject(context.Background(),
		minioResource.BucketName, "test-bucket/hello.txt", bytes.NewBufferString("hello"), -1, minio.PutObjectOptions{},
	)
	require.NoError(t, err)
	etag1 := uploadInfo.ETag

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write([]byte("hello compressed"))
	require.NoError(t, err)
	err = gz.Close()
	require.NoError(t, err)

	uploadInfo, err = minioResource.Client.PutObject(context.Background(),
		minioResource.BucketName, "test-bucket/hello.txt.gz", &b, -1, minio.PutObjectOptions{},
	)
	require.NoError(t, err)
	etag2 := uploadInfo.ETag

	uploadInfo, err = minioResource.Client.PutObject(context.Background(),
		minioResource.BucketName, "test-bucket/empty", bytes.NewBuffer([]byte{}), -1, minio.PutObjectOptions{},
	)
	require.NoError(t, err)
	etag3 := uploadInfo.ETag

	files, err := minioResource.Contents(context.Background(), "test-bucket/")
	require.NoError(t, err)

	// LastModified is set after the file is uploaded, so we can't compare it
	lo.ForEach(files, func(f File, _ int) {
		switch f.Key {
		case "test-bucket/hello.txt":
			require.Equal(t, "hello", f.Content)
			require.Equal(t, etag1, f.Etag)
		case "test-bucket/hello.txt.gz":
			require.Equal(t, "hello compressed", f.Content)
			require.Equal(t, etag2, f.Etag)
		case "test-bucket/empty":
			require.Equal(t, "", f.Content)
			require.Equal(t, etag3, f.Etag)
		default:
			t.Fatalf("unexpected file: %s", f.Key)
		}
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := minioResource.Contents(ctx, "test-bucket/")
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("can upload a folder", func(t *testing.T) {
		expectedContents := map[string]string{
			"test-upload-folder/file1.txt": "file1",
			"test-upload-folder/file2.txt": "file2",
		}

		tempDir := t.TempDir()
		// add some files to the temp dir
		file1, err := os.Create(tempDir + "/file1.txt")
		require.NoError(t, err, "could not create file1.txt")
		_, err = file1.WriteString(expectedContents["test-upload-folder/file1.txt"])
		require.NoError(t, err, "could not write to file1.txt")

		file2, err := os.Create(tempDir + "/file2.txt")
		require.NoError(t, err, "could not create file2.txt")
		_, err = file2.WriteString(expectedContents["test-upload-folder/file2.txt"])
		require.NoError(t, err, "could not write to file2.txt")

		err = minioResource.UploadFolder(tempDir, "test-upload-folder")
		require.NoError(t, err, "could not upload folder")

		files, err := minioResource.Contents(context.Background(), "test-upload-folder")
		require.NoError(t, err)
		require.Len(t, files, 2, "should have uploaded 2 files")

		lo.ForEach(files, func(f File, _ int) {
			if expectedContent, exists := expectedContents[f.Key]; exists {
				require.Equal(t, expectedContent, f.Content)
			}
		})
	})
}
