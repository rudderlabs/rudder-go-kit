package sftp

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/sftp"
	mock_sftp "github.com/rudderlabs/rudder-go-kit/sftp/mock_sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSSH(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSSHClient := mock_sftp.NewMockSSHClient(ctrl)

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
			if tc.expectedError == nil {
				mockSSHClient.EXPECT().Dial(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ssh.Client{}, nil)
			}

			sshClient, err := NewSSHClient(tc.config, mockSSHClient)
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

func TestSFTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSFTPClient := mock_sftp.NewMockSFTPClient(ctrl)

	// Prepare mock response for CreateNew method
	mockSFTPClient.EXPECT().CreateNew(gomock.Any()).Return(&sftp.Client{}, nil)

	// Call the NewSFTPClient function with the given SSH client and mock SFTP client
	sftpClient, err := NewSFTPClient(&ssh.Client{}, mockSFTPClient)
	assert.Nil(t, err)
	assert.NotNil(t, sftpClient)
}
