package database

import (
	"database/sql"
	"fmt"
)

type ReportView struct {
	ID               int64
	ReporterID       int64
	ReporterUsername string
	PostID           sql.NullInt64
	CommentID        sql.NullInt64
	Reason           string
	Status           string
	CreatedAt        string
	TargetType       string
	TargetTitle      string
	TargetURL        string
}

func GetPendingReportViews(
	db *sql.DB,
) ([]ReportView, error) {
	rows, err := db.Query(`
		SELECT
			r.id,
			r.reporter_id,
			reporter.username,
			r.post_id,
			r.comment_id,
			r.reason,
			r.status,
			r.created_at,
			CASE
				WHEN r.post_id IS NOT NULL
					THEN 'post'
				ELSE 'comment'
			END,
			CASE
				WHEN r.post_id IS NOT NULL
					THEN p.title
				ELSE 'Comment on ' || cp.title
			END,
			CASE
				WHEN r.post_id IS NOT NULL
					THEN '/posts/' || p.id
				ELSE '/posts/' || cp.id || '#comment-' || c.id
			END
		FROM reports r

		JOIN users reporter
			ON reporter.id = r.reporter_id

		LEFT JOIN posts p
			ON p.id = r.post_id

		LEFT JOIN comments c
			ON c.id = r.comment_id

		LEFT JOIN posts cp
			ON cp.id = c.post_id

		WHERE r.status = 'pending'

		ORDER BY r.created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf(
			"get pending report views: %w",
			err,
		)
	}
	defer rows.Close()

	var reports []ReportView

	for rows.Next() {
		var report ReportView

		if err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.ReporterUsername,
			&report.PostID,
			&report.CommentID,
			&report.Reason,
			&report.Status,
			&report.CreatedAt,
			&report.TargetType,
			&report.TargetTitle,
			&report.TargetURL,
		); err != nil {
			return nil, fmt.Errorf(
				"scan report view: %w",
				err,
			)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate report views: %w",
			err,
		)
	}

	return reports, nil
}
