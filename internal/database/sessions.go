package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt string
	CreatedAt string
}

func CreateSession(
	db *sql.DB,
	id string,
	userID int64,
	expiresAt string,
) error {
	_, err := db.Exec(`
		INSERT INTO sessions (
			id,
			user_id,
			expires_at
		)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id)
		DO UPDATE SET
			id = excluded.id,
			expires_at = excluded.expires_at,
			created_at = CURRENT_TIMESTAMP
	`,
		id,
		userID,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf(
			"create session: %w",
			err,
		)
	}

	return nil
}

func GetSessionByID(
	db *sql.DB,
	id string,
) (Session, error) {
	var session Session

	err := db.QueryRow(`
		SELECT
			id,
			user_id,
			expires_at,
			created_at
		FROM sessions
		WHERE id = ?
	`, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}

	if err != nil {
		return Session{}, fmt.Errorf(
			"get session by id: %w",
			err,
		)
	}

	return session, nil
}

func GetSessionByUserID(
	db *sql.DB,
	userID int64,
) (Session, error) {
	var session Session

	err := db.QueryRow(`
		SELECT
			id,
			user_id,
			expires_at,
			created_at
		FROM sessions
		WHERE user_id = ?
	`, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}

	if err != nil {
		return Session{}, fmt.Errorf(
			"get session by user id: %w",
			err,
		)
	}

	return session, nil
}

func DeleteSessionByID(
	db *sql.DB,
	id string,
) error {
	result, err := db.Exec(`
		DELETE FROM sessions
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf(
			"delete session by id: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after deleting session: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func DeleteSessionByUserID(
	db *sql.DB,
	userID int64,
) error {
	result, err := db.Exec(`
		DELETE FROM sessions
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf(
			"delete session by user id: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after deleting session: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func DeleteExpiredSessions(db *sql.DB) error {
	now := time.Now().
		UTC().
		Format(time.RFC3339)

	_, err := db.Exec(`
		DELETE FROM sessions
		WHERE expires_at <= ?
	`, now)
	if err != nil {
		return fmt.Errorf(
			"delete expired sessions: %w",
			err,
		)
	}

	return nil
}
