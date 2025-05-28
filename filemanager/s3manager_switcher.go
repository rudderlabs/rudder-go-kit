package filemanager

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type S3Manager interface {
	FileManager
	Bucket() string
	SelectObjects(ctx context.Context, sqlExpession, key string) (<-chan []byte, error)
}

func NewS3Manager(conf *kitconfig.Config, config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (S3Manager, error) {
	v2Enabled := conf.GetReloadableBoolVar(false, "FileManager.useAwsSdkV2")
	s3Manager, err := newS3ManagerV1(conf, config, log, defaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 V1 manager: %w", err)
	}
	s3ManagerV2, err := newS3ManagerV2(conf, config, log, defaultTimeout)
	if err != nil {
		if v2Enabled.Load() { // if v2 is enabled, return error
			return nil, fmt.Errorf("failed to create S3 V2 manager: %w", err)
		} else {
			// if v2 is not enabled, log the error
			log.Errorn("Failed to create S3 V2 manager, falling back to V1", logger.NewErrorField(err))
		}
	}
	return &switchingS3Manager{
		isV2ManagerEnabled: v2Enabled,
		s3Manager:          s3Manager,
		s3ManagerV2:        s3ManagerV2,
	}, nil
}

type switchingS3Manager struct {
	isV2ManagerEnabled kitconfig.ValueLoader[bool]
	s3ManagerV2        *s3ManagerV2
	s3Manager          *s3ManagerV1
}

// ListFilesWithPrefix starts a list session for files with given prefix
func (s *switchingS3Manager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return s.getManager().ListFilesWithPrefix(ctx, startAfter, prefix, maxItems)
}

// Download retrieves an object with the given key and writes it to the provided writer.
// You can Pass *os.File instead of io.WriterAt to write the downloaded data on disk.
func (s *switchingS3Manager) Download(ctx context.Context, writer io.WriterAt, key string) error {
	return s.getManager().Download(ctx, writer, key)
}

// Upload uploads the passed in file to the file manager
func (s *switchingS3Manager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	return s.getManager().Upload(ctx, file, prefixes...)
}

// UploadReader uploads the passed io.Reader to the file manager
func (s *switchingS3Manager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	return s.getManager().UploadReader(ctx, objName, rdr)
}

// Delete deletes the file(s) with given key(s)
func (s *switchingS3Manager) Delete(ctx context.Context, keys []string) error {
	return s.getManager().Delete(ctx, keys)
}

// Prefix returns the prefix for the file manager
func (s *switchingS3Manager) Prefix() string {
	return s.getManager().Prefix()
}

// SetTimeout overrides the default timeout for the file manager
func (s *switchingS3Manager) SetTimeout(timeout time.Duration) {
	s.getManager().SetTimeout(timeout)
}

// GetObjectNameFromLocation gets the object name/key name from the object location url
func (s *switchingS3Manager) GetObjectNameFromLocation(location string) (string, error) {
	return s.getManager().GetObjectNameFromLocation(location)
}

// GetDownloadKeyFromFileLocation gets the download key from the object location url
func (s *switchingS3Manager) GetDownloadKeyFromFileLocation(location string) string {
	return s.getManager().GetDownloadKeyFromFileLocation(location)
}

func (s *switchingS3Manager) Bucket() string {
	return s.getManager().Bucket()
}

func (s *switchingS3Manager) SelectObjects(ctx context.Context, sqlExpession, key string) (<-chan []byte, error) {
	return s.getManager().SelectObjects(ctx, sqlExpession, key)
}

func (s *switchingS3Manager) getManager() S3Manager {
	if s.isV2ManagerEnabled.Load() && s.s3ManagerV2 != nil {
		return s.s3ManagerV2
	}
	return s.s3Manager
}
