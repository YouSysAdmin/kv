package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// AES holds the key and data for encryption or decryption.
type AES struct {
	key  string
	data string
}

// NewAES creates a new AES instance with the provided key and data.
func NewAES(key, data string) *AES {
	return &AES{
		key:  key,
		data: data,
	}
}

// AESKeySize represents the size of an AES key in bytes.
type AESKeySize int

const (
	// AES128 represents a 128-bit AES key (16 bytes).
	AES128 AESKeySize = 16
	// AES192 represents a 192-bit AES key (24 bytes).
	AES192 AESKeySize = 24
	// AES256 represents a 256-bit AES key (32 bytes).
	AES256 AESKeySize = 32
)

const (
	// PrefixAES128 is the prefix used to identify AES-128 encrypted strings.
	PrefixAES128 string = "aes128:"
	// PrefixAES192 is the prefix used to identify AES-192 encrypted strings.
	PrefixAES192 string = "aes192:"
	// PrefixAES256 is the prefix used to identify AES-256 encrypted strings.
	PrefixAES256 string = "aes256:"
)

// Prefix returns the prefix associated with the AES key size.
func (k AESKeySize) Prefix() (string, error) {
	switch k {
	case AES128:
		return PrefixAES128, nil
	case AES192:
		return PrefixAES192, nil
	case AES256:
		return PrefixAES256, nil
	default:
		return "", fmt.Errorf("%w: unsupported key length %d", ErrorInvalidKeyLength, k)
	}
}

// IsValid checks if the AES key size is valid.
func (k AESKeySize) IsValid() bool {
	return k == AES128 || k == AES192 || k == AES256
}

var (
	ErrorChipperTooShort     = errors.New("ciphertext too short")
	ErrorInvalidKey          = errors.New("invalid encryption key")
	ErrorGCMCreationFailed   = errors.New("failed to create GCM cipher mode")
	ErrorNonceReadFailed     = errors.New("failed to read nonce")
	ErrorBase64Decode        = errors.New("failed to decode base64 ciphertext")
	ErrorDecryptionFailed    = errors.New("decryption failed")
	ErrorInvalidKeyLength    = errors.New("invalid AES key length")
	ErrorKeyGenerationFailed = errors.New("failed to generate AES key")
)

// Encrypt encrypts plaintext using AES-GCM and returns a base64-encoded ciphertext with prefix.
func (a *AES) Encrypt() (string, error) {
	if err := ValidateAESKey(a.key); err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(a.key))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorInvalidKey, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorGCMCreationFailed, err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: %v", ErrorNonceReadFailed, err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(a.data), nil)
	prefix := fmt.Sprintf("aes%d:", len(a.key)*8)
	return prefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-GCM ciphertext string.
func (a *AES) Decrypt() (string, error) {
	if err := ValidateAESKey(a.key); err != nil {
		return "", err
	}

	prefix, err := AESKeySize(len(a.key)).Prefix()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(a.data, PrefixAES128) ||
		strings.HasPrefix(a.data, PrefixAES192) ||
		strings.HasPrefix(a.data, PrefixAES256) {
		if !strings.HasPrefix(a.data, prefix) {
			return "", fmt.Errorf("%w: expected prefix %q not found", ErrorInvalidKey, prefix)
		}
		a.data = a.data[len(prefix):]
	}

	ciphertext, err := base64.StdEncoding.DecodeString(a.data)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorBase64Decode, err)
	}

	block, err := aes.NewCipher([]byte(a.key))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorInvalidKey, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorGCMCreationFailed, err)
	}

	if len(ciphertext) < aesGCM.NonceSize() {
		return "", fmt.Errorf("%w: ciphertext length %d is less than nonce size %d", ErrorChipperTooShort, len(ciphertext), aesGCM.NonceSize())
	}

	nonce, ciphertext := ciphertext[:aesGCM.NonceSize()], ciphertext[aesGCM.NonceSize():]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrorDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// GenerateRandomAESKey generates a random AES key of the given bit length.
func GenerateRandomAESKey(bits AESKeySize) (string, error) {
	if !bits.IsValid() {
		return "", fmt.Errorf("%w: %d", ErrorInvalidKeyLength, bits)
	}
	keyLen := int(bits)

	rawKey := make([]byte, keyLen)
	if _, err := io.ReadFull(rand.Reader, rawKey); err != nil {
		return "", fmt.Errorf("%w: %v", ErrorKeyGenerationFailed, err)
	}

	b64str := base64.StdEncoding.EncodeToString(rawKey)
	b64str = strings.TrimRight(b64str, "=")

	if len(b64str) > keyLen {
		b64str = b64str[:keyLen]
	}
	return b64str, nil
}

// ValidateAESKey checks if the key length is valid for AES.
func ValidateAESKey(key string) error {
	if AESKeySize(len(key)).IsValid() {
		return nil
	}
	return fmt.Errorf("%w: got %d chars", ErrorInvalidKeyLength, len(key))
}
