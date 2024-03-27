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
	Host       string
	Port       int
	User       string
	AuthMethod string
	PrivateKey string
	Password   string // Password for password-based authentication
}

// SSHClient interface abstracts the SSH client
type SSHClient interface {
	Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}

// SFTPClient interface abstracts the SFTP client
type SFTPClient interface {
	CreateNew(client *ssh.Client) (*sftp.Client, error)
	UploadFile(localFilePath, remoteDir string) error
	DownloadFile(remoteFilePath, localDir string) error
	DeleteFile(remoteFilePath string) error
}

// SSHClientImpl is a real implementation of SSHClient
type SSHClientImpl struct{}

// Dial establishes an SSH connection
func (r *SSHClientImpl) Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial(network, addr, config)
}

// SFTPClientImpl is a real implementation of SFTPClient
type SFTPClientImpl struct {
	client *sftp.Client
}

// Create creates a file on the remote server
func (r *SFTPClientImpl) Create(remoteFilePath string) (io.WriteCloser, error) {
	return r.client.Create(remoteFilePath)
}

// Create creates an SFTP client from an existing SSH client
func (r *SFTPClientImpl) CreateNew(client *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(client)
}

// NewSSHClient establishes an SSH connection and returns an SSH client
func NewSSHClient(config *SSHConfig, client SSHClient) (*ssh.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config should not be nil")
	}

	if config.Host == "" {
		return nil, fmt.Errorf("host should not be empty")
	}

	if config.User == "" {
		return nil, fmt.Errorf("user should not be empty")
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
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use it only for testing purposes. Not recommended for production.
	}

	if client == nil {
		// Use real implementation if client is not provided
		client = &SSHClientImpl{}
	}
	sshClient, err := client.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot dial SSH host %q: %w", config.Host, err)
	}
	return sshClient, nil
}

func NewSFTPClient(sshClient *ssh.Client, client SFTPClient) (*SFTPClientImpl, error) {
	if client == nil {
		// Use real implementation if client is not provided
		realClient := &SFTPClientImpl{}
		sftpClient, err := realClient.CreateNew(sshClient)
		if err != nil {
			return nil, err // Return the error
		}
		return &SFTPClientImpl{client: sftpClient}, nil
	}
	sftpClient, err := client.CreateNew(sshClient)
	if err != nil {
		return nil, err // Return the error
	}
	return &SFTPClientImpl{client: sftpClient}, nil
}

// UploadFile uploads a file to the remote server
func (r *SFTPClientImpl) UploadFile(localFilePath, remoteDir string) error {
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
		return err
	}

	fmt.Printf("Uploaded %s to %s\n", localFilePath, remoteDir)
	return nil
}

// DownloadFile downloads a file from the remote server
func (r *SFTPClientImpl) DownloadFile(remoteFilePath, localDir string) error {
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
func (r *SFTPClientImpl) DeleteFile(remoteFilePath string) error {
	err := r.client.Remove(remoteFilePath)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", remoteFilePath)
	return nil
}
