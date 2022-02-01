package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"os"

	"github.com/LF-Engineering/insights-datasource-shared/aws/ssm"
)

// Encryption ..
type Encryption struct {
	bytes []byte
	block cipher.Block
}

// NewEncryptionClient returns an instance of the encryption client
func NewEncryptionClient() (*Encryption, error) {
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return nil, errors.New("ENCRYPTION_KEY is empty")
	}

	encryptionByteString := os.Getenv("ENCRYPTION_BYTES")
	if encryptionByteString == "" {
		return nil, errors.New("ENCRYPTION_BYTES is empty")
	}

	c, err := ssm.NewSSMClient()
	if err != nil {
		return nil, err
	}

	p1 := c.Param(encryptionKey, true, false, "secureString", "secureString", "")
	valueKey, err := p1.GetValue()
	if err != nil {
		return nil, err
	}

	p2 := c.Param(encryptionByteString, true, false, "secureString", "secureString", "")
	valueBytes, err := p2.GetValue()
	if err != nil {
		return nil, err
	}

	encryptionByte := []byte(valueBytes)

	block, err := aes.NewCipher([]byte(valueKey))
	if err != nil {
		return nil, err
	}
	return &Encryption{block: block, bytes: encryptionByte}, nil
}

// Encrypt returns a ciphertext from a plaintext
func (e Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(e.block, e.bytes)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt returns a plaintext from a ciphertext
func (e Encryption) Decrypt(text string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(e.block, e.bytes)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}
