package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrOAuthAccountNotFound is returned when a requested OAuth account does not exist.
var ErrOAuthAccountNotFound = errors.New("oauth account not found")

// OAuthAccount represents an external OAuth account linked to a forum user.
type OAuthAccount struct {
	ID             int64
	UserID         int64
	Provider       string
	ProviderUserID string
	CreatedAt      string
}

// CreateOAuthAccount creates a new OAuth account linked to a user and returns its database ID.
func CreateOAuthAccount(
	db *sql.DB,
	userID int64,
	provider string,
	providerUserID string,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO oauth_accounts (
			user_id,
			provider,
			provider_user_id
		)
		VALUES (?, ?, ?)
	`, userID, provider, providerUserID)
	if err != nil {
		return 0, fmt.Errorf("create oauth account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created oauth account id: %w", err)
	}

	return id, nil
}

// GetOAuthAccount retrieves an OAuth account by its provider and provider user ID.
func GetOAuthAccount(
	db *sql.DB,
	provider string,
	providerUserID string,
) (OAuthAccount, error) {
	var account OAuthAccount

	err := db.QueryRow(`
		SELECT
			id,
			user_id,
			provider,
			provider_user_id,
			created_at
		FROM oauth_accounts
		WHERE provider = ?
		  AND provider_user_id = ?
	`, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return OAuthAccount{}, ErrOAuthAccountNotFound
	}

	if err != nil {
		return OAuthAccount{}, fmt.Errorf("get oauth account: %w", err)
	}

	return account, nil
}

// GetOAuthAccountsByUser retrieves all OAuth accounts linked to a user.
func GetOAuthAccountsByUser(
	db *sql.DB,
	userID int64,
) ([]OAuthAccount, error) {
	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			provider,
			provider_user_id,
			created_at
		FROM oauth_accounts
		WHERE user_id = ?
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get oauth accounts by user: %w", err)
	}
	defer rows.Close()

	var accounts []OAuthAccount

	for rows.Next() {
		var account OAuthAccount

		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oauth account: %w", err)
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oauth accounts: %w", err)
	}

	return accounts, nil
}

// DeleteOAuthAccount deletes an OAuth account linked to a user for the specified provider.
func DeleteOAuthAccount(
	db *sql.DB,
	userID int64,
	provider string,
) error {
	result, err := db.Exec(`
		DELETE FROM oauth_accounts
		WHERE user_id = ?
		  AND provider = ?
	`, userID, provider)
	if err != nil {
		return fmt.Errorf("delete oauth account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after deleting oauth account: %w", err)
	}

	if rowsAffected == 0 {
		return ErrOAuthAccountNotFound
	}

	return nil
}