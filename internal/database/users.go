package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrUserNotFound is returned when a requested user does not exist.
var ErrUserNotFound = errors.New("user not found")

// ErrModeratorNotFound is returned when a requested moderator does not exist.
var ErrModeratorNotFound = errors.New("moderator not found")

// User represents a forum user stored in the database.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    string
}

// CreateUser creates a new user and returns its database ID.
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

// GetUserByID retrieves a user by its database ID.
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

// GetUserByEmail retrieves a user by their email address.
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

// GetUserByUsername retrieves a user by their username.
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

// GetModerators retrieves all users with the moderator role.
func GetModerators(db *sql.DB) ([]User, error) {
	rows, err := db.Query(`
		SELECT
			id,
			username,
			email,
			password_hash,
			role,
			created_at
		FROM users
		WHERE role = 'moderator'
		ORDER BY username ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get moderators: %w", err)
	}
	defer rows.Close()

	var moderators []User

	for rows.Next() {
		var moderator User

		if err := rows.Scan(
			&moderator.ID,
			&moderator.Username,
			&moderator.Email,
			&moderator.PasswordHash,
			&moderator.Role,
			&moderator.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan moderator: %w", err)
		}

		moderators = append(moderators, moderator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate moderators: %w", err)
	}

	return moderators, nil
}

// UpdateUserRole updates the role assigned to a user.
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

// DemoteModerator changes a moderator's role back to a regular user.
func DemoteModerator(
	db *sql.DB,
	userID int64,
) error {
	result, err := db.Exec(`
		UPDATE users
		SET role = 'user'
		WHERE id = ?
		AND role = 'moderator'
	`, userID)
	if err != nil {
		return fmt.Errorf("demote moderator: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get demote moderator affected rows: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrModeratorNotFound
	}

	return nil
}