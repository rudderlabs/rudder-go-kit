package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/rudderlabs/rudder-go-kit/googleutil"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type GCSConfig struct {
	Bucket           string
	Prefix           string
	Credentials      string
	EndPoint         *string
	ForcePathStyle   *bool
	DisableSSL       *bool
	JSONReads        bool
	UploadIfNotExist bool
}

// NewGCSManager creates a new file manager for Google Cloud Storage
func NewGCSManager(
	config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration,
) (*GcsManager, error) {
	conf := gcsConfig(config)
	return &GcsManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config: conf,
	}, nil
}

func (m *GcsManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &gcsListSession{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager: m,
	}
}

// Download retrieves an object with the given key and writes it to the provided writer.
// Pass *os.File as output to write the downloaded file on disk.
func (m *GcsManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	client, err := m.getClient(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	obj := client.Bucket(m.config.Bucket).Object(key)

	var rc *storage.Reader
	if downloadOpts.isRangeRequest {
		length := downloadOpts.length
		if length <= 0 {
			length = -1
		}
		rc, err = obj.NewRangeReader(ctx, downloadOpts.offset, length)
	} else {
		rc, err = obj.NewReader(ctx)
	}
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	writer := &writerAtAdapter{w: output}
	_, err = io.Copy(writer, rc)
	return err
}

func (m *GcsManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *GcsManager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	client, err := m.getClient(ctx)
	if err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	object := client.Bucket(m.config.Bucket).Object(objName)
	if m.config.UploadIfNotExist {
		object = object.If(storage.Conditions{DoesNotExist: true})
	}

	w := object.NewWriter(ctx)
	if _, err := io.Copy(w, rdr); err != nil {
		return UploadedFile{}, fmt.Errorf("copying file to writer: %w", err)
	}

	if err := w.Close(); err != nil {
		var ge *googleapi.Error
		if errors.As(err, &ge) && ge.Code == http.StatusPreconditionFailed {
			return UploadedFile{}, ErrPreConditionFailed
		}
		return UploadedFile{}, fmt.Errorf("closing writer: %w", err)
	}

	return UploadedFile{Location: m.objectURL(w.Attrs()), ObjectName: objName}, err
}

func (m *GcsManager) Delete(ctx context.Context, keys []string) (err error) {
	client, err := m.getClient(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	for _, key := range keys {
		if err := client.Bucket(m.config.Bucket).Object(key).Delete(ctx); err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
			return err
		}
	}
	return
}

func (m *GcsManager) Prefix() string {
	return m.config.Prefix
}

func (m *GcsManager) GetObjectNameFromLocation(location string) (string, error) {
	splitStr := strings.Split(location, m.config.Bucket)
	object := strings.TrimLeft(splitStr[len(splitStr)-1], "/")
	return object, nil
}

func (m *GcsManager) GetDownloadKeyFromFileLocation(location string) string {
	splitStr := strings.Split(location, m.config.Bucket)
	key := strings.TrimLeft(splitStr[len(splitStr)-1], "/")
	return key
}

func (m *GcsManager) objectURL(objAttrs *storage.ObjectAttrs) string {
	if m.config.EndPoint != nil && *m.config.EndPoint != "" {
		endpoint := strings.TrimSuffix(*m.config.EndPoint, "/")
		return fmt.Sprintf("%s/%s/%s", endpoint, objAttrs.Bucket, objAttrs.Name)
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", objAttrs.Bucket, objAttrs.Name)
}

func (m *GcsManager) getClient(ctx context.Context) (*storage.Client, error) {
	m.clientMu.Lock()
	defer m.clientMu.Unlock()

	if m.client != nil {
		return m.client, nil
	}

	options := []option.ClientOption{option.WithTelemetryDisabled()}
	if m.config.EndPoint != nil && *m.config.EndPoint != "" {
		options = append(options, option.WithEndpoint(*m.config.EndPoint))
	}
	if !googleutil.ShouldSkipCredentialsInit(m.config.Credentials) {
		if err := googleutil.CompatibleGoogleCredentialsJSON([]byte(m.config.Credentials)); err != nil {
			return m.client, err
		}
		options = append(options, option.WithCredentialsJSON([]byte(m.config.Credentials)))
	}
	if m.config.JSONReads {
		options = append(options, storage.WithJSONReads())
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	var err error
	m.client, err = storage.NewClient(ctx, options...)
	return m.client, err
}

type GcsManager struct {
	*baseManager
	config *GCSConfig

	client   *storage.Client
	clientMu sync.Mutex
}

func gcsConfig(config map[string]interface{}) *GCSConfig {
	var bucketName, prefix, credentials string
	var endPoint *string
	var forcePathStyle, disableSSL *bool
	var jsonReads, uploadIfNotExist bool

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
	if config["uploadIfNotExist"] != nil {
		tmp, ok := config["uploadIfNotExist"].(bool)
		if ok {
			uploadIfNotExist = tmp
		}
	}
	return &GCSConfig{
		Bucket:           bucketName,
		Prefix:           prefix,
		Credentials:      credentials,
		EndPoint:         endPoint,
		ForcePathStyle:   forcePathStyle,
		DisableSSL:       disableSSL,
		JSONReads:        jsonReads,
		UploadIfNotExist: uploadIfNotExist,
	}
}

type gcsListSession struct {
	*baseListSession
	manager *GcsManager

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
	for maxItems > 0 {
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
