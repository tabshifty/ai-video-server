package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const passwordVaultCipherPrefix = "v1:"

// PasswordVaultCipher encrypts and decrypts admin password vault secrets.
type PasswordVaultCipher struct {
	aead cipher.AEAD
}

func NewPasswordVaultCipher(secret string) (*PasswordVaultCipher, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("password vault key is required")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create password vault cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create password vault gcm: %w", err)
	}
	return &PasswordVaultCipher{aead: aead}, nil
}

func (c *PasswordVaultCipher) Encrypt(plain string) (string, error) {
	if c == nil || c.aead == nil {
		return "", fmt.Errorf("password vault cipher is not configured")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate password vault nonce: %w", err)
	}
	payload := c.aead.Seal(nonce, nonce, []byte(plain), nil)
	return passwordVaultCipherPrefix + base64.StdEncoding.EncodeToString(payload), nil
}

func (c *PasswordVaultCipher) Decrypt(ciphertext string) (string, error) {
	if c == nil || c.aead == nil {
		return "", fmt.Errorf("password vault cipher is not configured")
	}
	encoded, ok := strings.CutPrefix(strings.TrimSpace(ciphertext), passwordVaultCipherPrefix)
	if !ok {
		return "", fmt.Errorf("unsupported password vault ciphertext")
	}
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode password vault ciphertext: %w", err)
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "", fmt.Errorf("invalid password vault ciphertext")
	}
	plain, err := c.aead.Open(nil, payload[:nonceSize], payload[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt password vault ciphertext: %w", err)
	}
	return string(plain), nil
}
