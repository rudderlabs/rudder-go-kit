package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type Option func(*keygen)

type keygen struct {
	saveTo *string
}

func SaveTo(saveTo string) Option {
	return func(k *keygen) {
		k.saveTo = &saveTo
	}
}

// NewRSAKeyPair generates a new private and public key pair
func NewRSAKeyPair(bitSize int, opts ...Option) (string, string, error) {
	var k keygen
	for _, opt := range opts {
		opt(&k)
	}

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	if k.saveTo == nil {
		return string(privateKeyBytes), string(publicKeyBytes), nil
	}

	privateKeyPath := filepath.Join(*k.saveTo, "id_rsa")
	if err := writeKeyToFile(privateKeyBytes, privateKeyPath); err != nil {
		return "", "", fmt.Errorf("failed to write private key to %q: %w", privateKeyPath, err)
	}

	publicKeyPath := filepath.Join(*k.saveTo, "id_rsa.pub")
	if err := writeKeyToFile(publicKeyBytes, publicKeyPath); err != nil {
		return "", "", fmt.Errorf("failed to write public key to %q: %w", publicKeyPath, err)
	}

	return privateKeyPath, publicKeyPath, nil
}

// generatePrivateKey creates an RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Private key in PEM format
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privateKey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privateKey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicRsaKey), nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	return os.WriteFile(saveFileTo, keyBytes, 0o600)
}
