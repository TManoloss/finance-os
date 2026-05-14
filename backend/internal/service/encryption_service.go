package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptionService fornece métodos para criptografar e descriptografar dados usando AES-GCM.
type EncryptionService struct {
	key []byte
}

// NewEncryptionService cria uma nova instância do EncryptionService com a chave fornecida.
// A chave deve ter 32 bytes para AES-256. Aceita também chave em base64 (44 chars).
func NewEncryptionService(key string) (*EncryptionService, error) {
	var keyBytes []byte
	var err error

	if len(key) == 32 {
		keyBytes = []byte(key)
	} else if len(key) == 44 {
		keyBytes, err = base64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, errors.New("invalid base64 encryption key")
		}
		if len(keyBytes) != 32 {
			return nil, errors.New("decoded encryption key must be 32 bytes")
		}
	} else {
		return nil, errors.New("encryption key must be 32 bytes or 44 characters base64")
	}

	return &EncryptionService{key: keyBytes}, nil
}

// Encrypt criptografa uma string usando AES-GCM e retorna o resultado em base64.
// O nonce é incluído no início do ciphertext.
func (s *EncryptionService) Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal anexa o ciphertext ao nonce e retorna o resultado.
	// O primeiro argumento (dst) é onde o resultado será anexado.
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt descriptografa uma string em base64 e retorna o texto original.
func (s *EncryptionService) Decrypt(cipherTextBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
