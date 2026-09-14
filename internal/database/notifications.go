package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotificationNotFound = errors.New("notification not found")

type Notification struct {
	ID        int64
	UserID    int64
	ActorID   sql.NullInt64
	Type      string
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	IsRead    bool
	CreatedAt string
}

func CreateNotification(
	db *sql.DB,
	userID int64,
	actorID sql.NullInt64,
	notificationType string,
	postID sql.NullInt64,
	commentID sql.NullInt64,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO notifications (
			user_id,
			actor_id,
			type,
			post_id,
			comment_id
		)
		VALUES (?, ?, ?, ?, ?)
	`, userID, actorID, notificationType, postID, commentID)
	if err != nil {
		return 0, fmt.Errorf("create notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created notification id: %w", err)
	}

	return id, nil
}

func GetNotificationsByUser(db *sql.DB, userID int64) ([]Notification, error) {
	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			actor_id,
			type,
			post_id,
			comment_id,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get notifications by user: %w", err)
	}
	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var notification Notification

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ActorID,
			&notification.Type,
			&notification.PostID,
			&notification.CommentID,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

func GetUnreadNotificationCount(db *sql.DB, userID int64) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ?
		  AND is_read = 0
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread notification count: %w", err)
	}

	return count, nil
}

func MarkNotificationAsRead(db *sql.DB, notificationID, userID int64) error {
	result, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE id = ?
		  AND user_id = ?
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after marking notification as read: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func MarkAllNotificationsAsRead(db *sql.DB, userID int64) error {
	_, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ?
		  AND is_read = 0
	`, userID)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}

	return nil
}

func DeleteNotification(db *sql.DB, notificationID, userID int64) error {
	result, err := db.Exec(`
		DELETE FROM notifications
		WHERE id = ?
		  AND user_id = ?
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after deleting notification: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}
