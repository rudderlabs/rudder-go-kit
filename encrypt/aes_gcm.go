package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type encryptionAESGCM struct {
	level int
}

func (e *encryptionAESGCM) Encrypt(src []byte, key string) ([]byte, error) {
	if len(key) != e.level/8 {
		return nil, fmt.Errorf("key length must be %d bytes", e.level/8)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, src, nil)
	return ciphertext, nil
}

func (e *encryptionAESGCM) Decrypt(src []byte, key string) ([]byte, error) {
	if len(key) != e.level/8 {
		return nil, fmt.Errorf("key length must be %d bytes", e.level/8)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(src) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := src[:nonceSize], src[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
