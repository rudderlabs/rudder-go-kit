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
		t.Run("integrity check", func(t *testing.T) {
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

					t.Log("to test integrity check properties manipulate the ciphertext")
					ciphertext[0] = ciphertext[0] + 1
					decrypted, err := encrypter.Decrypt(ciphertext, key)

					require.Error(t, err, "decryption should fail, instead got %q", decrypted)
					require.Empty(t, decrypted)
				})
			}
		})
		t.Run("invalid key decryption", func(t *testing.T) {
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

					anotherKey, err := generateRandomString(int(tt.level / 8))
					require.NoError(t, err)

					t.Log("use a different key to decrypt the ciphertext")
					decrypted, err := encrypter.Decrypt(ciphertext, anotherKey)
					require.Error(t, err, "decryption should fail, instead got %q", decrypted)
					require.Empty(t, decrypted)
				})
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
		{EncryptionAlgoAESGCM, EncryptionLevelAES128, "aes-gcm:128"},
		{EncryptionAlgoAESGCM, EncryptionLevelAES192, "aes-gcm:192"},
		{EncryptionAlgoAESGCM, EncryptionLevelAES256, "aes-gcm:256"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := SerializeSettings(tt.algo, tt.level)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestEncryptionAESGCM_Encrypt(t *testing.T) {
	tests := []struct {
		level  EncryptionLevel
		keyLen int
	}{
		{EncryptionLevelAES128, 16},
		{EncryptionLevelAES192, 24},
		{EncryptionLevelAES256, 32},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			enc := &encryptionAESGCM{level: int(tt.level)}
			key, err := generateRandomString(tt.keyLen)
			require.NoError(t, err)

			plaintext := loremIpsumDolor
			ciphertext1, err := enc.Encrypt(plaintext, key)
			require.NoError(t, err)

			decrypted1, err := enc.Decrypt(ciphertext1, key)
			require.NoError(t, err)

			if !bytes.Equal(decrypted1, plaintext) {
				t.Errorf("Decrypted data = %v, want %v", decrypted1, plaintext)
			}

			// encrypt same cipher text with same key should produce different cipher text
			ciphertext2, err := enc.Encrypt(plaintext, key)
			require.NoError(t, err)

			decrypted2, err := enc.Decrypt(ciphertext2, key)
			require.NoError(t, err)

			if !bytes.Equal(decrypted2, plaintext) {
				t.Errorf("Decrypted data = %v, want %v", decrypted2, plaintext)
			}

			if bytes.Equal(ciphertext2, ciphertext1) {
				t.Errorf("cipher text should be different everytime = %v, want %v", ciphertext1, ciphertext2)
			}
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
		{"aes-gcm:128", EncryptionAlgoAESGCM, EncryptionLevelAES128, false},
		{"aes-gcm:192", EncryptionAlgoAESGCM, EncryptionLevelAES192, false},
		{"aes-gcm:256", EncryptionAlgoAESGCM, EncryptionLevelAES256, false},
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
