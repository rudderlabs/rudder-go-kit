package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/api/iterator"

	"github.com/rudderlabs/rudder-go-kit/googleutil"
	"github.com/rudderlabs/rudder-go-kit/logger"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSConfig struct {
	Bucket         string
	Prefix         string
	Credentials    string
	EndPoint       *string
	ForcePathStyle *bool
	DisableSSL     *bool
	JSONReads      bool
}

// NewGCSManager creates a new file manager for Google Cloud Storage
func NewGCSManager(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (*gcsManager, error) {
	return &gcsManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config: gcsConfig(config),
	}, nil
}

func (manager *gcsManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &gcsListSession{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager: manager,
	}
}

func (manager *gcsManager) Download(ctx context.Context, output *os.File, key string) error {
	client, err := manager.getClient(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	rc, err := client.Bucket(manager.config.Bucket).Object(key).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(output, rc)
	return err
}

func (manager *gcsManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	fileName := path.Join(manager.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))

	client, err := manager.getClient(ctx)
	if err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	obj := client.Bucket(manager.config.Bucket).Object(fileName)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, file); err != nil {
		err = fmt.Errorf("copying file to GCS: %v", err)
		if closeErr := w.Close(); closeErr != nil {
			return UploadedFile{}, fmt.Errorf("closing writer: %q, while: %w", closeErr, err)
		}

		return UploadedFile{}, err
	}

	if err := w.Close(); err != nil {
		return UploadedFile{}, fmt.Errorf("closing writer: %w", err)
	}
	attrs := w.Attrs()

	return UploadedFile{Location: manager.objectURL(attrs), ObjectName: fileName}, err
}

func (manager *gcsManager) Delete(ctx context.Context, keys []string) (err error) {
	client, err := manager.getClient(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()

	for _, key := range keys {
		if err := client.Bucket(manager.config.Bucket).Object(key).Delete(ctx); err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
			return err
		}
	}
	return
}

func (manager *gcsManager) Prefix() string {
	return manager.config.Prefix
}

func (manager *gcsManager) GetObjectNameFromLocation(location string) (string, error) {
	splitStr := strings.Split(location, manager.config.Bucket)
	object := strings.TrimLeft(splitStr[len(splitStr)-1], "/")
	return object, nil
}

func (manager *gcsManager) GetDownloadKeyFromFileLocation(location string) string {
	splitStr := strings.Split(location, manager.config.Bucket)
	key := strings.TrimLeft(splitStr[len(splitStr)-1], "/")
	return key
}

func (manager *gcsManager) objectURL(objAttrs *storage.ObjectAttrs) string {
	if manager.config.EndPoint != nil && *manager.config.EndPoint != "" {
		endpoint := strings.TrimSuffix(*manager.config.EndPoint, "/")
		return fmt.Sprintf("%s/%s/%s", endpoint, objAttrs.Bucket, objAttrs.Name)
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", objAttrs.Bucket, objAttrs.Name)
}

func (manager *gcsManager) getClient(ctx context.Context) (*storage.Client, error) {
	var err error

	ctx, cancel := context.WithTimeout(ctx, manager.getTimeout())
	defer cancel()
	if manager.client != nil {
		return manager.client, err
	}
	options := []option.ClientOption{}

	if manager.config.EndPoint != nil && *manager.config.EndPoint != "" {
		options = append(options, option.WithEndpoint(*manager.config.EndPoint))
	}
	if !googleutil.ShouldSkipCredentialsInit(manager.config.Credentials) {
		if err = googleutil.CompatibleGoogleCredentialsJSON([]byte(manager.config.Credentials)); err != nil {
			return manager.client, err
		}
		options = append(options, option.WithCredentialsJSON([]byte(manager.config.Credentials)))
	}
	if manager.config.JSONReads {
		options = append(options, storage.WithJSONReads())
	}

	manager.client, err = storage.NewClient(ctx, options...)
	return manager.client, err
}

type gcsManager struct {
	*baseManager
	config *GCSConfig

	client *storage.Client
}

func gcsConfig(config map[string]interface{}) *GCSConfig {
	var bucketName, prefix, credentials string
	var endPoint *string
	var forcePathStyle, disableSSL *bool
	var jsonReads bool

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
	if config["credentials"] != nil {
		tmp, ok := config["credentials"].(string)
		if ok {
			credentials = tmp
		}
	}
	if config["endPoint"] != nil {
		tmp, ok := config["endPoint"].(string)
		if ok {
			endPoint = &tmp
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
	if config["jsonReads"] != nil {
		tmp, ok := config["jsonReads"].(bool)
		if ok {
			jsonReads = tmp
		}
	}
	return &GCSConfig{
		Bucket:         bucketName,
		Prefix:         prefix,
		Credentials:    credentials,
		EndPoint:       endPoint,
		ForcePathStyle: forcePathStyle,
		DisableSSL:     disableSSL,
		JSONReads:      jsonReads,
	}
}

type gcsListSession struct {
	*baseListSession
	manager *gcsManager

	Iterator *storage.ObjectIterator
}

func (l *gcsListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	maxItems := l.maxItems
	fileObjects = make([]*FileInfo, 0)

	// Create GCS storage client
	client, err := manager.getClient(l.ctx)
	if err != nil {
		return
	}

	// Create GCS Bucket handle
	if l.Iterator == nil {
		l.Iterator = client.Bucket(manager.config.Bucket).Objects(l.ctx, &storage.Query{
			Prefix:      l.prefix,
			Delimiter:   "",
			StartOffset: l.startAfter,
		})
	}
	var attrs *storage.ObjectAttrs
	for {
		if maxItems <= 0 {
			break
		}
		attrs, err = l.Iterator.Next()
		if err == iterator.Done || err != nil {
			if err == iterator.Done {
				err = nil
			}
			break
		}
		fileObjects = append(fileObjects, &FileInfo{attrs.Name, attrs.Updated})
		maxItems--
	}
	return
}
