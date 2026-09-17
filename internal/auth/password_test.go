package auth

import (
	"errors"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "StrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("expected password hash, got empty string")
	}

	if hash == password {
		t.Fatal("password was stored without hashing")
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Fatalf("correct password was rejected: %v", err)
	}
}

func TestHashPasswordRejectsEmptyPassword(t *testing.T) {
	tests := []string{
		"",
		" ",
		"\t",
		"\n",
	}

	for _, password := range tests {
		_, err := HashPassword(password)

		if !errors.Is(err, ErrEmptyPassword) {
			t.Fatalf(
				"expected ErrEmptyPassword for %q, got %v",
				password,
				err,
			)
		}
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("CorrectPassword123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	err = CheckPassword(hash, "WrongPassword123!")

	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf(
			"expected ErrWrongPassword, got %v",
			err,
		)
	}
}

func TestCheckPasswordRejectsEmptyPassword(t *testing.T) {
	hash, err := HashPassword("CorrectPassword123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	err = CheckPassword(hash, "")

	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf(
			"expected ErrWrongPassword, got %v",
			err,
		)
	}
}
