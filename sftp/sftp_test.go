package sftp

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rudderlabs/rudder-go-kit/sftp/mock_sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	localFilePath := "testdata/upload/local/test_file.json"
	remoteDir := "testdata/upload/remote"
	remoteFilePath := "testdata/upload/remote/test_file.json"

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	err := os.MkdirAll(remoteDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(remoteDir)

	tempFile, err := os.Create(remoteFilePath)
	require.NoError(t, err)

	mockSFTPClient.EXPECT().Create(remoteFilePath).Return(tempFile, nil)

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

	localDir := "testdata/download/local"
	localFilePath := "testdata/download/local/test_file.json"
	remoteFilePath := "testdata/download/remote/test_file.json"

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	remoteFile, err := os.Open(remoteFilePath)
	require.NoError(t, err)
	defer remoteFile.Close()

	err = os.MkdirAll(localDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(localDir)

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

	remoteFilePath := "testdata/remote/test_file.json"

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().Remove(remoteFilePath).Return(nil)

	sftpManager := &SFTPManagerImpl{client: mockSFTPClient}

	err := sftpManager.DeleteFile(remoteFilePath)
	require.NoError(t, err)
}
