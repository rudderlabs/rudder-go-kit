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
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return string(bytes), nil
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
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			key, err := generateRandomString(int(tt.level / 8))
			require.NoError(t, err)

			plaintext := loremIpsumDolor
			ciphertext, err := encrypter.Encrypt(plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := encrypter.Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Decrypted data = %v, want %v", decrypted, plaintext)
			}
		})
	}
}
