package session

import (
	"crypto/rand"
	"encoding/base64"
)

const sessionIDBytes = 32

// NewID returns a cryptographically random, URL-safe session token.
func NewID() (string, error) {
	random := make([]byte, sessionIDBytes)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}
