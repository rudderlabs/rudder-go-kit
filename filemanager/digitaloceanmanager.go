package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type DigitalOceanConfig struct {
	Bucket         string  `mapstructure:"bucketName"`
	Prefix         string  `mapstructure:"prefix"`
	EndPoint       string  `mapstructure:"endPoint"`
	AccessKeyID    string  `mapstructure:"accessKeyID"`
	AccessKey      string  `mapstructure:"accessKey"`
	Region         *string `mapstructure:"region"`
	ForcePathStyle *bool   `mapstructure:"s3ForcePathStyle"`
	DisableSSL     *bool   `mapstructure:"disableSSL"`
}

type digitalOceanManager struct {
	*baseManager
	config *DigitalOceanConfig

	sessionConfig *awsutil.SessionConfig
	client        *s3.Client
	clientMu      sync.Mutex
}

func NewDigitalOceanManager(
	config map[string]any, log logger.Logger, defaultTimeout func() time.Duration,
) (*digitalOceanManager, error) {
	var doConfig DigitalOceanConfig
	if err := mapstructure.Decode(config, &doConfig); err != nil {
		return nil, fmt.Errorf("failed to decode DigitalOcean config: %w", err)
	}

	sessionConfig, err := awsutil.NewSimpleSessionConfig(config, "digitalocean")
	if err != nil {
		return nil, fmt.Errorf("failed to create DigitalOcean session config: %w", err)
	}

	// DigitalOcean Spaces requires a region, but it's often embedded in the endpoint
	if doConfig.Region == nil || *doConfig.Region == "" {
		region := getSpacesLocation(doConfig.EndPoint)
		doConfig.Region = &region
		sessionConfig.Region = region
	}

	return &digitalOceanManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:        &doConfig,
		sessionConfig: sessionConfig,
	}, nil
}

func (m *digitalOceanManager) getClient(ctx context.Context) (*s3.Client, error) {
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

	cnf, err := awsutil.CreateAWSConfig(ctx, m.sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create DigitalOcean AWS config: %w", err)
	}

	client := s3.NewFromConfig(cnf, func(o *s3.Options) {
		if m.config.EndPoint != "" {
			o.BaseEndpoint = aws.String(m.config.EndPoint)
		}
		o.UsePathStyle = aws.ToBool(m.config.ForcePathStyle)
		o.EndpointOptions.DisableHTTPS = aws.ToBool(m.config.DisableSSL)
		if m.timeout != 0 {
			o.HTTPClient = &http.Client{
				Timeout: m.timeout,
			}
		}
	})

	m.client = client
	return m.client, nil
}

func (m *digitalOceanManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &digitalOceanListSession{
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

func (m *digitalOceanManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	client, err := m.getClient(ctx)
	if err != nil {
		return fmt.Errorf("digitalocean client: %w", err)
	}

	downloader := s3manager.NewDownloader(client)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(key),
	}
	if downloadOpts.isRangeRequest {
		var rangeOpt string
		if downloadOpts.length > 0 {
			rangeOpt = fmt.Sprintf("bytes=%d-%d", downloadOpts.offset, downloadOpts.offset+downloadOpts.length-1)
		} else {
			rangeOpt = fmt.Sprintf("bytes=%d-", downloadOpts.offset)
		}
		getObjectInput.Range = aws.String(rangeOpt)
	}

	_, err = downloader.Download(ctx, output, getObjectInput)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to download from DigitalOcean Spaces: %w", err)
	}
	return nil
}

func (m *digitalOceanManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *digitalOceanManager) UploadReader(ctx context.Context, fileName string, rdr io.Reader) (UploadedFile, error) {
	if m.config.Bucket == "" {
		return UploadedFile{}, errors.New("no storage bucket configured to uploader")
	}

	uploadInput := &s3.PutObjectInput{
		ACL:    types.ObjectCannedACLBucketOwnerFullControl,
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(fileName),
		Body:   rdr,
	}

	client, err := m.getClient(ctx)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("digitalocean client: %w", err)
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
		return UploadedFile{}, fmt.Errorf("failed to upload to DigitalOcean Spaces: %w", err)
	}

	return UploadedFile{Location: output.Location, ObjectName: fileName}, nil
}

func (m *digitalOceanManager) Delete(ctx context.Context, keys []string) error {
	client, err := m.getClient(ctx)
	if err != nil {
		return fmt.Errorf("digitalocean client: %w", err)
	}

	if len(keys) == 0 {
		return nil
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
				m.logger.Errorn("Error while deleting DigitalOcean Spaces objects",
					logger.NewStringField("error_code", apiErr.ErrorCode()),
					obskit.Error(err),
				)
			} else {
				m.logger.Errorn("Error while deleting DigitalOcean Spaces objects", obskit.Error(err))
			}
			return fmt.Errorf("failed to delete DigitalOcean Spaces objects: %w", err)
		}
	}
	return nil
}

func (m *digitalOceanManager) Prefix() string {
	return m.config.Prefix
}

func (m *digitalOceanManager) GetObjectNameFromLocation(location string) (string, error) {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("failed to parse location URL: %w", err)
	}
	trimmedURL := strings.TrimLeft(parsedUrl.Path, "/")
	if (m.config.ForcePathStyle != nil && *m.config.ForcePathStyle) ||
		(!strings.Contains(parsedUrl.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket)), nil
	}
	return trimmedURL, nil
}

func (m *digitalOceanManager) GetDownloadKeyFromFileLocation(location string) string {
	parsedURL, err := url.Parse(location)
	if err != nil {
		m.logger.Errorn("error while parsing location url", obskit.Error(err))
		return ""
	}
	trimmedURL := strings.TrimLeft(parsedURL.Path, "/")
	if (m.config.ForcePathStyle != nil && *m.config.ForcePathStyle) ||
		(!strings.Contains(parsedURL.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket))
	}
	return trimmedURL
}

func getSpacesLocation(location string) (region string) {
	r, _ := regexp.Compile(`\.*.*\\.digitaloceanspaces\\.com`)
	subLocation := r.FindString(location)
	regionTokens := strings.Split(subLocation, ".")
	if len(regionTokens) == 3 {
		region = regionTokens[0]
	}
	return region
}

type digitalOceanListSession struct {
	*baseListSession
	manager *digitalOceanManager

	continuationToken *string
	isTruncated       bool
}

func (l *digitalOceanListSession) Next() ([]*FileInfo, error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Debugn("Manager is truncated, so returning here", logger.NewBoolField("isTruncated", l.isTruncated))
		return nil, nil
	}

	fileObjects := make([]*FileInfo, 0)

	client, err := manager.getClient(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting Digital Ocean Spaces session: %w", err)
	}

	listObjectsInput := s3.ListObjectsV2Input{
		Bucket:            aws.String(manager.config.Bucket),
		Prefix:            aws.String(l.prefix),
		MaxKeys:           aws.Int32(int32(l.maxItems)),
		ContinuationToken: l.continuationToken,
	}
	if l.startAfter != "" {
		listObjectsInput.StartAfter = aws.String(l.startAfter)
	}

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	resp, err := client.ListObjectsV2(ctx, &listObjectsInput)
	if err != nil {
		manager.logger.Errorn("Error while listing Digital Ocean Spaces objects", obskit.Error(err))
		return nil, err
	}
	if resp.IsTruncated != nil {
		l.isTruncated = *resp.IsTruncated
	}
	l.isTruncated = *resp.IsTruncated
	l.continuationToken = resp.NextContinuationToken
	for _, item := range resp.Contents {
		fileObjects = append(fileObjects, &FileInfo{*item.Key, *item.LastModified})
	}
	return fileObjects, nil
}
