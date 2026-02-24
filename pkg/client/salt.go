package salt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// EncryptEnv reads the .env file, encrypts it using AES-GCM, and writes .env.enc.
// It uses SHA-256 to derive a 32-byte key from the provided passphrase and salt.
func EncryptEnv(passphrase, salt string) error {
	const inputFile = ".env"
	const outputFile = ".env.enc"

	// Read the plaintext .env file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", inputFile, err)
	}

	// Derive a 32-byte key for AES-256 using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(passphrase + salt))
	key := hasher.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Galois/Counter Mode (GCM) provides both confidentiality and authenticity
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create a unique nonce for this encryption operation
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Seal the data. The resulting slice contains [nonce][ciphertext+tag]
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write the encrypted file with restricted 0600 permissions
	err = os.WriteFile(outputFile, ciphertext, 0600)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", outputFile, err)
	}

	return nil
}
