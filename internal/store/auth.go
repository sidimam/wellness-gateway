package store

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

const pbkdfIterations = 600_000

// HashPassword restituisce (hash, salt) in esadecimale.
func HashPassword(password string) (string, string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}
	h, err := pbkdf2.Key(sha256.New, password, salt, pbkdfIterations, 32)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(h), hex.EncodeToString(salt), nil
}

// VerifyPassword confronta la password con hash e salt.
func VerifyPassword(password, hashHex, saltHex string) bool {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, pbkdfIterations, 32)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// NewToken genera un token casuale (32 byte, esadecimale).
func NewToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewID genera un identificatore breve.
func NewID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
