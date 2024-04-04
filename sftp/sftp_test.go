package sftp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/sftp/mock_sftp"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/sshserver"
)

func TestSSHClientConfig(t *testing.T) {
	// Read private key
	privateKey, err := os.ReadFile("testdata/ssh/test_key")
	require.NoError(t, err)

	type testCase struct {
		description   string
		config        *SSHConfig
		expectedError error
	}

	testCases := []testCase{
		{
			description:   "WithNilConfig",
			config:        nil,
			expectedError: fmt.Errorf("config should not be nil"),
		},
		{
			description: "WithEmptyHostName",
			config: &SSHConfig{
				HostName:   "",
				Port:       22,
				User:       "someUser",
				AuthMethod: "passwordAuth",
				Password:   "somePassword",
			},
			expectedError: fmt.Errorf("hostname should not be empty"),
		},
		{
			description: "WithEmptyPort",
			config: &SSHConfig{
				HostName:   "someHostName",
				User:       "someUser",
				AuthMethod: "passwordAuth",
				Password:   "somePassword",
			},
			expectedError: fmt.Errorf("port should not be empty"),
		},
		{
			description: "WithPassword",
			config: &SSHConfig{
				HostName:   "someHostName",
				Port:       22,
				User:       "someUser",
				AuthMethod: "passwordAuth",
				Password:   "somePassword",
			},
			expectedError: nil,
		},
		{
			description: "WithPrivateKey",
			config: &SSHConfig{
				HostName:   "someHostName",
				Port:       22,
				User:       "someUser",
				AuthMethod: "keyAuth",
				PrivateKey: string(privateKey),
			},
			expectedError: nil,
		},
		{
			description: "WithUnsupportedAuthMethod",
			config: &SSHConfig{
				HostName:   "HostName",
				Port:       22,
				User:       "someUser",
				AuthMethod: "invalidAuth",
				PrivateKey: "somePrivateKey",
			},
			expectedError: fmt.Errorf("unsupported authentication method"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			sshConfig, err := sshClientConfig(tc.config)
			if tc.expectedError != nil {
				assert.EqualError(t, tc.expectedError, err.Error())
				assert.Nil(t, sshConfig)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, sshConfig)
			}
		})
	}
}

func TestUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create local and remote directories within the temporary directory
	baseDir := t.TempDir()
	localDir := filepath.Join(baseDir, "local")
	remoteDir := filepath.Join(baseDir, "remote")
	err := os.MkdirAll(localDir, 0o755)
	require.NoError(t, err)
	err = os.MkdirAll(remoteDir, 0o755)
	require.NoError(t, err)

	// Set up local and remote file paths within their respective directories
	localFilePath := filepath.Join(localDir, "test_file.json")
	remoteFilePath := filepath.Join(remoteDir, "test_file.json")

	// Create local file and write data to it
	localFile, err := os.Create(localFilePath)
	require.NoError(t, err)
	defer func() { _ = localFile.Close() }()
	data := []byte(`{"foo": "bar"}`)
	err = os.WriteFile(localFilePath, data, 0o644)
	require.NoError(t, err)

	// Create remote file
	remoteFile, err := os.Create(remoteFilePath)
	require.NoError(t, err)

	mockSFTPClient := mock_sftp.NewMockClient(ctrl)
	mockSFTPClient.EXPECT().Create(remoteFilePath).Return(remoteFile, nil)

	fileManager := &fileManagerImpl{client: mockSFTPClient}

	err = fileManager.Upload(localFilePath, remoteDir)
	require.NoError(t, err)
	remoteFileContents, err := os.ReadFile(remoteFilePath)
	require.NoError(t, err)
	localFileContents, err := os.ReadFile(localFilePath)
	require.NoError(t, err)
	assert.Equal(t, localFileContents, remoteFileContents)
}

func TestDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create local and remote directories within the temporary directory
	baseDir := t.TempDir()
	localDir := filepath.Join(baseDir, "local")
	remoteDir := filepath.Join(baseDir, "remote")
	err := os.MkdirAll(localDir, 0o755)
	require.NoError(t, err)
	err = os.MkdirAll(remoteDir, 0o755)
	require.NoError(t, err)

	// Set up local and remote file paths within their respective directories
	localFilePath := filepath.Join(localDir, "test_file.json")
	remoteFilePath := filepath.Join(remoteDir, "test_file.json")

	// Create remote file and write data to it
	remoteFile, err := os.Create(remoteFilePath)
	require.NoError(t, err)
	defer func() { _ = remoteFile.Close() }()
	data := []byte(`{"foo": "bar"}`)
	err = os.WriteFile(remoteFilePath, data, 0o644)
	require.NoError(t, err)

	mockSFTPClient := mock_sftp.NewMockClient(ctrl)
	mockSFTPClient.EXPECT().Open(remoteFilePath).Return(remoteFile, nil)

	fileManager := &fileManagerImpl{client: mockSFTPClient}

	err = fileManager.Download(remoteFilePath, localDir)
	require.NoError(t, err)
	remoteFileContents, err := os.ReadFile(remoteFilePath)
	require.NoError(t, err)
	localFileContents, err := os.ReadFile(localFilePath)
	require.NoError(t, err)
	assert.Equal(t, localFileContents, remoteFileContents)
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	remoteFilePath := "someRemoteFilePath"
	mockSFTPClient := mock_sftp.NewMockClient(ctrl)
	mockSFTPClient.EXPECT().Remove(remoteFilePath).Return(nil)

	fileManager := &fileManagerImpl{client: mockSFTPClient}

	err := fileManager.Delete(remoteFilePath)
	require.NoError(t, err)
}

func TestSFTP(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	// Let's setup the SSH server
	publicKeyPath, err := filepath.Abs("testdata/ssh/test_key.pub")
	require.NoError(t, err)
	sshServer, err := sshserver.Setup(pool, t,
		sshserver.WithPublicKeyPath(publicKeyPath),
		sshserver.WithCredentials("linuxserver.io", ""),
	)
	require.NoError(t, err)
	sshServerHost := fmt.Sprintf("localhost:%d", sshServer.Port)
	t.Logf("SSH server is listening on %s", sshServerHost)

	// Read private key
	privateKey, err := os.ReadFile("testdata/ssh/test_key")
	require.NoError(t, err)

	// Setup ssh client
	hostname, portStr, err := net.SplitHostPort(sshServerHost)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	sshClient, err := NewSSHClient(&SSHConfig{
		User:        "linuxserver.io",
		HostName:    hostname,
		Port:        port,
		AuthMethod:  "keyAuth",
		PrivateKey:  string(privateKey),
		DialTimeout: 10 * time.Second,
	})
	require.NoError(t, err)

	// Create session
	session, err := sshClient.NewSession()
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	remoteDir := filepath.Join("/tmp", "remote")
	err = session.Run(fmt.Sprintf("mkdir -p %s", remoteDir))
	require.NoError(t, err)

	sftpManger, err := NewFileManager(sshClient)
	require.NoError(t, err)

	// Create local and remote directories within the temporary directory
	baseDir := t.TempDir()
	localDir := filepath.Join(baseDir, "local")

	err = os.MkdirAll(localDir, 0o755)
	require.NoError(t, err)

	// Set up local and remote file paths within their respective directories
	localFilePath := filepath.Join(localDir, "test_file.json")
	remoteFilePath := filepath.Join(remoteDir, "test_file.json")

	// Create local file and write data to it
	localFile, err := os.Create(localFilePath)
	require.NoError(t, err)
	defer func() { _ = localFile.Close() }()
	data := []byte(`{"foo": "bar"}`)
	err = os.WriteFile(localFilePath, data, 0o644)
	require.NoError(t, err)

	err = sftpManger.Upload(localFilePath, remoteDir)
	require.NoError(t, err)

	err = sftpManger.Download(remoteFilePath, baseDir)
	require.NoError(t, err)

	localFileContents, err := os.ReadFile(localFilePath)
	require.NoError(t, err)
	downloadedFileContents, err := os.ReadFile(filepath.Join(baseDir, "test_file.json"))
	require.NoError(t, err)
	// Compare the contents of the local file and the downloaded file from the remote server
	assert.Equal(t, localFileContents, downloadedFileContents)

	err = sftpManger.Delete(remoteFilePath)
	require.NoError(t, err)
}
