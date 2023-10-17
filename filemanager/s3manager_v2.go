package filemanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
	appConfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const ServiceName = "s3"

type S3ManagerV2 struct {
	*baseManager
	config *S3Config

	sessionConfig *awsutil.SessionConfig
	client        *s3.Client
	clientMu      sync.Mutex
}

// NewS3ManagerV2 creates a new file manager for S3 using v2 aws sdk
func NewS3ManagerV2(
	config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration,
) (*S3ManagerV2, error) {
	var s3Config S3Config
	if err := mapstructure.Decode(config, &s3Config); err != nil {
		return nil, err
	}

	sessionConfig, err := awsutil.NewSimpleSessionConfig(config, ServiceName)
	if err != nil {
		return nil, err
	}

	s3Config.RegionHint = appConfig.GetString("AWS_S3_REGION_HINT", "us-east-1")

	return &S3ManagerV2{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:        &s3Config,
		sessionConfig: sessionConfig,
	}, nil
}

func (m *S3ManagerV2) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &s3ListSessionV2{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager:     m,
		isTruncated: true,
	}
}

// Download downloads a file from S3
func (m *S3ManagerV2) Download(ctx context.Context, output *os.File, key string) error {
	client, err := m.getClient(ctx)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}

	downloader := s3manager.NewDownloader(client)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	_, err = downloader.Download(ctx, output,
		&s3.GetObjectInput{
			Bucket: aws.String(m.config.Bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			// handle NoSuchKey error
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

// Upload uploads a file to S3
func (m *S3ManagerV2) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	fileName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))

	uploadInput := &s3.PutObjectInput{
		ACL:    types.ObjectCannedACLBucketOwnerFullControl,
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	}

	if m.config.EnableSSE {
		uploadInput.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	client, err := m.getClient(ctx)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("s3 client: %w", err)
	}
	uploader := s3manager.NewUploader(client)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	output, err := uploader.Upload(ctx, uploadInput)
	if err != nil {
		var regionError *aws.MissingRegionError
		if errors.As(err, &regionError) {
			err = fmt.Errorf(`missing region for bucket %q: %w`, m.config.Bucket, regionError)
		}
		return UploadedFile{}, err
	}

	return UploadedFile{Location: output.Location, ObjectName: fileName}, err
}

func (m *S3ManagerV2) Delete(ctx context.Context, keys []string) (err error) {
	client, err := m.getClient(ctx)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}

	var objects []types.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{Key: aws.String(key)})
	}

	batchSize := 1000 // max accepted by DeleteObjects API
	chunks := lo.Chunk(objects, batchSize)
	for _, chunk := range chunks {
		deleteCtx, cancel := context.WithTimeout(ctx, m.getTimeout())
		_, err := client.DeleteObjects(deleteCtx, &s3.DeleteObjectsInput{
			Bucket: aws.String(m.config.Bucket),
			Delete: &types.Delete{
				Objects: chunk,
			},
		})
		cancel()
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				m.logger.Errorf(`Error while deleting S3 objects: %v, error code: %q`, err.Error(), apiErr.ErrorCode())
			} else {
				m.logger.Errorf(`Error while deleting S3 objects: %v`, err.Error())
			}

			return err
		}
	}
	return nil
}

func (m *S3ManagerV2) Prefix() string {
	return m.config.Prefix
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://bucket-name.s3.amazonaws.com/key - >> key
*/
func (m *S3ManagerV2) GetObjectNameFromLocation(location string) (string, error) {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	trimmedURL := strings.TrimLeft(parsedUrl.Path, "/")
	if (m.config.S3ForcePathStyle != nil && *m.config.S3ForcePathStyle) ||
		(!strings.Contains(parsedUrl.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket)), nil
	}
	return trimmedURL, nil
}

func (m *S3ManagerV2) GetDownloadKeyFromFileLocation(location string) string {
	parsedURL, err := url.Parse(location)
	if err != nil {
		fmt.Println("error while parsing location url: ", err)
	}
	trimmedURL := strings.TrimLeft(parsedURL.Path, "/")
	if (m.config.S3ForcePathStyle != nil && *m.config.S3ForcePathStyle) ||
		(!strings.Contains(parsedURL.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket))
	}
	return trimmedURL
}

func (m *S3ManagerV2) getClient(ctx context.Context) (*s3.Client, error) {
	m.clientMu.Lock()
	defer m.clientMu.Unlock()

	if m.client != nil {
		return m.client, nil
	}

	if m.config.Bucket == "" {
		return nil, errors.New("no storage bucket configured to downloader")
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	if !m.config.UseGlue || m.config.Region == nil {
		region, err := s3manager.GetBucketRegion(ctx, s3.New(s3.Options{
			Region: aws.ToString(&m.config.RegionHint),
		}), m.config.Bucket)
		if err != nil {
			return nil, err
		}
		m.config.Region = aws.String(region)
		m.sessionConfig.Region = region
	}

	cnf, err := awsutil.CreateAWSConfig(ctx, m.sessionConfig)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cnf, func(o *s3.Options) {
		if m.config.Endpoint != nil {
			o.BaseEndpoint = aws.String("http://" + *m.config.Endpoint)
		}

		o.UsePathStyle = aws.ToBool(m.config.S3ForcePathStyle)
		o.EndpointOptions.DisableHTTPS = aws.ToBool(m.config.DisableSSL)
		if m.timeout != 0 {
			o.HTTPClient = &http.Client{
				Timeout: m.timeout,
			}
		}
	})

	m.client = client

	return m.client, err
}

func (m *S3ManagerV2) getTimeout() time.Duration {
	if m.timeout > 0 {
		return m.timeout
	}
	if m.defaultTimeout != nil {
		return m.defaultTimeout()
	}
	return defaultTimeout
}

type s3ListSessionV2 struct {
	*baseListSession
	manager *S3ManagerV2

	continuationToken *string
	isTruncated       bool
}

func (l *s3ListSessionV2) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Infof("Manager is truncated: %v so returning here", l.isTruncated)
		return
	}
	fileObjects = make([]*FileInfo, 0)

	client, err := manager.getClient(l.ctx)
	if err != nil {
		return []*FileInfo{}, fmt.Errorf("s3 client: %w", err)
	}

	listObjectsV2Input := s3.ListObjectsV2Input{
		Bucket:  aws.String(manager.config.Bucket),
		Prefix:  aws.String(l.prefix),
		MaxKeys: int32(l.maxItems),
		// Delimiter: aws.String("/"),
		ContinuationToken: l.continuationToken,
	}
	// startAfter is to resume a paused task.
	if l.startAfter != "" {
		listObjectsV2Input.StartAfter = aws.String(l.startAfter)
	}

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	// Get the list of items
	resp, err := client.ListObjectsV2(ctx, &listObjectsV2Input)
	if err != nil {
		manager.logger.Errorf("Error while listing S3 objects: %v", err)
		return
	}
	l.isTruncated = resp.IsTruncated
	l.continuationToken = resp.NextContinuationToken
	for _, item := range resp.Contents {
		fileObjects = append(fileObjects, &FileInfo{*item.Key, *item.LastModified})
	}
	return
}
