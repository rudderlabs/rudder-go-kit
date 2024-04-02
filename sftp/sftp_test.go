package sftp

import (
	"fmt"
	"os"
	"testing"

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
