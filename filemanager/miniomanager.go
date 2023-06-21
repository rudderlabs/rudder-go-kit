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
func NewMinioManager(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (*minioManager, error) {
	return &minioManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config: minioConfig(config),
	}, nil
}

func (manager *minioManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &minioListSession{
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

func (manager *minioManager) Download(ctx context.Context, file *os.File, key string) error {
	minioClient, err := manager.getClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	err = minioClient.FGetObject(ctx, manager.config.Bucket, key, file.Name(), minio.GetObjectOptions{})
	return err
}

func (manager *minioManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	if manager.config.Bucket == "" {
		return UploadedFile{}, errors.New("no storage bucket configured to uploader")
	}

	minioClient, err := manager.getClient()
	if err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, manager.config.Bucket)
	if err != nil {
		return UploadedFile{}, fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		if err = minioClient.MakeBucket(ctx, manager.config.Bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			return UploadedFile{}, fmt.Errorf("creating bucket: %w", err)
		}
	}

	fileName := path.Join(manager.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))

	_, err = minioClient.FPutObject(ctx, manager.config.Bucket, fileName, file.Name(), minio.PutObjectOptions{})
	if err != nil {
		return UploadedFile{}, err
	}

	return UploadedFile{Location: manager.objectUrl(fileName), ObjectName: fileName}, nil
}

func (manager *minioManager) Delete(ctx context.Context, keys []string) (err error) {
	objectChannel := make(chan minio.ObjectInfo, len(keys))
	for _, key := range keys {
		objectChannel <- minio.ObjectInfo{Key: key}
	}
	close(objectChannel)

	minioClient, err := manager.getClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	tmp := <-minioClient.RemoveObjects(ctx, manager.config.Bucket, objectChannel, minio.RemoveObjectsOptions{})
	return tmp.Err
}

func (manager *minioManager) Prefix() string {
	return manager.config.Prefix
}

/*
GetObjectNameFromLocation gets the object name/key name from the object location url

	https://minio-endpoint/bucket-name/key1 - >> key1
	http://minio-endpoint/bucket-name/key2 - >> key2
*/
func (manager *minioManager) GetObjectNameFromLocation(location string) (string, error) {
	var baseURL string
	if manager.config.UseSSL {
		baseURL += "https://"
	} else {
		baseURL += "http://"
	}
	baseURL += manager.config.EndPoint + "/"
	baseURL += manager.config.Bucket + "/"
	return location[len(baseURL):], nil
}

func (manager *minioManager) GetDownloadKeyFromFileLocation(location string) string {
	parsedUrl, err := url.Parse(location)
	if err != nil {
		fmt.Println("error while parsing location url: ", err)
	}
	trimedUrl := strings.TrimLeft(parsedUrl.Path, "/")
	return strings.TrimPrefix(trimedUrl, fmt.Sprintf(`%s/`, manager.config.Bucket))
}

func (manager *minioManager) objectUrl(objectName string) string {
	protocol := "http"
	if manager.config.UseSSL {
		protocol = "https"
	}
	return protocol + "://" + manager.config.EndPoint + "/" + manager.config.Bucket + "/" + objectName
}

func (manager *minioManager) getClient() (*minio.Client, error) {
	var err error
	if manager.client == nil {
		manager.client, err = minio.New(manager.config.EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(manager.config.AccessKeyID, manager.config.SecretAccessKey, ""),
			Secure: manager.config.UseSSL,
		})
		if err != nil {
			return &minio.Client{}, err
		}
	}
	return manager.client, nil
}

type minioManager struct {
	*baseManager
	config *MinioConfig

	client *minio.Client
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
	manager *minioManager

	continuationToken string
	isTruncated       bool
}

func (l *minioListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	if !l.isTruncated {
		manager.logger.Infof("Manager is truncated: %v so returning here", l.isTruncated)
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
