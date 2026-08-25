package schwabdev

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}
	if key == nil {
		t.Fatal("GenerateKey() returned nil key")
	}
	// A Fernet key is 32 bytes; verify it's not all zeros
	var zero [32]byte
	if *key == zero {
		t.Fatal("GenerateKey() returned an all-zero key")
	}
}

func TestValidateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	encoded := key.Encode()

	decoded, err := ValidateKey(encoded)
	if err != nil {
		t.Fatalf("ValidateKey() returned error: %v", err)
	}
	if decoded == nil {
		t.Fatal("ValidateKey() returned nil key")
	}
	// The decoded key should match the original
	if *decoded != *key {
		t.Fatal("ValidateKey() decoded key does not match original")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	plaintext := "my secret token value"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}

	if ciphertext == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt() returned error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptNilKey(t *testing.T) {
	plaintext := "my secret token value"

	result, err := Encrypt(plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt() with nil key returned error: %v", err)
	}
	if result != plaintext {
		t.Fatalf("Encrypt() with nil key = %q, want %q", result, plaintext)
	}
}

func TestDecryptNilKeyUnencrypted(t *testing.T) {
	plaintext := "my secret token value"

	result, err := Decrypt(plaintext, nil)
	if err != nil {
		t.Fatalf("Decrypt() with nil key on unencrypted text returned error: %v", err)
	}
	if result != plaintext {
		t.Fatalf("Decrypt() with nil key on unencrypted text = %q, want %q", result, plaintext)
	}
}

func TestDecryptNilKeyEncrypted(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	plaintext := "my secret token value"
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}

	result, err := Decrypt(ciphertext, nil)
	if err != ErrDecryptionFailed {
		t.Fatalf("Decrypt() with nil key on encrypted text returned error %v, want ErrDecryptionFailed", err)
	}
	if result != "" {
		t.Fatalf("Decrypt() with nil key on encrypted text = %q, want empty string", result)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	plaintext := "my secret token value"
	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}

	result, err := Decrypt(ciphertext, key2)
	if err != ErrDecryptionFailed {
		t.Fatalf("Decrypt() with wrong key returned error %v, want ErrDecryptionFailed", err)
	}
	if result != "" {
		t.Fatalf("Decrypt() with wrong key = %q, want empty string", result)
	}
}

func TestEncodeKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	encoded := EncodeKey(key)
	if encoded == "" {
		t.Fatal("EncodeKey() returned empty string")
	}

	// Verify it's valid base64
	_, err = base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("EncodeKey() result is not valid URL-safe base64: %v", err)
	}
}

func TestEncodeKeyNil(t *testing.T) {
	result := EncodeKey(nil)
	if result != "" {
		t.Fatalf("EncodeKey(nil) = %q, want empty string", result)
	}
}

func TestEncryptReturnsPrefix(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() returned error: %v", err)
	}

	plaintext := "my secret token value"
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}

	if !strings.HasPrefix(ciphertext, EncryptionPrefix) {
		t.Fatalf("Encrypt() result %q does not have %q prefix", ciphertext, EncryptionPrefix)
	}
}
