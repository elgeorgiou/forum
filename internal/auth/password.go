package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrWrongPassword = errors.New("incorrect password")
)

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
