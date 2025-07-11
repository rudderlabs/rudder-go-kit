package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awsS3Manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/async"
	"github.com/rudderlabs/rudder-go-kit/awsutil"
	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type S3Config struct {
	Bucket string `mapstructure:"bucketName"`
	Prefix string `mapstructure:"Prefix"`
	// Region           *string `mapstructure:"region"`
	Endpoint         *string `mapstructure:"endpoint"`
	S3ForcePathStyle *bool   `mapstructure:"s3ForcePathStyle"`
	DisableSSL       *bool   `mapstructure:"disableSSL"`
	EnableSSE        bool    `mapstructure:"enableSSE"`
	RegionHint       string  `mapstructure:"regionHint"`
}

// newS3ManagerV1 creates a new file manager for S3
func newS3ManagerV1(
	kitconfig *kitconfig.Config, config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration,
) (*s3ManagerV1, error) {
	var s3Config S3Config
	if err := mapstructure.Decode(config, &s3Config); err != nil {
		return nil, err
	}

	sessionConfig, err := awsutil.NewSimpleSessionConfig(config, s3.ServiceName)
	if err != nil {
		return nil, err
	}

	s3Config.RegionHint = kitconfig.GetString("AWS_S3_REGION_HINT", "us-east-1")

	return &s3ManagerV1{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:        &s3Config,
		sessionConfig: sessionConfig,
	}, nil
}

func (m *s3ManagerV1) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &s3ListSession{
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

// Download retrieves an object with the given key and writes it to the provided writer.
// Pass *os.File as output to write the downloaded file on disk.
func (m *s3ManagerV1) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)

	sess, err := m.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error starting S3 session: %w", err)
	}

	downloader := awsS3Manager.NewDownloader(sess)

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
	_, err = downloader.DownloadWithContext(ctx, output, getObjectInput)
	if err != nil {
		if codeErr, ok := err.(codeError); ok && codeErr.Code() == "NoSuchKey" {
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

// Upload uploads a file to S3
func (m *s3ManagerV1) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

// UploadReader uploads an object to S3 using the provided reader and object name.
// It supports server-side encryption if enabled in the configuration.
// Returns an UploadedFile containing the file's location and object name, or an error.
func (m *s3ManagerV1) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	uploadInput := &awsS3Manager.UploadInput{
		ACL:    aws.String("bucket-owner-full-control"),
		Bucket: aws.String(m.config.Bucket),
		Key:    aws.String(objName),
		Body:   rdr,
	}
	if m.config.EnableSSE {
		uploadInput.ServerSideEncryption = aws.String("AES256")
	}

	uploadSession, err := m.GetSession(ctx)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("error starting S3 session: %w", err)
	}
	s3manager := awsS3Manager.NewUploader(uploadSession)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	output, err := s3manager.UploadWithContext(ctx, uploadInput)
	if err != nil {
		if codeErr, ok := err.(codeError); ok && codeErr.Code() == "MissingRegion" {
			err = fmt.Errorf("bucket %q not found", m.config.Bucket)
		}
		return UploadedFile{}, fmt.Errorf("error uploading file to S3: %w", err)
	}

	return UploadedFile{Location: output.Location, ObjectName: objName}, err
}

func (m *s3ManagerV1) Delete(ctx context.Context, keys []string) (err error) {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error starting S3 session: %w", err)
	}

	var objects []*s3.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, &s3.ObjectIdentifier{Key: aws.String(key)})
	}

	svc := s3.New(sess)

	batchSize := 1000 // max accepted by DeleteObjects API
	chunks := lo.Chunk(objects, batchSize)
	for _, chunk := range chunks {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(m.config.Bucket),
			Delete: &s3.Delete{
				Objects: chunk,
			},
		}

		deleteCtx, cancel := context.WithTimeout(ctx, m.getTimeout())
		_, err := svc.DeleteObjectsWithContext(deleteCtx, input)
		cancel()

		if err != nil {
			if codeErr, ok := err.(codeError); ok {
				m.logger.Errorn("Error while deleting S3 objects", logger.NewStringField("code", codeErr.Code()), obskit.Error(err))
			} else {
				m.logger.Errorn("Error while deleting S3 objects", obskit.Error(err))
			}
			return err
		}
	}
	return nil
}

func (m *s3ManagerV1) Prefix() string {
	return m.config.Prefix
}

func (m *s3ManagerV1) Bucket() string {
	return m.config.Bucket
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://bucket-name.s3.amazonaws.com/key - >> key
*/
func (m *s3ManagerV1) GetObjectNameFromLocation(location string) (string, error) {
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

func (m *s3ManagerV1) GetDownloadKeyFromFileLocation(location string) string {
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

func (m *s3ManagerV1) GetSession(ctx context.Context) (*session.Session, error) {
	m.sessionMu.Lock()
	defer m.sessionMu.Unlock()

	if m.session != nil {
		return m.session, nil
	}

	if m.config.Bucket == "" {
		return nil, errors.New("no storage bucket configured to downloader")
	}

	// if m.config.Region == nil || *m.config.Region == "" {
	// 	getRegionSession, err := session.NewSession()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	// 	defer cancel()

	// 	region, err := awsS3Manager.GetBucketRegion(ctx, getRegionSession, m.config.Bucket, m.config.RegionHint)
	// 	if err != nil {
	// 		m.logger.Errorn("Failed to fetch AWS region for bucket",
	// 			logger.NewStringField("bucket", m.config.Bucket), obskit.Error(err),
	// 		)
	// 		// Failed to get Region probably due to VPC restrictions
	// 		// Will proceed to try with AccessKeyID and AccessKey
	// 	}
	// 	m.config.Region = aws.String(region)
	// 	m.sessionConfig.Region = region
	// } else {
	// 	m.sessionConfig.Region = *m.config.Region
	// }

	var err error
	m.session, err = awsutil.CreateSession(m.sessionConfig)
	if err != nil {
		return nil, err
	}
	return m.session, err
}

func (m *s3ManagerV1) SelectObjects(ctx context.Context, selectConfig SelectConfig) (<-chan SelectResult, func()) {
	s := async.SingleSender[SelectResult]{}
	ctx, selectResultChan, leave := s.Begin(ctx)

	go func() {
		defer s.Close()
		sess, err := m.GetSession(ctx)
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("error starting S3 session: %w", err)})
			return
		}
		svc := s3.New(sess)

		inputSerialization, outputSerialization, err := createS3SelectSerialization(selectConfig.InputFormat, selectConfig.OutputFormat)
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("error extracting input/output serialization: %w", err)})
			return
		}

		input := &s3.SelectObjectContentInput{
			Bucket:              aws.String(m.Bucket()),
			Key:                 aws.String(selectConfig.Key),
			Expression:          aws.String(selectConfig.SQLExpression),
			ExpressionType:      aws.String(s3.ExpressionTypeSql),
			InputSerialization:  inputSerialization,
			OutputSerialization: outputSerialization,
		}
		selectObject, err := svc.SelectObjectContentWithContext(ctx, input)
		if err != nil {
			s.Send(SelectResult{Error: fmt.Errorf("error selecting object: %w", err)})
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
				case *s3.RecordsEvent:
					s.Send(SelectResult{Data: e.Payload})
				case *s3.EndEvent:
					return
				}
			}
		}
	}()
	return selectResultChan, leave
}

func createS3SelectSerialization(inputFormat SelectObjectInputFormat, outputFormat SelectObjectOutputFormat) (*s3.InputSerialization, *s3.OutputSerialization, error) {
	var inputSerialization *s3.InputSerialization
	switch inputFormat {
	case SelectObjectInputFormatParquet:
		inputSerialization = &s3.InputSerialization{
			Parquet: &s3.ParquetInput{},
		}
	default:
		return nil, nil, fmt.Errorf("invalid input format: %s", inputFormat)
	}

	var outputSerialization *s3.OutputSerialization
	switch outputFormat {
	case SelectObjectOutputFormatCSV:
		outputSerialization = &s3.OutputSerialization{
			CSV: &s3.CSVOutput{
				RecordDelimiter: aws.String("\n"),
				FieldDelimiter:  aws.String(","),
			},
		}
	case SelectObjectOutputFormatJSON:
		outputSerialization = &s3.OutputSerialization{
			JSON: &s3.JSONOutput{},
		}
	default:
		return nil, nil, fmt.Errorf("invalid output format: %s", outputFormat)
	}

	return inputSerialization, outputSerialization, nil
}

type s3ManagerV1 struct {
	*baseManager
	config *S3Config

	sessionConfig *awsutil.SessionConfig
	session       *session.Session
	sessionMu     sync.Mutex
}

func (m *s3ManagerV1) getTimeout() time.Duration {
	if m.timeout > 0 {
		return m.timeout
	}
	if m.defaultTimeout != nil {
		return m.defaultTimeout()
	}
	return defaultTimeout
}

type s3ListSession struct {
	*baseListSession
	manager *s3ManagerV1

	continuationToken *string
	isTruncated       bool
}

func (l *s3ListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Debugn("Manager is truncated, so returning here", logger.NewBoolField("isTruncated", l.isTruncated))
		return
	}
	fileObjects = make([]*FileInfo, 0)

	sess, err := manager.GetSession(l.ctx)
	if err != nil {
		return []*FileInfo{}, fmt.Errorf("error starting S3 session: %w", err)
	}
	// Create S3 service client
	svc := s3.New(sess)
	listObjectsV2Input := s3.ListObjectsV2Input{
		Bucket:  aws.String(manager.config.Bucket),
		Prefix:  aws.String(l.prefix),
		MaxKeys: &l.maxItems,
		// Delimiter: aws.String("/"),
	}
	// startAfter is to resume a paused task.
	if l.startAfter != "" {
		listObjectsV2Input.StartAfter = aws.String(l.startAfter)
	}

	if l.continuationToken != nil {
		listObjectsV2Input.ContinuationToken = l.continuationToken
	}

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	// Get the list of items
	resp, err := svc.ListObjectsV2WithContext(ctx, &listObjectsV2Input)
	if err != nil {
		manager.logger.Errorn("Error while listing S3 objects", obskit.Error(err))
		return
	}
	if resp.IsTruncated != nil {
		l.isTruncated = *resp.IsTruncated
	}
	l.continuationToken = resp.NextContinuationToken
	for _, item := range resp.Contents {
		fileObjects = append(fileObjects, &FileInfo{*item.Key, *item.LastModified})
	}
	return
}

type codeError interface {
	Code() string
}
