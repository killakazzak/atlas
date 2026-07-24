package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher hashes and verifies plaintext passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	// Compare returns nil when password matches hash, ErrInvalidCredentials otherwise.
	Compare(hash, password string) error
}

// BcryptHasher implements PasswordHasher using bcrypt at the default cost.
type BcryptHasher struct{}

// Hash returns a bcrypt hash of password.
func (BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare returns nil when password matches hash, ErrInvalidCredentials otherwise.
func (BcryptHasher) Compare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
