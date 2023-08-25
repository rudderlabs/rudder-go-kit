package resource

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

type MINIOResource struct {
	Endpoint     string
	BucketName   string
	Port         string
	AccessKey    string
	SecretKey    string
	SiteRegion   string
	ResourceName string
	Client       *minio.Client
}

const (
	minioRegion     = "us-east-1"
	minioAccessKey  = "MYACCESSKEY"
	minioSecretKey  = "MYSECRETKEY"
	minioBucketName = "testbucket"
)

func SetupMINIO(pool *dockertest.Pool, d cleaner) (*MINIOResource, error) {
	// Setup MINIO
	var minioClient *minio.Client

	options := &dockertest.RunOptions{
		Hostname:   "minio",
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ACCESS_KEY=" + minioAccessKey,
			"MINIO_SECRET_KEY=" + minioSecretKey,
			"MINIO_SITE_REGION=" + minioBucketName,
		},
	}

	minioContainer, err := pool.RunWithOptions(options)
	if err != nil {
		return nil, err
	}
	d.Cleanup(func() {
		if err := pool.Purge(minioContainer); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	minioEndpoint := fmt.Sprintf("localhost:%s", minioContainer.GetPort("9000/tcp"))

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	// the minio client does not do service discovery for you (i.e. it does not check if connection can be established), so we have to use the health check
	if err := pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", minioEndpoint)
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer func() { httputil.CloseResponse(resp) }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// now we can instantiate minio client
	minioClient, err = minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	if err = minioClient.MakeBucket(context.Background(), minioBucketName, minio.MakeBucketOptions{Region: minioRegion}); err != nil {
		return nil, err
	}
	return &MINIOResource{
		Endpoint:     minioEndpoint,
		BucketName:   minioBucketName,
		Port:         minioContainer.GetPort("9000/tcp"),
		AccessKey:    minioAccessKey,
		SecretKey:    minioSecretKey,
		SiteRegion:   minioRegion,
		Client:       minioClient,
		ResourceName: minioContainer.Container.Name,
	}, nil
}
