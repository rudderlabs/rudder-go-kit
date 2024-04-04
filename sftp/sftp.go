package sftp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
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

// fileManagerImpl is a real implementation of FileManager
type fileManagerImpl struct {
	client Client
}

func NewFileManager(sshClient *ssh.Client) (FileManager, error) {
	sftpClient, err := NewSFTPClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("cannot create SFTP client: %w", err)
	}
	return &fileManagerImpl{client: sftpClient}, nil
}

// Upload uploads a file to the remote server
func (fm *fileManagerImpl) Upload(localFilePath, remoteDir string) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("cannot open local file: %w", err)
	}
	defer func() {
		_ = localFile.Close()
	}()

	remoteFileName := filepath.Join(remoteDir, filepath.Base(localFilePath))
	remoteFile, err := fm.client.Create(remoteFileName)
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
	remoteFile, err := fm.client.Open(remoteFilePath)
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
