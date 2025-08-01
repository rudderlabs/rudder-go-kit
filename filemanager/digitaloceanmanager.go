package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

type DigitalOceanConfig struct {
	Bucket         string
	Prefix         string
	EndPoint       string
	AccessKeyID    string
	AccessKey      string
	Region         *string
	ForcePathStyle *bool
	DisableSSL     *bool
}

// newDigitalOceanManagerV1 creates a new file manager for digital ocean spaces
func newDigitalOceanManagerV1(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) *digitalOceanManagerV1 {
	return &digitalOceanManagerV1{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		Config: digitalOceanConfig(config),
	}
}

func (m *digitalOceanManagerV1) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
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

// Download retrieves an object with the given key and writes it to the provided writer.
// Pass *os.File as output to write the downloaded file on disk.
func (m *digitalOceanManagerV1) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	downloadSession, err := m.getSession()
	if err != nil {
		return fmt.Errorf("error starting Digital Ocean Spaces session: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	downloader := s3manager.NewDownloader(downloadSession)

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(m.Config.Bucket),
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
	return err
}

func (m *digitalOceanManagerV1) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.Config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *digitalOceanManagerV1) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	if m.Config.Bucket == "" {
		return UploadedFile{}, errors.New("no storage bucket configured to uploader")
	}

	uploadInput := &s3manager.UploadInput{
		ACL:    aws.String("bucket-owner-full-control"),
		Bucket: aws.String(m.Config.Bucket),
		Key:    aws.String(objName),
		Body:   rdr,
	}
	uploadSession, err := m.getSession()
	if err != nil {
		return UploadedFile{}, fmt.Errorf("error starting Digital Ocean Spaces session: %w", err)
	}
	doManager := s3manager.NewUploader(uploadSession)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	output, err := doManager.UploadWithContext(ctx, uploadInput)
	if err != nil {
		var awsError awserr.Error
		if errors.As(err, &awsError) && awsError.Code() == "MissingRegion" {
			err = fmt.Errorf("bucket %q not found", m.Config.Bucket)
		}
		return UploadedFile{}, err
	}

	return UploadedFile{Location: output.Location, ObjectName: objName}, err
}

func (m *digitalOceanManagerV1) Delete(ctx context.Context, keys []string) error {
	sess, err := m.getSession()
	if err != nil {
		return fmt.Errorf("error starting Digital Ocean Spaces session: %w", err)
	}

	objects := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = &s3.ObjectIdentifier{Key: aws.String(key)}
	}

	svc := s3.New(sess)

	batchSize := 1000 // max accepted by DeleteObjects API
	chunks := lo.Chunk(objects, batchSize)
	for _, chunk := range chunks {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(m.Config.Bucket),
			Delete: &s3.Delete{
				Objects: chunk,
			},
		}

		_ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
		_, err := svc.DeleteObjectsWithContext(_ctx, input)
		if err != nil {
			var errCode string
			var awsErr awserr.Error
			if errors.As(err, &awsErr) {
				errCode = awsErr.Code()
			}
			m.logger.Errorn("Error while deleting digital ocean spaces objects", obskit.Error(err),
				logger.NewStringField("code", errCode),
			)
			cancel()
			return err
		}
		cancel()
	}
	return nil
}

func (m *digitalOceanManagerV1) Prefix() string {
	return m.Config.Prefix
}

func (m *digitalOceanManagerV1) GetDownloadKeyFromFileLocation(location string) string {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		fmt.Println("error while parsing location url: ", err)
	}
	trimmedURL := strings.TrimLeft(parsedUrl.Path, "/")
	if (m.Config.ForcePathStyle != nil && *m.Config.ForcePathStyle) || (!strings.Contains(parsedUrl.Host, m.Config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.Config.Bucket))
	}
	return trimmedURL
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://rudder.sgp1.digitaloceanspaces.com/key - >> key
*/
func (m *digitalOceanManagerV1) GetObjectNameFromLocation(location string) (string, error) {
	parsedURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	trimmedURL := strings.TrimLeft(parsedURL.Path, "/")
	if (m.Config.ForcePathStyle != nil && *m.Config.ForcePathStyle) || (!strings.Contains(parsedURL.Host, m.Config.Bucket)) {
		return strings.TrimPrefix(trimmedURL, fmt.Sprintf(`%s/`, m.Config.Bucket)), nil
	}
	return trimmedURL, nil
}

func (m *digitalOceanManagerV1) getSession() (*session.Session, error) {
	var region string
	if m.Config.Region != nil {
		region = *m.Config.Region
	} else {
		region = getSpacesLocation(m.Config.EndPoint)
	}
	return session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(m.Config.AccessKeyID, m.Config.AccessKey, ""),
		Endpoint:         aws.String(m.Config.EndPoint),
		DisableSSL:       m.Config.DisableSSL,
		S3ForcePathStyle: m.Config.ForcePathStyle,
	})
}

func getSpacesLocation(location string) (region string) {
	r, _ := regexp.Compile(`\.*.*\.digitaloceanspaces\.com`) // skipcq: GO-S1009
	subLocation := r.FindString(location)
	regionTokens := strings.Split(subLocation, ".")
	if len(regionTokens) == 3 {
		region = regionTokens[0]
	}
	return region
}

type digitalOceanManagerV1 struct {
	*baseManager
	Config *DigitalOceanConfig
}

func digitalOceanConfig(config map[string]interface{}) *DigitalOceanConfig {
	var bucketName, prefix, endPoint, accessKeyID, accessKey string
	var region *string
	var forcePathStyle, disableSSL *bool
	if config["bucketName"] != nil {
		tmp, ok := config["bucketName"].(string)
		if ok {
			bucketName = tmp
		}
	}
	if config["prefix"] != nil {
		tmp, ok := config["prefix"].(string)
		if ok {
			prefix = tmp
		}
	}
	if config["endPoint"] != nil {
		tmp, ok := config["endPoint"].(string)
		if ok {
			endPoint = tmp
		}
	}
	if config["accessKeyID"] != nil {
		tmp, ok := config["accessKeyID"].(string)
		if ok {
			accessKeyID = tmp
		}
	}
	if config["accessKey"] != nil {
		tmp, ok := config["accessKey"].(string)
		if ok {
			accessKey = tmp
		}
	}
	if config["region"] != nil {
		tmp, ok := config["region"].(string)
		if ok {
			region = &tmp
		}
	}
	if config["forcePathStyle"] != nil {
		tmp, ok := config["forcePathStyle"].(bool)
		if ok {
			forcePathStyle = &tmp
		}
	}
	if config["disableSSL"] != nil {
		tmp, ok := config["disableSSL"].(bool)
		if ok {
			disableSSL = &tmp
		}
	}
	return &DigitalOceanConfig{
		Bucket:         bucketName,
		EndPoint:       endPoint,
		Prefix:         prefix,
		AccessKeyID:    accessKeyID,
		AccessKey:      accessKey,
		Region:         region,
		ForcePathStyle: forcePathStyle,
		DisableSSL:     disableSSL,
	}
}

type digitalOceanListSession struct {
	*baseListSession
	manager *digitalOceanManagerV1

	continuationToken *string
	isTruncated       bool
}

func (l *digitalOceanListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Debugn("Manager is truncated, so returning here", logger.NewBoolField("isTruncated", l.isTruncated))
		return
	}
	fileObjects = make([]*FileInfo, 0)

	sess, err := manager.getSession()
	if err != nil {
		return []*FileInfo{}, fmt.Errorf("error starting Digital Ocean Spaces session: %w", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	listObjectsV2Input := s3.ListObjectsV2Input{
		Bucket:  aws.String(manager.Config.Bucket),
		Prefix:  aws.String(l.prefix),
		MaxKeys: &l.maxItems,
	}
	// startAfter is to resume a paused task.
	if l.startAfter != "" {
		listObjectsV2Input.StartAfter = aws.String(l.startAfter)
	}
	if l.continuationToken != nil {
		listObjectsV2Input.ContinuationToken = l.continuationToken
	}

	// Get the list of items
	resp, err := svc.ListObjectsV2WithContext(ctx, &listObjectsV2Input)
	if err != nil {
		manager.logger.Errorn("Error while listing Digital Ocean Spaces objects", obskit.Error(err))
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
