package schwabdev

import (
	"errors"
	"strings"

	"github.com/fernet/fernet-go"
)

const encPrefix = "enc:"

// Encrypt encrypts plaintext using the provided Fernet key.
// If key is nil, returns plaintext unchanged (no encryption mode).
// Otherwise, encrypts and returns with "enc:" prefix.
func Encrypt(plaintext string, key *fernet.Key) (string, error) {
	// No encryption key provided - return plaintext
	if key == nil {
		return plaintext, nil
	}

	// Encrypt the plaintext
	token, err := fernet.EncryptAndSign([]byte(plaintext), key)
	if err != nil {
		return "", err
	}

	// Return with prefix
	return encPrefix + string(token), nil
}

// Decrypt decrypts ciphertext using the provided Fernet key.
// If ciphertext doesn't have "enc:" prefix, returns it unchanged.
// If ciphertext has prefix but key is nil, returns error.
// Otherwise, removes prefix and decrypts.
func Decrypt(ciphertext string, key *fernet.Key) (string, error) {
	// Not encrypted - return as-is
	if !strings.HasPrefix(ciphertext, encPrefix) {
		return ciphertext, nil
	}

	// Encrypted but no key - error
	if key == nil {
		return "", errors.New("cannot decrypt token: no encryption key provided")
	}

	// Remove prefix and decrypt
	token := ciphertext[len(encPrefix):]
	message := fernet.VerifyAndDecrypt([]byte(token), 0, []*fernet.Key{key})
	if message == nil {
		return "", errors.New("decryption failed: invalid token or key")
	}

	return string(message), nil
}

// GenerateKey generates a new Fernet encryption key.
// Returns the generated key or an error if generation fails.
func GenerateKey() (*fernet.Key, error) {
	key := new(fernet.Key)
	err := key.Generate()
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ValidateKey validates and decodes a Fernet key from a string.
// The key can be in hexadecimal, standard base64, or URL-safe base64 format.
// Returns the decoded key or an error if validation fails.
func ValidateKey(keyString string) (*fernet.Key, error) {
	if keyString == "" {
		return nil, errors.New("empty key")
	}

	key, err := fernet.DecodeKey(keyString)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// EncodeKey encodes a Fernet key to URL-safe base64 string.
// Returns the encoded key string.
func EncodeKey(key *fernet.Key) string {
	if key == nil {
		return ""
	}
	return key.Encode()
}
