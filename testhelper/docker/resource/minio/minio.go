package minio

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
)

type Resource struct {
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Region          string
	Client          *minio.Client
}

type File struct {
	Key     string
	Content string
	Etag    string
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...func(*Config)) (*Resource, error) {
	const (
		bucket          = "rudder-saas"
		region          = "us-east-1"
		accessKeyId     = "MYACCESSKEY"
		secretAccessKey = "MYSECRETKEY"
	)

	c := &Config{
		Tag:     "latest",
		Options: []string{},
	}
	for _, opt := range opts {
		opt(c)
	}

	var networkID string
	if c.Network != nil {
		networkID = c.Network.ID
	}

	minioContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        c.Tag,
		NetworkID:  networkID,
		Cmd:        []string{"server", "/data"},
		Env: append([]string{
			fmt.Sprintf("MINIO_ACCESS_KEY=%s", accessKeyId),
			fmt.Sprintf("MINIO_SECRET_KEY=%s", secretAccessKey),
			fmt.Sprintf("MINIO_SITE_REGION=%s", region),
			"MINIO_API_SELECT_PARQUET=on",
		}, c.Options...),
		PortBindings: internal.IPv4PortBindings([]string{"9000"}),
	}, internal.DefaultHostConfig)
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %s", err)
	}
	d.Cleanup(func() {
		if err := pool.Purge(minioContainer); err != nil {
			log.Printf("Could not purge minio resource: %s \n", err)
		}
	})

	endpoint := fmt.Sprintf("localhost:%s", minioContainer.GetPort("9000/tcp"))

	// check if minio server is up & running.
	if err := pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", endpoint)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer func() { httputil.CloseResponse(resp) }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create minio client: %w", err)
	}

	// creating bucket inside minio where testing will happen.
	if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		return nil, fmt.Errorf("could not create bucket %q: %w", bucket, err)
	}

	return &Resource{
		BucketName:      bucket,
		AccessKeyID:     accessKeyId,
		AccessKeySecret: secretAccessKey,
		Endpoint:        endpoint,
		Region:          region,
		Client:          client,
	}, nil
}

func (r *Resource) Contents(ctx context.Context, prefix string) ([]File, error) {
	contents := make([]File, 0)

	opts := minio.ListObjectsOptions{
		Recursive: true,
		Prefix:    prefix,
	}
	for objInfo := range r.Client.ListObjects(ctx, r.BucketName, opts) {
		if objInfo.Err != nil {
			return nil, objInfo.Err
		}

		o, err := r.Client.GetObject(ctx, r.BucketName, objInfo.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}

		var r io.Reader
		br := bufio.NewReader(o)
		magic, err := br.Peek(2)
		// check if the file is gzipped using the magic number
		if err == nil && magic[0] == 31 && magic[1] == 139 {
			r, err = gzip.NewReader(br)
			if err != nil {
				return nil, fmt.Errorf("gunzip: %w", err)
			}
		} else {
			r = br
		}

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		contents = append(contents, File{
			Key:     objInfo.Key,
			Content: string(b),
			Etag:    objInfo.ETag,
		})
	}

	slices.SortStableFunc(contents, func(a, b File) int {
		return strings.Compare(a.Key, b.Key)
	})

	return contents, nil
}

func (r *Resource) ToFileManagerConfig(prefix string) map[string]any {
	return map[string]any{
		"bucketName":       r.BucketName,
		"accessKeyID":      r.AccessKeyID,
		"secretAccessKey":  r.AccessKeySecret,
		"accessKey":        r.AccessKeySecret,
		"enableSSE":        false,
		"prefix":           prefix,
		"endPoint":         r.Endpoint,
		"s3ForcePathStyle": true,
		"disableSSL":       true,
		"useSSL":           false,
		"region":           r.Region,
	}
}
