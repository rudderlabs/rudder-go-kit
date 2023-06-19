package filemanager

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awsS3Manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/awsutil"
	appConfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type S3Config struct {
	Bucket           string  `mapstructure:"bucketName"`
	Prefix           string  `mapstructure:"Prefix"`
	Region           *string `mapstructure:"region"`
	Endpoint         *string `mapstructure:"endpoint"`
	S3ForcePathStyle *bool   `mapstructure:"s3ForcePathStyle"`
	DisableSSL       *bool   `mapstructure:"disableSSL"`
	EnableSSE        bool    `mapstructure:"enableSSE"`
	RegionHint       string  `mapstructure:"regionHint"`
	UseGlue          bool    `mapstructure:"useGlue"`
}

// NewS3Manager creates a new file manager for S3
func NewS3Manager(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (*s3Manager, error) {
	var s3Config S3Config
	if err := mapstructure.Decode(config, &s3Config); err != nil {
		return nil, err
	}
	regionHint := appConfig.GetString("AWS_S3_REGION_HINT", "us-east-1")
	s3Config.RegionHint = regionHint
	sessionConfig, err := awsutil.NewSimpleSessionConfig(config, s3.ServiceName)
	if err != nil {
		return nil, err
	}
	return &s3Manager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:        &s3Config,
		sessionConfig: sessionConfig,
	}, nil
}

func (manager *s3Manager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &s3ListSession{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager:     manager,
		isTruncated: true,
	}
}

// Download downloads a file from S3
func (manager *s3Manager) Download(ctx context.Context, output *os.File, key string) error {
	sess, err := manager.getSession(ctx)
	if err != nil {
		return fmt.Errorf("error starting S3 session: %w", err)
	}

	downloader := awsS3Manager.NewDownloader(sess)

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	_, err = downloader.DownloadWithContext(ctx, output,
		&s3.GetObjectInput{
			Bucket: aws.String(manager.config.Bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == ErrKeyNotFound.Error() {
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

// Upload uploads a file to S3
func (manager *s3Manager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	fileName := path.Join(manager.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))

	uploadInput := &awsS3Manager.UploadInput{
		ACL:    aws.String("bucket-owner-full-control"),
		Bucket: aws.String(manager.config.Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	}
	if manager.config.EnableSSE {
		uploadInput.ServerSideEncryption = aws.String("AES256")
	}

	uploadSession, err := manager.getSession(ctx)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("error starting S3 session: %w", err)
	}
	s3manager := awsS3Manager.NewUploader(uploadSession)

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	output, err := s3manager.UploadWithContext(ctx, uploadInput)
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok && awsError.Code() == "MissingRegion" {
			err = fmt.Errorf(fmt.Sprintf(`Bucket '%s' not found.`, manager.config.Bucket))
		}
		return UploadedFile{}, err
	}

	return UploadedFile{Location: output.Location, ObjectName: fileName}, err
}

func (manager *s3Manager) Delete(ctx context.Context, keys []string) (err error) {
	sess, err := manager.getSession(ctx)
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
			Bucket: aws.String(manager.config.Bucket),
			Delete: &s3.Delete{
				Objects: chunk,
			},
		}

		_ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
		defer cancel()

		_, err := svc.DeleteObjectsWithContext(_ctx, input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				manager.logger.Errorf(`Error while deleting S3 objects: %v, error code: %v`, aerr.Error(), aerr.Code())
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				manager.logger.Errorf(`Error while deleting S3 objects: %v`, aerr.Error())
			}
			return err
		}
	}
	return nil
}

func (manager *s3Manager) Prefix() string {
	return manager.config.Prefix
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://bucket-name.s3.amazonaws.com/key - >> key
*/
func (manager *s3Manager) GetObjectNameFromLocation(location string) (string, error) {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	trimedUrl := strings.TrimLeft(parsedUrl.Path, "/")
	if (manager.config.S3ForcePathStyle != nil && *manager.config.S3ForcePathStyle) || (!strings.Contains(parsedUrl.Host, manager.config.Bucket)) {
		return strings.TrimPrefix(trimedUrl, fmt.Sprintf(`%s/`, manager.config.Bucket)), nil
	}
	return trimedUrl, nil
}

func (manager *s3Manager) GetDownloadKeyFromFileLocation(location string) string {
	parsedURL, err := url.Parse(location)
	if err != nil {
		fmt.Println("error while parsing location url: ", err)
	}
	trimmedURL := strings.TrimLeft(parsedURL.Path, "/")
	if (manager.config.S3ForcePathStyle != nil && *manager.config.S3ForcePathStyle) || (!strings.Contains(parsedURL.Host, manager.config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, manager.config.Bucket))
	}
	return trimmedURL
}

func (manager *s3Manager) getSession(ctx context.Context) (*session.Session, error) {
	if manager.session != nil {
		return manager.session, nil
	}

	if manager.config.Bucket == "" {
		return nil, errors.New("no storage bucket configured to downloader")
	}
	if !manager.config.UseGlue || manager.config.Region == nil {
		getRegionSession, err := session.NewSession()
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
		defer cancel()

		region, err := awsS3Manager.GetBucketRegion(ctx, getRegionSession, manager.config.Bucket, manager.config.RegionHint)
		if err != nil {
			manager.logger.Errorf("Failed to fetch AWS region for bucket %s. Error %v", manager.config.Bucket, err)
			/// Failed to Get Region probably due to VPC restrictions, Will proceed to try with AccessKeyID and AccessKey
		}
		manager.config.Region = aws.String(region)
		manager.sessionConfig.Region = region
	}

	var err error
	manager.session, err = awsutil.CreateSession(manager.sessionConfig)
	if err != nil {
		return nil, err
	}
	return manager.session, err
}

type s3Manager struct {
	*baseManager
	config *S3Config

	sessionConfig *awsutil.SessionConfig
	session       *session.Session
}

func (manager *s3Manager) getTimeout() time.Duration {
	if manager.timeout > 0 {
		return manager.timeout
	}
	if manager.defaultTimeout != nil {
		return manager.defaultTimeout()
	}
	return defaultTimeout
}

type s3ListSession struct {
	*baseListSession
	manager *s3Manager

	continuationToken *string
	isTruncated       bool
}

func (l *s3ListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Infof("Manager is truncated: %v so returning here", l.isTruncated)
		return
	}
	fileObjects = make([]*FileInfo, 0)

	sess, err := manager.getSession(l.ctx)
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
		manager.logger.Errorf("Error while listing S3 objects: %v", err)
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
