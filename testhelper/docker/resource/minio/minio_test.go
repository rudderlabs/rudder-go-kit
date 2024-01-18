package minio

import (
	"context"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

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
		})
		require.NoError(t, err)

		it := fm.ListFilesWithPrefix(context.Background(), "", "some-prefix", 1)
		items, err := it.Next()
		require.NoError(t, err)
		require.Len(t, items, 1)
	})
}
