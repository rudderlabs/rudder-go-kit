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

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

type MinioConfig struct {
	Bucket          string
	Prefix          string
	EndPoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

// NewMinioManager creates a new file manager for minio
func NewMinioManager(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (*MinioManager, error) {
	return &MinioManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config: minioConfig(config),
	}, nil
}

func (m *MinioManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &minioListSession{
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
func (m *MinioManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	minioClient, err := m.getClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	getObjectOpts := minio.GetObjectOptions{}
	if downloadOpts.isRangeRequest {
		end := int64(0)
		if downloadOpts.length > 0 {
			end = downloadOpts.offset + downloadOpts.length - 1
		}
		if err := getObjectOpts.SetRange(downloadOpts.offset, end); err != nil {
			return fmt.Errorf("setting range for minio download: %w", err)
		}
	}

	if file, ok := output.(*os.File); ok {
		return minioClient.FGetObject(ctx, m.config.Bucket, key, file.Name(), getObjectOpts)
	}

	// Use GetObject with WriterAt interface instead of FGetObject
	obj, err := minioClient.GetObject(ctx, m.config.Bucket, key, getObjectOpts)
	if err != nil {
		return err
	}
	defer func() { _ = obj.Close() }()

	writer := &writerAtAdapter{w: output}
	_, err = io.Copy(writer, obj)
	return err
}

func (m *MinioManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *MinioManager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	if m.config.Bucket == "" {
		return UploadedFile{}, errors.New("no storage bucket configured to uploader")
	}

	minioClient, err := m.getClient()
	if err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, m.config.Bucket)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		if err = minioClient.MakeBucket(ctx, m.config.Bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			return UploadedFile{}, fmt.Errorf("creating bucket: %w", err)
		}
	}

	// Check if output is *os.File to use FGetObject
	if file, ok := rdr.(*os.File); ok {
		_, err = minioClient.FPutObject(ctx, m.config.Bucket, objName, file.Name(), minio.PutObjectOptions{})
	} else {
		_, err = minioClient.PutObject(ctx, m.config.Bucket, objName, rdr, -1, minio.PutObjectOptions{})
	}
	if err != nil {
		return UploadedFile{}, err
	}

	return UploadedFile{Location: m.objectUrl(objName), ObjectName: objName}, nil
}

func (m *MinioManager) Delete(ctx context.Context, keys []string) (err error) {
	objectChannel := make(chan minio.ObjectInfo, len(keys))
	for _, key := range keys {
		objectChannel <- minio.ObjectInfo{Key: key}
	}
	close(objectChannel)

	minioClient, err := m.getClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	tmp := <-minioClient.RemoveObjects(ctx, m.config.Bucket, objectChannel, minio.RemoveObjectsOptions{})
	return tmp.Err
}

func (m *MinioManager) Prefix() string {
	return m.config.Prefix
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://minio-endpoint/bucket-name/key1 - >> key1
	http://minio-endpoint/bucket-name/key2 - >> key2
*/
func (m *MinioManager) GetObjectNameFromLocation(location string) (string, error) {
	var baseURL string
	if m.config.UseSSL {
		baseURL += "https://"
	} else {
		baseURL += "http://"
	}
	baseURL += m.config.EndPoint + "/"
	baseURL += m.config.Bucket + "/"
	return location[len(baseURL):], nil
}

func (m *MinioManager) GetDownloadKeyFromFileLocation(location string) string {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		fmt.Println("error while parsing location url: ", err)
	}
	trimedUrl := strings.TrimLeft(parsedUrl.Path, "/")
	return strings.TrimPrefix(trimedUrl, fmt.Sprintf(`%s/`, m.config.Bucket))
}

func (m *MinioManager) objectUrl(objectName string) string {
	protocol := "http"
	if m.config.UseSSL {
		protocol = "https"
	}
	return protocol + "://" + m.config.EndPoint + "/" + m.config.Bucket + "/" + objectName
}

func (m *MinioManager) getClient() (*minio.Client, error) {
	m.clientOnce.Do(func() {
		m.client, m.clientErr = minio.New(m.config.EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(m.config.AccessKeyID, m.config.SecretAccessKey, ""),
			Secure: m.config.UseSSL,
		})
		if m.clientErr != nil {
			m.client = &minio.Client{}
		}
	})

	return m.client, m.clientErr
}

type MinioManager struct {
	*baseManager
	config *MinioConfig

	client     *minio.Client
	clientErr  error
	clientOnce sync.Once
}

func minioConfig(config map[string]interface{}) *MinioConfig {
	var bucketName, prefix, endPoint, accessKeyID, secretAccessKey string
	var useSSL, ok bool
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
	if config["secretAccessKey"] != nil {
		tmp, ok := config["secretAccessKey"].(string)
		if ok {
			secretAccessKey = tmp
		}
	}
	if config["useSSL"] != nil {
		if useSSL, ok = config["useSSL"].(bool); !ok {
			useSSL = false
		}
	}

	return &MinioConfig{
		Bucket:          bucketName,
		Prefix:          prefix,
		EndPoint:        endPoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		UseSSL:          useSSL,
	}
}

type minioListSession struct {
	*baseListSession
	manager *MinioManager

	continuationToken string
	isTruncated       bool
}

func (l *minioListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Debugn("Manager is truncated, so returning here", logger.NewBoolField("isTruncated", l.isTruncated))
		return
	}
	fileObjects = make([]*FileInfo, 0)

	// Created minio core
	core, err := minio.NewCore(manager.config.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(manager.config.AccessKeyID, manager.config.SecretAccessKey, ""),
		Secure: manager.config.UseSSL,
	})
	if err != nil {
		return
	}

	// List the Objects in the bucket
	result, err := core.ListObjectsV2(manager.config.Bucket, l.prefix, l.startAfter, l.continuationToken, "", int(l.maxItems))
	if err != nil {
		return
	}

	for idx := range result.Contents {
		fileObjects = append(fileObjects, &FileInfo{result.Contents[idx].Key, result.Contents[idx].LastModified})
	}
	l.isTruncated = result.IsTruncated
	l.continuationToken = result.NextContinuationToken
	return
}
