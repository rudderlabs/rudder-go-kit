package sftp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	// PasswordAuth indicates password-based authentication
	PasswordAuth = "passwordAuth"
	// KeyAuth indicates key-based authentication
	KeyAuth = "keyAuth"
)

// SSHConfig represents the configuration for SSH connection
type SSHConfig struct {
	Hostname   string
	Port       int
	Username   string
	AuthMethod string
	PrivateKey string
	Password   string // Password for password-based authentication
}

// SFTPClient represents an SFTP client
type SFTPClient struct {
	client *sftp.Client
}

// NewSSHClient establishes an SSH connection and returns an SSH client
func NewSSHClient(config *SSHConfig) (*ssh.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config should not be nil")
	}

	var authMethods []ssh.AuthMethod

	switch config.AuthMethod {
	case PasswordAuth:
		authMethods = []ssh.AuthMethod{ssh.Password(config.Password)}
	case KeyAuth:

		privateKey, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, err
		}
		authMethods = []ssh.AuthMethod{ssh.PublicKeys(privateKey)}
	default:
		return nil, fmt.Errorf("unsupported authentication method")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use it only for testing purposes. Not recommended for production.
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Hostname, config.Port), sshConfig)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

// NewSFTPClient creates a new SFTP client from an existing SSH client
func NewSFTPClient(sshClient *ssh.Client) (*SFTPClient, error) {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}
	return &SFTPClient{client: sftpClient}, nil
}

// UploadFile uploads a file to the remote server
func (s *SFTPClient) UploadFile(localFilePath, remoteDir string) error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFileName := filepath.Join(remoteDir, filepath.Base(localFilePath))
	remoteFile, err := s.client.Create(remoteFileName)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return err
	}

	fmt.Printf("Uploaded %s to %s\n", localFilePath, remoteDir)
	return nil
}

// DownloadFile downloads a file from the remote server
func (s *SFTPClient) DownloadFile(remoteFilePath, localDir string) error {
	remoteFile, err := s.client.Open(remoteFilePath)
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
func (s *SFTPClient) DeleteFile(remoteFilePath string) error {
	err := s.client.Remove(remoteFilePath)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", remoteFilePath)
	return nil
}
