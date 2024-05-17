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
	Reset() error
}

// fileManagerImpl is a real implementation of FileManager
type fileManagerImpl struct {
	client Client
	config *SSHConfig
}

// Upload uploads a file to the remote server
func (fm *fileManagerImpl) Upload(localFilePath, remoteFilePath string) error {
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
	err := fm.client.Remove(remoteFilePath)
	if err != nil {
		return fmt.Errorf("cannot delete file: %w", err)
	}
	return nil
}

func (fm *fileManagerImpl) Reset() error {
	newFm, err := NewFileManager(fm.config)
	if err != nil {
		return err
	}
	fm.client = newFm.client
	return nil
}

func NewFileManager(config *SSHConfig) (*fileManagerImpl, error) {
	sshClient, err := newSSHClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating SSH client: %w", err)
	}
	sftpClient, err := newSFTPClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("cannot create SFTP client: %w", err)
	}
	return &fileManagerImpl{client: sftpClient, config: config}, nil
}

type retryableFileManagerImpl struct {
	fileManager FileManager
}

func (r *retryableFileManagerImpl) Upload(localFilePath, remoteFilePath string) error {
	fileOperation := func() error {
		return r.fileManager.Upload(localFilePath, remoteFilePath)
	}
	return r.retryOnConnectionLost(fileOperation)
}

func (r *retryableFileManagerImpl) Download(remoteFilePath, localDir string) error {
	fileOperation := func() error {
		return r.fileManager.Download(remoteFilePath, localDir)
	}
	return r.retryOnConnectionLost(fileOperation)
}

func (r *retryableFileManagerImpl) Delete(remoteFilePath string) error {
	fileOperation := func() error {
		return r.fileManager.Delete(remoteFilePath)
	}
	return r.retryOnConnectionLost(fileOperation)
}

func (r *retryableFileManagerImpl) Reset() error {
	return r.fileManager.Reset()
}

func (r *retryableFileManagerImpl) retryOnConnectionLost(fileOperation func() error) error {
	err := fileOperation()
	if err == nil || !isConnectionLostError(err) {
		return err // Operation successful or non-retryable error
	}

	if err := r.Reset(); err != nil {
		return err
	}

	// Retry the operation
	return fileOperation()
}

func NewRetryableFileManager(config *SSHConfig) (FileManager, error) {
	baseFileManager, err := NewFileManager(config)
	if err != nil {
		return nil, err
	}
	return &retryableFileManagerImpl{fileManager: baseFileManager}, nil
}

func isConnectionLostError(err error) bool {
	return errors.Is(err, sftp.ErrSshFxConnectionLost)
}
