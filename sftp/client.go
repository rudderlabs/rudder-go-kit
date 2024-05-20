//go:generate mockgen -destination=mock_sftp/mock_sftp_client.go -package mock_sftp github.com/rudderlabs/rudder-go-kit/sftp Client
package sftp

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHConfig represents the configuration for SSH connection
type SSHConfig struct {
	HostName    string
	Port        int
	User        string
	AuthMethod  string
	PrivateKey  string
	Password    string // Password for password-based authentication
	DialTimeout time.Duration
}

// sshClientConfig constructs an SSH client configuration based on the provided SSHConfig.
func sshClientConfig(config *SSHConfig) (*ssh.ClientConfig, error) {
	if config == nil {
		return nil, errors.New("config should not be nil")
	}

	if config.HostName == "" {
		return nil, errors.New("hostname should not be empty")
	}

	if config.Port == 0 {
		return nil, errors.New("port should not be empty")
	}

	if config.User == "" {
		return nil, errors.New("user should not be empty")
	}

	var authMethods ssh.AuthMethod

	switch config.AuthMethod {
	case PasswordAuth:
		authMethods = ssh.Password(config.Password)
	case KeyAuth:
		privateKey, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("cannot parse private key: %w", err)
		}
		authMethods = ssh.PublicKeys(privateKey)
	default:
		return nil, errors.New("unsupported authentication method")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            []ssh.AuthMethod{authMethods},
		Timeout:         config.DialTimeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return sshConfig, nil
}

// newSSHClient establishes an SSH connection and returns an SSH client
func newSSHClient(config *SSHConfig) (*ssh.Client, error) {
	sshConfig, err := sshClientConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot configure SSH client: %w", err)
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.HostName, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot dial SSH host %q:%d: %w", config.HostName, config.Port, err)
	}
	return sshClient, nil
}

type clientImpl struct {
	sftpClient *sftp.Client
	config     *SSHConfig
}

type Client interface {
	OpenFile(path string, f int) (io.ReadWriteCloser, error)
	Remove(path string) error
	MkdirAll(path string) error
	Reset() error
}

// newSFTPClient creates an SFTP client with existing SSH client
func newSFTPClient(client *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(client)
}

func newSFTPClientFromConfig(config *SSHConfig) (*sftp.Client, error) {
	sshClient, err := newSSHClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating SSH client: %w", err)
	}
	return newSFTPClient(sshClient)
}

func newClient(config *SSHConfig) (Client, error) {
	sftpClient, err := newSFTPClientFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating SFTP client: %w", err)
	}
	return &clientImpl{
		sftpClient: sftpClient,
		config:     config,
	}, nil
}

func (c *clientImpl) OpenFile(path string, f int) (io.ReadWriteCloser, error) {
	return c.sftpClient.OpenFile(path, f)
}

func (c *clientImpl) Remove(path string) error {
	return c.sftpClient.Remove(path)
}

func (c *clientImpl) MkdirAll(path string) error {
	return c.sftpClient.MkdirAll(path)
}

func (c *clientImpl) Reset() error {
	newSFTPClient, err := newSFTPClientFromConfig(c.config)
	if err != nil {
		return err
	}
	c.sftpClient = newSFTPClient
	return nil
}
