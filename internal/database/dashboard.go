package database

import (
	"database/sql"
	"fmt"
)

// DashboardStats represents the summary statistics displayed on the dashboard.
type DashboardStats struct {
	PendingReports int
	TotalPosts     int
	TotalComments  int
	TotalUsers     int
}

// GetDashboardStats retrieves the summary statistics for the dashboard.
func GetDashboardStats(
	db *sql.DB,
) (DashboardStats, error) {
	var stats DashboardStats

	err := db.QueryRow(`
		SELECT
			(SELECT COUNT(*)
			 FROM reports
			 WHERE status = 'pending'),

			(SELECT COUNT(*)
			 FROM posts),

			(SELECT COUNT(*)
			 FROM comments),

			(SELECT COUNT(*)
			 FROM users)
	`).Scan(
		&stats.PendingReports,
		&stats.TotalPosts,
		&stats.TotalComments,
		&stats.TotalUsers,
	)
	if err != nil {
		return DashboardStats{},
			fmt.Errorf("get dashboard stats: %w", err)
	}

	return stats, nil
}

// GetPendingReportCount returns the number of pending reports.
func GetPendingReportCount(
	db *sql.DB,
) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM reports
		WHERE status = 'pending'
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"get pending report count: %w",
			err,
		)
	}

	return count, nil
}

// GetPendingModeratorRequestCount returns the number of pending moderator requests.
func GetPendingModeratorRequestCount(
	db *sql.DB,
) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM moderator_requests
		WHERE status = 'pending'
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"get pending moderator request count: %w",
			err,
		)
	}

	return count, nil
}