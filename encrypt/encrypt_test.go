package encrypt

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

var loremIpsumDolor = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`)

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func Test_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		algo  EncryptionAlgorithm
		level EncryptionLevel
	}{
		{EncryptionAlgoAESCFB, EncryptionLevelAES128},
		{EncryptionAlgoAESCFB, EncryptionLevelAES192},
		{EncryptionAlgoAESCFB, EncryptionLevelAES256},

		{EncryptionAlgoAESGCM, EncryptionLevelAES128},
		{EncryptionAlgoAESGCM, EncryptionLevelAES192},
		{EncryptionAlgoAESGCM, EncryptionLevelAES256},
	}

	for _, tt := range tests {
		t.Run(tt.algo.String()+"_"+tt.level.String(), func(t *testing.T) {
			encrypter, err := New(tt.algo, tt.level)
			require.NoError(t, err)

			key, err := generateRandomString(int(tt.level / 8))
			require.NoError(t, err)

			plaintext := loremIpsumDolor
			_, err = encrypter.Encrypt(plaintext, key[:len(key)-1])
			require.Error(t, err)

			ciphertext, err := encrypter.Encrypt(plaintext, key)
			require.NoError(t, err)

			decrypted, err := encrypter.Decrypt(ciphertext, key)
			require.NoError(t, err)

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Decrypted data = %v, want %v", decrypted, plaintext)
			}
		})
	}
}

func Test_SerializeSettings(t *testing.T) {
	tests := []struct {
		algo   EncryptionAlgorithm
		level  EncryptionLevel
		expect string
	}{
		{EncryptionAlgoAESCFB, EncryptionLevelAES128, "cfb:aes-128"},
		{EncryptionAlgoAESCFB, EncryptionLevelAES192, "cfb:aes-192"},
		{EncryptionAlgoAESCFB, EncryptionLevelAES256, "cfb:aes-256"},
		{EncryptionAlgoAESGCM, EncryptionLevelAES128, "gcm:aes-128"},
		{EncryptionAlgoAESGCM, EncryptionLevelAES192, "gcm:aes-192"},
		{EncryptionAlgoAESGCM, EncryptionLevelAES256, "gcm:aes-256"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := SerializeSettings(tt.algo, tt.level)
			require.Equal(t, tt.expect, result)
		})
	}
}

func Test_DeserializeSettings(t *testing.T) {
	tests := []struct {
		input  string
		algo   EncryptionAlgorithm
		level  EncryptionLevel
		hasErr bool
	}{
		{"cfb:aes-128", EncryptionAlgoAESCFB, EncryptionLevelAES128, false},
		{"cfb:aes-192", EncryptionAlgoAESCFB, EncryptionLevelAES192, false},
		{"cfb:aes-256", EncryptionAlgoAESCFB, EncryptionLevelAES256, false},
		{"gcm:aes-128", EncryptionAlgoAESGCM, EncryptionLevelAES128, false},
		{"gcm:aes-192", EncryptionAlgoAESGCM, EncryptionLevelAES192, false},
		{"gcm:aes-256", EncryptionAlgoAESGCM, EncryptionLevelAES256, false},
		{"invalid:settings", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			algo, level, err := DeserializeSettings(tt.input)
			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.algo, algo)
				require.Equal(t, tt.level, level)
			}
		})
	}
}
