package minio

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

type Resource struct {
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Region          string
	Client          *minio.Client
}

func (mr *Resource) ToFileManagerConfig(prefix string) map[string]any {
	return map[string]any{
		"bucketName":       mr.BucketName,
		"accessKeyID":      mr.AccessKeyID,
		"secretAccessKey":  mr.AccessKeySecret,
		"accessKey":        mr.AccessKeySecret,
		"enableSSE":        false,
		"prefix":           prefix,
		"endPoint":         mr.Endpoint,
		"s3ForcePathStyle": true,
		"disableSSL":       true,
		"useSSL":           false,
		"region":           mr.Region,
	}
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

	minioContainer, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        c.Tag,
		Cmd:        []string{"server", "/data"},
		Env: append([]string{
			fmt.Sprintf("MINIO_ACCESS_KEY=%s", accessKeyId),
			fmt.Sprintf("MINIO_SECRET_KEY=%s", secretAccessKey),
			fmt.Sprintf("MINIO_SITE_REGION=%s", region),
			"MINIO_API_SELECT_PARQUET=on",
		}, c.Options...),
	})
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
