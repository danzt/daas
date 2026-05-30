// Package crypto provides cryptographic utilities for the DaaS platform.
// All encryption uses AES-256-GCM (authenticated encryption with associated data).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// ErrDecryptionFailed is returned when decryption fails, typically due to
// a wrong key, tampered ciphertext, or invalid nonce.
var ErrDecryptionFailed = errors.New("decryption failed: invalid key or corrupted ciphertext")

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
//
// Key must be exactly 32 bytes (256-bit). Use base64.StdEncoding.DecodeString
// to decode keys stored as base64 env vars.
//
// Output format: [12-byte nonce][ciphertext+tag]
// The nonce is prepended to the ciphertext so Decrypt can extract it.
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES-256-GCM requires a 32-byte key, got %d bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM wrapper: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes for GCM
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends ciphertext+tag after nonce.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts AES-256-GCM ciphertext produced by Encrypt.
//
// Key must be exactly 32 bytes. Returns ErrDecryptionFailed if the key is
// wrong or the ciphertext has been tampered with (GCM authentication fails).
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES-256-GCM requires a 32-byte key, got %d bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM wrapper: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		// GCM authentication failure — wrong key or tampered ciphertext.
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}
