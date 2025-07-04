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

	"github.com/rudderlabs/rudder-go-kit/async"
	awsutil "github.com/rudderlabs/rudder-go-kit/awsutil_v2"
	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

const ServiceName = "s3"

// s3ManagerV2 manages S3 file operations using AWS SDK v2.
type s3ManagerV2 struct {
	*baseManager
	config *S3Config

	sessionConfig *awsutil.SessionConfig
	client        *s3.Client
	clientMu      sync.Mutex
}

// newS3ManagerV2 creates a new file manager for S3 using v2 AWS SDK.
func newS3ManagerV2(
	kitconfig *kitconfig.Config, config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration,
) (*s3ManagerV2, error) {
	var s3Config S3Config
	if err := mapstructure.Decode(config, &s3Config); err != nil {
		return nil, fmt.Errorf("failed to decode S3 config: %w", err)
	}

	sessionConfig, err := awsutil.NewSimpleSessionConfig(config, ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session config: %w", err)
	}

	s3Config.RegionHint = kitconfig.GetString("AWS_S3_REGION_HINT", "us-east-1")

	return &s3ManagerV2{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:        &s3Config,
		sessionConfig: sessionConfig,
	}, nil
}

// ListFilesWithPrefix returns a session for listing files with the given prefix.
func (m *s3ManagerV2) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
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

// Download downloads a file from S3 to the provided io.WriterAt.
func (m *s3ManagerV2) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	client, err := m.getClient(ctx)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
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
	if err == nil {
		return nil
	}
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		return ErrKeyNotFound
	}
	return fmt.Errorf("failed to download from S3: %w", err)
}

// Upload uploads a file to S3 and returns the uploaded file info.
func (m *s3ManagerV2) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	fileName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, fileName, file)
}

// UploadReader uploads data from an io.Reader to S3 with the given object name.
func (m *s3ManagerV2) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	if objName == "" {
		return UploadedFile{}, errors.New("object name cannot be empty")
	}

	uploadInput := &s3.PutObjectInput{
		ACL:    types.ObjectCannedACLBucketOwnerFullControl,
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(objName),
		Body:   rdr,
	}
	m.logger.Info("Uploading file to S3", logger.NewStringField("objName", objName))
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
	if err == nil {
		return UploadedFile{Location: output.Location, ObjectName: objName}, nil
	}
	var regionError *aws.MissingRegionError
	if errors.As(err, &regionError) {
		err = fmt.Errorf(`missing region for bucket %q: %w`, m.config.Bucket, regionError)
	}
	return UploadedFile{}, err
}

// Delete removes the specified keys from S3.
func (m *s3ManagerV2) Delete(ctx context.Context, keys []string) error {
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
				m.logger.Errorn("Error while deleting S3 objects",
					obskit.Error(err), logger.NewStringField("error_code", apiErr.ErrorCode()),
				)
			} else {
				m.logger.Errorn("Error while deleting S3 objects", obskit.Error(err))
			}
			return fmt.Errorf("failed to delete S3 objects: %w", err)
		}
	}
	return nil
}

// Prefix returns the configured S3 prefix.
func (m *s3ManagerV2) Prefix() string {
	return m.config.Prefix
}

func (m *s3ManagerV2) Bucket() string {
	return m.config.Bucket
}

// GetObjectNameFromLocation extracts the object key from the S3 object location URL.
// Example: https://bucket-name.s3.amazonaws.com/key -> key
func (m *s3ManagerV2) GetObjectNameFromLocation(location string) (string, error) {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("failed to parse location URL: %w", err)
	}
	trimmedURL := strings.TrimLeft(parsedUrl.Path, "/")
	if (m.config.S3ForcePathStyle != nil && *m.config.S3ForcePathStyle) ||
		(!strings.Contains(parsedUrl.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket)), nil
	}
	return trimmedURL, nil
}

// GetDownloadKeyFromFileLocation extracts the S3 key from a file location URL.
func (m *s3ManagerV2) GetDownloadKeyFromFileLocation(location string) string {
	parsedURL, err := url.Parse(location)
	if err != nil {
		m.logger.Errorn("error while parsing location url", obskit.Error(err))
		return ""
	}
	trimmedURL := strings.TrimLeft(parsedURL.Path, "/")
	if (m.config.S3ForcePathStyle != nil && *m.config.S3ForcePathStyle) ||
		(!strings.Contains(parsedURL.Host, m.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.config.Bucket))
	}
	return trimmedURL
}

// getClient returns a cached S3 client or creates a new one if needed.
func (m *s3ManagerV2) getClient(ctx context.Context) (*s3.Client, error) {
	m.clientMu.Lock()
	defer m.clientMu.Unlock()
	json, _ := jsonrs.Marshal(m.config)
	m.logger.Infon("S3 config", logger.NewStringField("config", string(json)))
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
			return nil, fmt.Errorf("failed to get bucket region: %w", err)
		}
		m.config.Region = aws.String(region)
		m.sessionConfig.Region = region
	}

	cnf, err := awsutil.CreateAWSConfig(ctx, m.sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	client := s3.NewFromConfig(cnf, func(o *s3.Options) {
		if m.config.Endpoint != nil && *m.config.Endpoint != "" {
			o.BaseEndpoint = aws.String(*m.config.Endpoint)
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

	return m.client, nil
}

// getTimeout returns the configured timeout for S3 operations.
func (m *s3ManagerV2) getTimeout() time.Duration {
	if m.timeout > 0 {
		return m.timeout
	}
	if m.defaultTimeout != nil {
		return m.defaultTimeout()
	}
	return defaultTimeout
}

// s3ListSessionV2 implements ListSession for S3 using AWS SDK v2.
type s3ListSessionV2 struct {
	*baseListSession
	manager *s3ManagerV2

	continuationToken *string
	isTruncated       bool
}

// Next returns the next batch of file objects from S3.
func (l *s3ListSessionV2) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Debugn("Manager is truncated: returning here", logger.NewBoolField("isTruncated", l.isTruncated))
		return nil, nil
	}
	fileObjects = make([]*FileInfo, 0)

	client, err := manager.getClient(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("s3 client: %w", err)
	}

	listObjectsV2Input := s3.ListObjectsV2Input{
		Bucket:            aws.String(manager.config.Bucket),
		Prefix:            aws.String(l.prefix),
		MaxKeys:           aws.Int32(int32(l.maxItems)),
		ContinuationToken: l.continuationToken,
	}
	if l.startAfter != "" {
		listObjectsV2Input.StartAfter = aws.String(l.startAfter)
	}

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	resp, err := client.ListObjectsV2(ctx, &listObjectsV2Input)
	if err != nil {
		manager.logger.Errorn("Error while listing S3 objects", obskit.Error(err))
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}
	l.isTruncated = *resp.IsTruncated
	l.continuationToken = resp.NextContinuationToken
	for _, item := range resp.Contents {
		fileObjects = append(fileObjects, &FileInfo{*item.Key, *item.LastModified})
	}
	return fileObjects, nil
}

func (m *s3ManagerV2) SelectObjects(ctx context.Context, selectConfig SelectConfig) (<-chan SelectResult, func()) {
	s := async.SingleSender[SelectResult]{}
	ctx, selectResultChan, leave := s.Begin(ctx)

	go func() {
		defer s.Close()
		client, err := m.getClient(ctx)
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("selecting objects: %w", err)})
			return
		}

		inputSerialization, outputSerialization, err := createS3SelectSerializationV2(selectConfig.InputFormat, selectConfig.OutputFormat)
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("error extracting input/output serialization: %w", err)})
			return
		}

		selectObject, err := client.SelectObjectContent(ctx, &s3.SelectObjectContentInput{
			Bucket:              aws.String(m.config.Bucket),
			Key:                 aws.String(selectConfig.Key),
			Expression:          aws.String(selectConfig.SQLExpression),
			ExpressionType:      types.ExpressionTypeSql,
			InputSerialization:  inputSerialization,
			OutputSerialization: outputSerialization,
		})
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("selecting object: %w", err)})
			return
		}

		stream := selectObject.GetStream()
		defer func() {
			if err := stream.Err(); err != nil && ctx.Err() == nil {
				s.Send(SelectResult{Error: err})
			}
			stream.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				s.Send(SelectResult{Error: ctx.Err()})
				return
			case event, ok := <-stream.Events():
				if !ok {
					return
				}
				switch e := event.(type) {
				case *types.SelectObjectContentEventStreamMemberRecords:
					s.Send(SelectResult{Data: e.Value.Payload})
				case *types.SelectObjectContentEventStreamMemberEnd:
					return
				}
			}
		}
	}()
	return selectResultChan, leave
}

func createS3SelectSerializationV2(inputFormat SelectObjectInputFormat, outputFormat SelectObjectOutputFormat) (*types.InputSerialization, *types.OutputSerialization, error) {
	var inputSerialization *types.InputSerialization
	switch inputFormat {
	case SelectObjectInputFormatParquet:
		inputSerialization = &types.InputSerialization{
			Parquet: &types.ParquetInput{},
		}
	default:
		return nil, nil, fmt.Errorf("invalid input format: %s", inputFormat)
	}

	var outputSerialization *types.OutputSerialization
	switch outputFormat {
	case SelectObjectOutputFormatCSV:
		outputSerialization = &types.OutputSerialization{
			CSV: &types.CSVOutput{
				RecordDelimiter: aws.String("\n"),
				FieldDelimiter:  aws.String(","),
			},
		}
	case SelectObjectOutputFormatJSON:
		outputSerialization = &types.OutputSerialization{
			JSON: &types.JSONOutput{},
		}
	default:
		return nil, nil, fmt.Errorf("invalid output format: %s", outputFormat)
	}

	return inputSerialization, outputSerialization, nil
}
