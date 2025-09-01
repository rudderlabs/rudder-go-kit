//go:generate mockgen -destination=mock_filemanager/mock_filemanager.go -package mock_filemanager github.com/rudderlabs/rudder-go-kit/filemanager FileManager
package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const defaultTimeout = 120 * time.Second

var (
	ErrKeyNotFound            = errors.New("NoSuchKey")
	ErrInvalidServiceProvider = errors.New("service provider not supported")
	ErrPreConditionFailed     = errors.New("precondition failed")
)

// Factory is a function that returns a new file manager
type Factory func(settings *Settings) (FileManager, error)

// UploadedFile contains information about the uploaded file
type UploadedFile struct {
	Location   string
	ObjectName string
}

// FileInfo contains information about a file
type FileInfo struct {
	Key          string
	LastModified time.Time
}

type (
	DownloadOption  func(*downloadOptions)
	downloadOptions struct {
		offset         int64
		length         int64
		isRangeRequest bool
	}
)

// If offset is negative, it will be set to 0
func WithDownloadOffSet(offset int64) DownloadOption {
	return WithDownloadOffSetAndLength(offset, 0)
}

// If offset is negative, it will be set to 0
// If length is zero it will read the entire file from the offset
// If length is negative, it will read the entire file from the offset
func WithDownloadOffSetAndLength(offset, length int64) DownloadOption {
	return func(o *downloadOptions) {
		if offset < 0 {
			offset = 0
		}
		if length < 0 {
			length = 0
		}
		o.offset = offset
		o.length = length
		o.isRangeRequest = true
	}
}

func applyDownloadOptions(opts ...DownloadOption) downloadOptions {
	downloadOpts := downloadOptions{}
	for _, opt := range opts {
		opt(&downloadOpts)
	}
	return downloadOpts
}

// FileManager is able to manage files in a storage provider
type FileManager interface {
	// ListFilesWithPrefix starts a list session for files with given prefix
	ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession
	// Download retrieves an object with the given key and writes it to the provided writer.
	// You can Pass *os.File instead of io.WriterAt to write the downloaded data on disk.
	Download(context.Context, io.WriterAt, string, ...DownloadOption) error
	// Upload uploads the passed in file to the file manager
	Upload(context.Context, *os.File, ...string) (UploadedFile, error)
	// UploadReader uploads the passed io.Reader to the file manager
	UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error)
	// Delete deletes the file(s) with given key(s)
	Delete(ctx context.Context, keys []string) error

	// Prefix returns the prefix for the file manager
	Prefix() string
	// SetTimeout overrides the default timeout for the file manager
	SetTimeout(timeout time.Duration)

	// GetObjectNameFromLocation gets the object name/key name from the object location url
	GetObjectNameFromLocation(string) (string, error)
	// GetDownloadKeyFromFileLocation gets the download key from the object location url
	GetDownloadKeyFromFileLocation(string) string
}

// ListSession is a session for listing files
type ListSession interface {
	// Next returns the next batch of files, until there are no more files for this session
	Next() (fileObjects []*FileInfo, err error)
}

// Settings for file manager
type Settings struct {
	Provider string
	Config   map[string]interface{}
	Logger   logger.Logger
	Conf     *config.Config
}

// New returns file manager backed by configured provider
func New(settings *Settings) (FileManager, error) {
	log := settings.Logger
	if log == nil {
		log = logger.NewLogger().Child("filemanager")
	}
	conf := settings.Conf
	if conf == nil {
		conf = config.Default
	}

	switch settings.Provider {
	case "S3_DATALAKE", "S3":
		return NewS3Manager(conf, settings.Config, log, getDefaultTimeout(conf, settings.Provider))
	case "GCS":
		return NewGCSManager(settings.Config, log, getDefaultTimeout(conf, settings.Provider))
	case "AZURE_BLOB":
		return NewAzureBlobManager(settings.Config, log, getDefaultTimeout(conf, settings.Provider))
	case "MINIO":
		return NewMinioManager(settings.Config, log, getDefaultTimeout(conf, settings.Provider))
	case "DIGITAL_OCEAN_SPACES":
		return NewDigitalOceanManager(settings.Config, log, getDefaultTimeout(conf, settings.Provider))
	}
	return nil, fmt.Errorf("%w: %s", ErrInvalidServiceProvider, settings.Provider)
}

func getDefaultTimeout(config *config.Config, destType string) func() time.Duration {
	return func() time.Duration {
		key := "timeout"
		defaultValueInTimescaleUnits := int64(120)
		timeScale := time.Second
		if config.IsSet("FileManager." + destType + "." + key) {
			return config.GetDuration("FileManager."+destType+"."+key, defaultValueInTimescaleUnits, timeScale)
		}
		if config.IsSet("FileManager." + key) {
			return config.GetDuration("FileManager."+key, defaultValueInTimescaleUnits, timeScale)
		}
		return func() time.Duration { // legacy keys used in rudder-server
			destOverrideFound := config.IsSet("BatchRouter." + destType + "." + key)
			if destOverrideFound {
				return config.GetDuration("BatchRouter."+destType+"."+key, defaultValueInTimescaleUnits, timeScale)
			} else {
				return config.GetDuration("BatchRouter."+key, defaultValueInTimescaleUnits, timeScale)
			}
		}()
	}
}
