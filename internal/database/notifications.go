package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotificationNotFound is returned when a notification cannot be found.
var ErrNotificationNotFound = errors.New(
	"notification not found",
)

// Notification types identify the action that triggered a notification.
const (
	NotificationPostLike       = "post_like"
	NotificationPostDislike    = "post_dislike"
	NotificationComment        = "comment"
	NotificationCommentLike    = "comment_like"
	NotificationCommentDislike = "comment_dislike"
	NotificationReport         = "report"
)

// Notification represents a notification stored in the database.
type Notification struct {
	ID        int64
	UserID    int64
	ActorID   sql.NullInt64
	Type      string
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	ReportID  sql.NullInt64
	IsRead    bool
	CreatedAt string
}

// NotificationView represents a notification together with related display data.
type NotificationView struct {
	ID             int64
	UserID         int64
	ActorID        sql.NullInt64
	ActorUsername  string
	Type           string
	PostID         sql.NullInt64
	PostTitle      string
	CommentID      sql.NullInt64
	ReportID       sql.NullInt64
	ReportResponse string
	ReportStatus   string
	IsRead         bool
	CreatedAt      string
}

// CreateNotification creates a notification and returns its database ID.
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
	`,
		userID,
		actorID,
		notificationType,
		postID,
		commentID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"create notification: %w",
			err,
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(
			"get created notification id: %w",
			err,
		)
	}

	return id, nil
}

// CreatePostNotification creates a notification for an action on a post.
func CreatePostNotification(
	db *sql.DB,
	postID int64,
	actorID int64,
	notificationType string,
) error {
	post, err := GetPostByID(
		db,
		postID,
	)
	if err != nil {
		return err
	}

	if post.UserID == actorID {
		return nil
	}

	_, err = CreateNotification(
		db,
		post.UserID,
		sql.NullInt64{
			Int64: actorID,
			Valid: true,
		},
		notificationType,
		sql.NullInt64{
			Int64: postID,
			Valid: true,
		},
		sql.NullInt64{},
	)
	if err != nil {
		return fmt.Errorf(
			"create post notification: %w",
			err,
		)
	}

	return nil
}

// CreateCommentNotification creates a notification when a comment is added to a post.
func CreateCommentNotification(
	db *sql.DB,
	postID int64,
	commentID int64,
	actorID int64,
) error {
	post, err := GetPostByID(
		db,
		postID,
	)
	if err != nil {
		return err
	}

	if post.UserID == actorID {
		return nil
	}

	_, err = CreateNotification(
		db,
		post.UserID,
		sql.NullInt64{
			Int64: actorID,
			Valid: true,
		},
		NotificationComment,
		sql.NullInt64{
			Int64: postID,
			Valid: true,
		},
		sql.NullInt64{
			Int64: commentID,
			Valid: true,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create comment notification: %w",
			err,
		)
	}

	return nil
}

// GetNotificationsByUser retrieves all notifications belonging to a user.
func GetNotificationsByUser(
	db *sql.DB,
	userID int64,
) ([]Notification, error) {
	rows, err := db.Query(`
		SELECT
			id,
			user_id,
			actor_id,
			type,
			post_id,
			comment_id,
			report_id,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get notifications by user: %w",
			err,
		)
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
			&notification.ReportID,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan notification: %w",
				err,
			)
		}

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate notifications: %w",
			err,
		)
	}

	return notifications, nil
}

// GetNotificationViewsByUser retrieves notifications with related display data for a user.
func GetNotificationViewsByUser(
	db *sql.DB,
	userID int64,
) ([]NotificationView, error) {
	rows, err := db.Query(`
		SELECT
			n.id,
			n.user_id,
			n.actor_id,
			COALESCE(u.username, ''),
			n.type,
			n.post_id,
			COALESCE(p.title, cp.title, ''),
			n.comment_id,
			n.report_id,
			COALESCE(r.response, ''),
			COALESCE(r.status, ''),
			n.is_read,
			n.created_at
		FROM notifications AS n

		LEFT JOIN users AS u
			ON u.id = n.actor_id

		LEFT JOIN posts AS p
			ON p.id = n.post_id

		LEFT JOIN comments AS c
			ON c.id = n.comment_id

		LEFT JOIN posts AS cp
			ON cp.id = c.post_id

		LEFT JOIN reports AS r
			ON r.id = n.report_id

		WHERE n.user_id = ?

		ORDER BY n.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get notification views by user: %w",
			err,
		)
	}
	defer rows.Close()

	var notifications []NotificationView

	for rows.Next() {
		var notification NotificationView

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ActorID,
			&notification.ActorUsername,
			&notification.Type,
			&notification.PostID,
			&notification.PostTitle,
			&notification.CommentID,
			&notification.ReportID,
			&notification.ReportResponse,
			&notification.ReportStatus,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan notification view: %w",
				err,
			)
		}

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate notification views: %w",
			err,
		)
	}

	return notifications, nil
}

// GetUnreadNotificationCount returns the number of unread notifications for a user.
func GetUnreadNotificationCount(
	db *sql.DB,
	userID int64,
) (int, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ?
		  AND is_read = 0
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"get unread notification count: %w",
			err,
		)
	}

	return count, nil
}

// MarkNotificationAsRead marks a specific notification belonging to a user as read.
func MarkNotificationAsRead(
	db *sql.DB,
	notificationID int64,
	userID int64,
) error {
	result, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE id = ?
		  AND user_id = ?
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf(
			"mark notification as read: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after marking notification as read: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// MarkAllNotificationsAsRead marks all unread notifications belonging to a user as read.
func MarkAllNotificationsAsRead(
	db *sql.DB,
	userID int64,
) error {
	_, err := db.Exec(`
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ?
		  AND is_read = 0
	`, userID)
	if err != nil {
		return fmt.Errorf(
			"mark all notifications as read: %w",
			err,
		)
	}

	return nil
}

// DeleteNotification deletes a specific notification belonging to a user.
func DeleteNotification(
	db *sql.DB,
	notificationID int64,
	userID int64,
) error {
	result, err := db.Exec(`
		DELETE FROM notifications
		WHERE id = ?
		  AND user_id = ?
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf(
			"delete notification: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after deleting notification: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}