package crypto_test

import (
	"bytes"
	"testing"

	"github.com/danzt/daas/api/internal/platform/crypto"
)

func makeKey(t *testing.T) []byte {
	t.Helper()
	// 32 bytes = AES-256 key.
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1) // deterministic but non-zero
	}
	return key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := makeKey(t)
	plaintext := []byte("super-secret-fiscal-api-key-12345")

	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must differ from plaintext")
	}

	decrypted, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	key := makeKey(t)
	plaintext := []byte("same plaintext")

	ct1, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}
	ct2, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	// Because nonces are random, two encryptions of the same plaintext
	// MUST produce different ciphertexts.
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of the same plaintext must produce different ciphertexts (nonce reuse)")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := makeKey(t)
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = 0xFF // different key
	}

	plaintext := []byte("sensitive data")
	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = crypto.Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key, got nil")
	}
}

func TestEncrypt_WrongKeySize(t *testing.T) {
	shortKey := make([]byte, 16) // AES-128 key, not 256
	_, err := crypto.Encrypt([]byte("test"), shortKey)
	if err == nil {
		t.Fatal("expected error for non-32-byte key")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := makeKey(t)
	ciphertext, err := crypto.Encrypt([]byte("important data"), key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Flip a byte in the ciphertext to simulate tampering.
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = crypto.Decrypt(ciphertext, key)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext, got nil")
	}
}
