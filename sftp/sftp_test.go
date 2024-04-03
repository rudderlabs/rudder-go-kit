package sftp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/sftp/mock_sftp"
)

func TestConfigureSSHClient(t *testing.T) {
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
			description: "WithEmptyHost",
			config: &SSHConfig{
				Host:       "",
				Port:       22,
				User:       "someUser",
				AuthMethod: "passwordAuth",
				Password:   "somePassword",
			},
			expectedError: fmt.Errorf("host should not be empty"),
		},
		{
			description: "WithPassword",
			config: &SSHConfig{
				Host:       "someHost",
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
				Host:       "someHost",
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
				Host:       "someHost",
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
			sshClient, err := ConfigureSSHClient(tc.config)
			if tc.expectedError != nil {
				assert.EqualError(t, tc.expectedError, err.Error())
				assert.Nil(t, sshClient)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, sshClient)
			}
		})
	}
}

func TestUploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create local and remote directories within the temporary directory
	localDir := filepath.Join(t.TempDir(), "local")
	remoteDir := filepath.Join(t.TempDir(), "remote")
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
	defer localFile.Close()
	data := []byte(`{"foo": "bar"}`)
	err = os.WriteFile(localFilePath, data, 0o644)
	require.NoError(t, err)

	// Create remote file
	remoteFile, err := os.Create(remoteFilePath)
	require.NoError(t, err)

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().Create(remoteFilePath).Return(remoteFile, nil)

	sftpManager := &SFTPManagerImpl{client: mockSFTPClient}

	err = sftpManager.UploadFile(localFilePath, remoteDir)
	require.NoError(t, err)
	remoteFileContents, err := os.ReadFile(remoteFilePath)
	require.NoError(t, err)
	localFileContents, err := os.ReadFile(localFilePath)
	require.NoError(t, err)
	assert.Equal(t, localFileContents, remoteFileContents)
}

func TestDownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create local and remote directories within the temporary directory
	localDir := filepath.Join(t.TempDir(), "local")
	remoteDir := filepath.Join(t.TempDir(), "remote")
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
	defer remoteFile.Close()
	data := []byte(`{"foo": "bar"}`)
	err = os.WriteFile(remoteFilePath, data, 0o644)
	require.NoError(t, err)

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().Open(remoteFilePath).Return(remoteFile, nil)

	sftpManager := &SFTPManagerImpl{client: mockSFTPClient}

	err = sftpManager.DownloadFile(remoteFilePath, localDir)
	require.NoError(t, err)
	remoteFileContents, err := os.ReadFile(remoteFilePath)
	require.NoError(t, err)
	localFileContents, err := os.ReadFile(localFilePath)
	require.NoError(t, err)
	assert.Equal(t, localFileContents, remoteFileContents)
}

func TestDeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	remoteFilePath := "someRemoteFilePath"
	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().Remove(remoteFilePath).Return(nil)

	sftpManager := &SFTPManagerImpl{client: mockSFTPClient}

	err := sftpManager.DeleteFile(remoteFilePath)
	require.NoError(t, err)
}
