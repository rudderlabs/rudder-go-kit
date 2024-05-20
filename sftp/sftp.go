package sftp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

const (
	// PasswordAuth indicates password-based authentication
	PasswordAuth = "passwordAuth"
	// KeyAuth indicates key-based authentication
	KeyAuth = "keyAuth"
)

// FileManager is an interface for managing files on a remote server
type FileManager interface {
	Upload(localFilePath, remoteDir string) error
	Download(remoteFilePath, localDir string) error
	Delete(remoteFilePath string) error
}

type Option func(impl *fileManagerImpl)

// WithRetryOnIdleConnection enables retrying the operation once in case of a "connection lost" error due to an idle connection.
func WithRetryOnIdleConnection() Option {
	return func(impl *fileManagerImpl) {
		impl.retryOnIdleConnection = true
	}
}

// fileManagerImpl is a real implementation of FileManager
type fileManagerImpl struct {
	client                Client
	retryOnIdleConnection bool
}

// Upload uploads a file to the remote server
func (fm *fileManagerImpl) Upload(localFilePath, remoteFilePath string) error {
	if fm.retryOnIdleConnection {
		return fm.retryOnConnectionLost(func() error {
			return fm.upload(localFilePath, remoteFilePath)
		})
	}

	return fm.upload(localFilePath, remoteFilePath)
}

func (fm *fileManagerImpl) upload(localFilePath, remoteFilePath string) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("cannot open local file: %w", err)
	}
	defer func() {
		_ = localFile.Close()
	}()

	// Create the directory if it does not exist
	remoteDir := filepath.Dir(remoteFilePath)
	if err := fm.client.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("cannot create remote directory: %w", err)
	}

	remoteFile, err := fm.client.OpenFile(remoteFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("cannot create remote file: %w", err)
	}
	defer func() {
		_ = remoteFile.Close()
	}()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

// Download downloads a file from the remote server
func (fm *fileManagerImpl) Download(remoteFilePath, localDir string) error {
	if fm.retryOnIdleConnection {
		return fm.retryOnConnectionLost(func() error {
			return fm.download(remoteFilePath, localDir)
		})
	}

	return fm.download(remoteFilePath, localDir)
}

func (fm *fileManagerImpl) download(remoteFilePath, localDir string) error {
	remoteFile, err := fm.client.OpenFile(remoteFilePath, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("cannot open remote file: %w", err)
	}
	defer func() {
		_ = remoteFile.Close()
	}()

	localFileName := filepath.Join(localDir, filepath.Base(remoteFilePath))
	localFile, err := os.Create(localFileName)
	if err != nil {
		return fmt.Errorf("cannot create local file: %w", err)
	}
	defer func() {
		_ = localFile.Close()
	}()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("cannot copy remote file content to local file: %w", err)
	}

	return nil
}

// Delete deletes a file on the remote server
func (fm *fileManagerImpl) Delete(remoteFilePath string) error {
	if fm.retryOnIdleConnection {
		return fm.retryOnConnectionLost(func() error {
			return fm.delete(remoteFilePath)
		})
	}

	return fm.delete(remoteFilePath)
}

func (fm *fileManagerImpl) delete(remoteFilePath string) error {
	err := fm.client.Remove(remoteFilePath)
	if err != nil {
		return fmt.Errorf("cannot delete file: %w", err)
	}
	return nil
}

func (fm *fileManagerImpl) reset() error {
	return fm.client.Reset()
}

// NewFileManager is not concurrent safe. It should not be used from multiple goroutines concurrently without additional synchronization.
func NewFileManager(config *SSHConfig, opts ...Option) (FileManager, error) {
	sftpClient, err := newClient(config)
	if err != nil {
		return nil, err
	}
	fm := &fileManagerImpl{client: sftpClient}
	for _, opt := range opts {
		opt(fm)
	}
	return fm, nil
}

func (fm *fileManagerImpl) retryOnConnectionLost(fileOperation func() error) error {
	err := fileOperation()
	if err == nil || !isConnectionLostError(err) {
		return err // Operation successful or non-retryable error
	}

	if err := fm.reset(); err != nil {
		return err
	}

	// Retry the operation
	return fileOperation()
}

func isConnectionLostError(err error) bool {
	return errors.Is(err, sftp.ErrSshFxConnectionLost)
}
