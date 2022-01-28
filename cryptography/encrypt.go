package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// Encryption ..
type Encryption struct {
	bytes []byte
	block cipher.Block
}

// NewEncryptionClient returns an instance of the encryption client
func NewEncryptionClient(key string, bytes []byte) (*Encryption, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	return &Encryption{block: block, bytes: bytes}, nil
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
