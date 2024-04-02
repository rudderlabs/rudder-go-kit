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

// SFTPManager interface abstracts the SFTP client
type SFTPManager interface {
	UploadFile(localFilePath, remoteDir string) error
	DownloadFile(remoteFilePath, localDir string) error
	DeleteFile(remoteFilePath string) error
}

// SFTPManagerImpl is a real implementation of SFTPManager
type SFTPManagerImpl struct {
	client SFTPClient
}

func NewSFTPManager(sshClient *ssh.Client) (SFTPManager, error) {
	sftpClient, err := NewSFTPClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("cannot create SFTP client: %w", err)
	}
	return &SFTPManagerImpl{client: sftpClient}, nil
}

// UploadFile uploads a file to the remote server
func (r *SFTPManagerImpl) UploadFile(localFilePath, remoteDir string) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFileName := filepath.Join(remoteDir, filepath.Base(localFilePath))
	remoteFile, err := r.client.Create(remoteFileName)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		fmt.Printf("Error copying file: %v\n", err)
		return err
	}

	fmt.Printf("Uploaded %s to %s\n", localFilePath, remoteDir)
	return nil
}

// DownloadFile downloads a file from the remote server
func (r *SFTPManagerImpl) DownloadFile(remoteFilePath, localDir string) error {
	remoteFile, err := r.client.Open(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFileName := filepath.Join(localDir, filepath.Base(remoteFilePath))
	localFile, err := os.Create(localFileName)
	if err != nil {
		return err
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded %s to %s\n", remoteFilePath, localDir)
	return nil
}

// DeleteFile deletes a file on the remote server
func (r *SFTPManagerImpl) DeleteFile(remoteFilePath string) error {
	err := r.client.Remove(remoteFilePath)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", remoteFilePath)
	return nil
}
