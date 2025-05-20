package filemanager

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
)

type switchingDigitalOceanManager struct {
	isV2ManagerEnabled    config.ValueLoader[bool]
	digitalOceanManagerV2 *DigitalOceanManagerV2
	digitalOceanManager   *DigitalOceanManager
}

// ListFilesWithPrefix starts a list session for files with given prefix
func (s *switchingDigitalOceanManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return s.getManager().ListFilesWithPrefix(ctx, startAfter, prefix, maxItems)
}

// Download retrieves an object with the given key and writes it to the provided writer.
// You can Pass *os.File instead of io.WriterAt to write the downloaded data on disk.
func (s *switchingDigitalOceanManager) Download(ctx context.Context, writer io.WriterAt, key string) error {
	return s.getManager().Download(ctx, writer, key)
}

// Upload uploads the passed in file to the file manager
func (s *switchingDigitalOceanManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	return s.getManager().Upload(ctx, file, prefixes...)
}

// UploadReader uploads the passed io.Reader to the file manager
func (s *switchingDigitalOceanManager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	return s.getManager().UploadReader(ctx, objName, rdr)
}

// Delete deletes the file(s) with given key(s)
func (s *switchingDigitalOceanManager) Delete(ctx context.Context, keys []string) error {
	return s.getManager().Delete(ctx, keys)
}

// Prefix returns the prefix for the file manager
func (s *switchingDigitalOceanManager) Prefix() string {
	return s.getManager().Prefix()
}

// SetTimeout overrides the default timeout for the file manager
func (s *switchingDigitalOceanManager) SetTimeout(timeout time.Duration) {
	s.getManager().SetTimeout(timeout)
}

// GetObjectNameFromLocation gets the object name/key name from the object location url
func (s *switchingDigitalOceanManager) GetObjectNameFromLocation(location string) (string, error) {
	return s.getManager().GetObjectNameFromLocation(location)
}

// GetDownloadKeyFromFileLocation gets the download key from the object location url
func (s *switchingDigitalOceanManager) GetDownloadKeyFromFileLocation(location string) string {
	return s.getManager().GetDownloadKeyFromFileLocation(location)
}

func (s *switchingDigitalOceanManager) getManager() FileManager {
	if s.isV2ManagerEnabled.Load() {
		return s.digitalOceanManagerV2
	}
	return s.digitalOceanManager
}
