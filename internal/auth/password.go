package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	// ErrEmptyPassword is returned when an empty password is provided for hashing.
	ErrEmptyPassword = errors.New("password cannot be empty")

	// ErrWrongPassword is returned when a password does not match its stored hash.
	ErrWrongPassword = errors.New("incorrect password")
)

// HashPassword hashes a non-empty password using bcrypt.
func HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcryptCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword compares a password with its stored bcrypt hash.
func CheckPassword(passwordHash, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrWrongPassword
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
	if err != nil {
		return ErrWrongPassword
	}

	return nil
}