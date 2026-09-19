package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrModeratorRequestNotFound = errors.New("moderator request not found")
	ErrModeratorRequestExists   = errors.New("pending moderator request already exists")
	ErrReportNotFound           = errors.New("report not found")
)

type ModeratorRequest struct {
	ID         int64
	UserID     int64
	Status     string
	CreatedAt  string
	ReviewedAt sql.NullString
	ReviewedBy sql.NullInt64
}

type ModeratorRequestView struct {
	ID        int64
	UserID    int64
	Username  string
	Email     string
	Status    string
	CreatedAt string
}

type Report struct {
	ID         int64
	ReporterID int64
	PostID     sql.NullInt64
	CommentID  sql.NullInt64
	Reason     string
	Status     string
	CreatedAt  string
	ReviewedAt sql.NullString
	ReviewedBy sql.NullInt64
}

func CreateModeratorRequest(db *sql.DB, userID int64) (int64, error) {
	var existingID int64

	err := db.QueryRow(`
		SELECT id
		FROM moderator_requests
		WHERE user_id = ?
		  AND status = 'pending'
		LIMIT 1
	`, userID).Scan(&existingID)

	if err == nil {
		return 0, ErrModeratorRequestExists
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("check pending moderator request: %w", err)
	}

	result, err := db.Exec(`
		INSERT INTO moderator_requests (user_id)
		VALUES (?)
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("create moderator request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created moderator request id: %w", err)
	}

	return id, nil
}

func GetModeratorRequestByID(
	db *sql.DB,
	requestID int64,
) (*ModeratorRequest, error) {
	request := &ModeratorRequest{}

	err := db.QueryRow(`
		SELECT
			id,
			user_id,
			status,
			created_at,
			reviewed_at,
			reviewed_by
		FROM moderator_requests
		WHERE id = ?
	`, requestID).Scan(
		&request.ID,
		&request.UserID,
		&request.Status,
		&request.CreatedAt,
		&request.ReviewedAt,
		&request.ReviewedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrModeratorRequestNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get moderator request by id: %w", err)
	}

	return request, nil
}

func GetPendingModeratorRequestByUser(
	db *sql.DB,
	userID int64,
) (*ModeratorRequest, error) {
	request := &ModeratorRequest{}

	err := db.QueryRow(`
		SELECT
			id,
			user_id,
			status,
			created_at,
			reviewed_at,
			reviewed_by
		FROM moderator_requests
		WHERE user_id = ?
		  AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(
		&request.ID,
		&request.UserID,
		&request.Status,
		&request.CreatedAt,
		&request.ReviewedAt,
		&request.ReviewedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrModeratorRequestNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get pending moderator request by user: %w", err)
	}

	return request, nil
}

func GetPendingModeratorRequests(
	db *sql.DB,
) ([]ModeratorRequest, error) {
	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			status,
			created_at,
			reviewed_at,
			reviewed_by
		FROM moderator_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get pending moderator requests: %w", err)
	}
	defer rows.Close()

	var requests []ModeratorRequest

	for rows.Next() {
		var request ModeratorRequest

		if err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.Status,
			&request.CreatedAt,
			&request.ReviewedAt,
			&request.ReviewedBy,
		); err != nil {
			return nil, fmt.Errorf("scan moderator request: %w", err)
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate moderator requests: %w", err)
	}

	return requests, nil
}

func GetPendingModeratorRequestViews(
	db *sql.DB,
) ([]ModeratorRequestView, error) {
	rows, err := db.Query(`
		SELECT
			mr.id,
			mr.user_id,
			u.username,
			u.email,
			mr.status,
			mr.created_at
		FROM moderator_requests mr
		JOIN users u
			ON u.id = mr.user_id
		WHERE mr.status = 'pending'
		ORDER BY mr.created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get pending moderator request views: %w", err)
	}
	defer rows.Close()

	var requests []ModeratorRequestView

	for rows.Next() {
		var request ModeratorRequestView

		if err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.Username,
			&request.Email,
			&request.Status,
			&request.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan moderator request view: %w", err)
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate moderator request views: %w", err)
	}

	return requests, nil
}

func ReviewModeratorRequest(
	db *sql.DB,
	requestID int64,
	status string,
	reviewerID int64,
) error {
	if status != "approved" && status != "rejected" {
		return errors.New("invalid moderator request status")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin moderator request review: %w", err)
	}

	defer tx.Rollback()

	var userID int64

	err = tx.QueryRow(`
		SELECT user_id
		FROM moderator_requests
		WHERE id = ?
		  AND status = 'pending'
	`, requestID).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrModeratorRequestNotFound
	}

	if err != nil {
		return fmt.Errorf("get moderator request user: %w", err)
	}

	result, err := tx.Exec(`
		UPDATE moderator_requests
		SET
			status = ?,
			reviewed_at = CURRENT_TIMESTAMP,
			reviewed_by = ?
		WHERE id = ?
		  AND status = 'pending'
	`, status, reviewerID, requestID)
	if err != nil {
		return fmt.Errorf("review moderator request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after reviewing moderator request: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrModeratorRequestNotFound
	}

	if status == "approved" {
		result, err = tx.Exec(`
			UPDATE users
			SET role = 'moderator'
			WHERE id = ?
			  AND role = 'user'
		`, userID)
		if err != nil {
			return fmt.Errorf("promote user to moderator: %w", err)
		}

		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return fmt.Errorf(
				"get affected rows after promoting moderator: %w",
				err,
			)
		}

		if rowsAffected == 0 {
			return ErrUserNotFound
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit moderator request review: %w", err)
	}

	return nil
}

func CreatePostReport(
	db *sql.DB,
	reporterID int64,
	postID int64,
	reason string,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO reports (
			reporter_id,
			post_id,
			reason
		)
		VALUES (?, ?, ?)
	`, reporterID, postID, reason)
	if err != nil {
		return 0, fmt.Errorf("create post report: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created post report id: %w", err)
	}

	return id, nil
}

func CreateCommentReport(
	db *sql.DB,
	reporterID int64,
	commentID int64,
	reason string,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO reports (
			reporter_id,
			comment_id,
			reason
		)
		VALUES (?, ?, ?)
	`, reporterID, commentID, reason)
	if err != nil {
		return 0, fmt.Errorf("create comment report: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created comment report id: %w", err)
	}

	return id, nil
}

func GetPendingReports(db *sql.DB) ([]Report, error) {
	rows, err := db.Query(`
		SELECT
			id,
			reporter_id,
			post_id,
			comment_id,
			reason,
			status,
			created_at,
			reviewed_at,
			reviewed_by
		FROM reports
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get pending reports: %w", err)
	}
	defer rows.Close()

	var reports []Report

	for rows.Next() {
		var report Report

		if err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.PostID,
			&report.CommentID,
			&report.Reason,
			&report.Status,
			&report.CreatedAt,
			&report.ReviewedAt,
			&report.ReviewedBy,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}

	return reports, nil
}

func ReviewReport(
	db *sql.DB,
	reportID int64,
	status string,
	reviewerID int64,
) error {
	if status != "resolved" && status != "rejected" {
		return errors.New("invalid report status")
	}

	result, err := db.Exec(`
		UPDATE reports
		SET
			status = ?,
			reviewed_at = CURRENT_TIMESTAMP,
			reviewed_by = ?
		WHERE id = ?
		  AND status = 'pending'
	`, status, reviewerID, reportID)
	if err != nil {
		return fmt.Errorf("review report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after reviewing report: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrReportNotFound
	}

	return nil
}
