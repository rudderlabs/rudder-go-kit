package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type encryptionAESCFB struct {
	level int
}

func (e *encryptionAESCFB) Encrypt(src []byte, key string) ([]byte, error) {
	if len(key) != e.level/8 {
		return nil, fmt.Errorf("key length must be %d bytes", e.level/8)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(src))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], src)

	return ciphertext, nil
}

func (e *encryptionAESCFB) Decrypt(src []byte, key string) ([]byte, error) {
	if len(key) != e.level/8 {
		return nil, fmt.Errorf("key length must be %d bytes", e.level/8)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	if len(src) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := src[:aes.BlockSize]
	src = src[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(src, src)

	return src, nil
}
