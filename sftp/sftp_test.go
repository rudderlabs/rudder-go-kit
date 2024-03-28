package sftp

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	mock_sftp "github.com/rudderlabs/rudder-go-kit/sftp/mock_sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureSSHClient(t *testing.T) {

	// Read private key
	privateKey, err := os.ReadFile("./testdata/ssh/test_key")
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

	// Mock SFTP client
	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(nil)

	err := mockSFTPClient.UploadFile("someLocalFilePath", "someRemoteDir")
	require.NoError(t, err)
}

func TestDownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().DownloadFile(gomock.Any(), gomock.Any()).Return(nil)

	err := mockSFTPClient.DownloadFile("someRemotePath", "someRemoteDir")
	require.NoError(t, err)
}

func TestDeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock SFTP client
	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)
	mockSFTPClient.EXPECT().DeleteFile(gomock.Any()).Return(nil)

	err := mockSFTPClient.DeleteFile("someRemotePath")
	require.NoError(t, err)
}
