package encrypt

import (
	"fmt"
	"strings"
)

// EncryptionAlgorithm is the interface that wraps the encryption algorithm method.
type EncryptionAlgorithm int

func (e EncryptionAlgorithm) String() string {
	switch e {
	case EncryptionAlgoAESGCM:
		return "aes-gcm"
	default:
		return ""
	}
}

// EncryptionLevel is the interface that wraps the encryption level method.
type EncryptionLevel int

func (e EncryptionLevel) String() string {
	switch e {
	case EncryptionLevelAES128, EncryptionLevelAES192, EncryptionLevelAES256:
		return fmt.Sprintf("%d", e)
	default:
		return ""
	}
}

func NewSettings(algo, level string) (EncryptionAlgorithm, EncryptionLevel, error) {
	switch algo {
	case "aes-gcm":
		switch level {
		case "128":
			return EncryptionAlgoAESGCM, EncryptionLevelAES128, nil
		case "192":
			return EncryptionAlgoAESGCM, EncryptionLevelAES192, nil
		case "256":
			return EncryptionAlgoAESGCM, EncryptionLevelAES256, nil
		default:
			return 0, 0, fmt.Errorf("unknown encryption level for %s: %s", algo, level)
		}
	default:
		return 0, 0, fmt.Errorf("unknown encryption algorithm: %s", algo)
	}
}

var (
	EncryptionAlgoAESGCM  = EncryptionAlgorithm(1)
	EncryptionLevelAES128 = EncryptionLevel(128)
	EncryptionLevelAES192 = EncryptionLevel(192)
	EncryptionLevelAES256 = EncryptionLevel(256)
)

func New(algo EncryptionAlgorithm, level EncryptionLevel) (*Encryptor, error) {
	var err error
	algo, level, err = NewSettings(algo.String(), level.String())
	if err != nil {
		return nil, err
	}

	switch algo {
	case EncryptionAlgoAESGCM:
		return &Encryptor{encryptionAESGCM: &encryptionAESGCM{level: int(level)}}, nil
	default:
		return nil, fmt.Errorf("unknown encryption algorithm: %d", algo)
	}
}

type Encryptor struct {
	*encryptionAESGCM
}

func (e *Encryptor) Encrypt(src []byte, key string) ([]byte, error) {
	if e.encryptionAESGCM != nil {
		return e.encryptionAESGCM.Encrypt(src, key)
	}
	return nil, fmt.Errorf("no encryption method available")
}

func (e *Encryptor) Decrypt(src []byte, key string) ([]byte, error) {
	if e.encryptionAESGCM != nil {
		return e.encryptionAESGCM.Decrypt(src, key)
	}
	return nil, fmt.Errorf("no decryption method available")
}

// SerializeSettings converts the EncryptionAlgorithm and EncryptionLevel to a string.
func SerializeSettings(algo EncryptionAlgorithm, level EncryptionLevel) string {
	return fmt.Sprintf("%s:%s", algo.String(), level.String())
}

// DeserializeSettings converts a string to EncryptionAlgorithm and EncryptionLevel.
func DeserializeSettings(settings string) (EncryptionAlgorithm, EncryptionLevel, error) {
	parts := strings.Split(settings, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid settings format")
	}
	return NewSettings(parts[0], parts[1])
}
