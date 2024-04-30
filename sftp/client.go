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

// NewSSHClient establishes an SSH connection and returns an SSH client
func NewSSHClient(config *SSHConfig) (*ssh.Client, error) {
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
	client *sftp.Client
}

type Client interface {
	Open(path string, flag int) (io.ReadWriteCloser, error)
	Remove(path string) error
	MkdirAll(path string) error
}

// newSFTPClient creates an SFTP client with existing SSH client
func newSFTPClient(client *ssh.Client) (Client, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("cannot create SFTP client: %w", err)
	}
	return &clientImpl{
		client: sftpClient,
	}, nil
}

func (c *clientImpl) Open(path string, flag int) (io.ReadWriteCloser, error) {
	return c.client.OpenFile(path, flag)
}

func (c *clientImpl) Remove(path string) error {
	return c.client.Remove(path)
}

func (c *clientImpl) MkdirAll(path string) error {
	return c.client.MkdirAll(path)
}
