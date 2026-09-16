package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    string
}

func CreateUser(
	db *sql.DB,
	username string,
	email string,
	passwordHash string,
) (int64, error) {
	result, err := db.Exec(
		`
		INSERT INTO users (
			username,
			email,
			password_hash
		)
		VALUES (?, ?, ?)
		`,
		username,
		email,
		passwordHash,
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created user id: %w", err)
	}

	return id, nil
}

func GetUserByID(db *sql.DB, id int64) (*User, error) {
	user := &User{}

	err := db.QueryRow(
		`
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE id = ?
		`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}

	err := db.QueryRow(
		`
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE email = ?
		`,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}

	err := db.QueryRow(
		`
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE username = ?
		`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func UpdateUserRole(
	db *sql.DB,
	userID int64,
	role string,
) error {
	result, err := db.Exec(
		`
		UPDATE users
		SET role = ?
		WHERE id = ?
		`,
		role,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
