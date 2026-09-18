package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrCommentNotFound = errors.New("comment not found")

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Content   string
	CreatedAt string
	UpdatedAt string
}

type CommentView struct {
	ID           int64
	PostID       int64
	UserID       int64
	Username     string
	Content      string
	CreatedAt    string
	UpdatedAt    string
	LikeCount    int
	DislikeCount int
	Score        int
}

func CreateComment(
	db *sql.DB,
	postID int64,
	userID int64,
	content string,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)
	`, postID, userID, content)
	if err != nil {
		return 0, fmt.Errorf("create comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created comment id: %w", err)
	}

	return id, nil
}

func GetCommentByID(
	db *sql.DB,
	commentID int64,
) (Comment, error) {
	var comment Comment

	err := db.QueryRow(`
		SELECT
			id,
			post_id,
			user_id,
			content,
			created_at,
			updated_at
		FROM comments
		WHERE id = ?
	`, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}

	if err != nil {
		return Comment{}, fmt.Errorf(
			"get comment by id: %w",
			err,
		)
	}

	return comment, nil
}

func GetCommentsByPost(
	db *sql.DB,
	postID int64,
) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT
			id,
			post_id,
			user_id,
			content,
			created_at,
			updated_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf(
			"get comments by post: %w",
			err,
		)
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan comment: %w",
				err,
			)
		}

		comments = append(
			comments,
			comment,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate comments: %w",
			err,
		)
	}

	return comments, nil
}

func GetCommentViewsByPost(
	db *sql.DB,
	postID int64,
) ([]CommentView, error) {
	rows, err := db.Query(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			u.username,
			c.content,
			c.created_at,
			c.updated_at,
			(
				SELECT COUNT(*)
				FROM comment_reactions AS cr
				WHERE cr.comment_id = c.id
				  AND cr.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM comment_reactions AS cr
				WHERE cr.comment_id = c.id
				  AND cr.reaction = -1
			)
		FROM comments AS c
		INNER JOIN users AS u
			ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf(
			"get comment views by post: %w",
			err,
		)
	}
	defer rows.Close()

	var comments []CommentView

	for rows.Next() {
		var comment CommentView

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.LikeCount,
			&comment.DislikeCount,
		); err != nil {
			return nil, fmt.Errorf(
				"scan comment view: %w",
				err,
			)
		}

		comment.Score =
			comment.LikeCount -
				comment.DislikeCount

		comments = append(
			comments,
			comment,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate comment views: %w",
			err,
		)
	}

	return comments, nil
}

func GetCommentsByUser(
	db *sql.DB,
	userID int64,
) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT
			id,
			post_id,
			user_id,
			content,
			created_at,
			updated_at
		FROM comments
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get comments by user: %w",
			err,
		)
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var comment Comment

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan comment: %w",
				err,
			)
		}

		comments = append(
			comments,
			comment,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate comments: %w",
			err,
		)
	}

	return comments, nil
}

func UpdateComment(
	db *sql.DB,
	commentID int64,
	content string,
) error {
	result, err := db.Exec(`
		UPDATE comments
		SET content = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, content, commentID)
	if err != nil {
		return fmt.Errorf(
			"update comment: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after updating comment: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func DeleteComment(
	db *sql.DB,
	commentID int64,
) error {
	result, err := db.Exec(`
		DELETE FROM comments
		WHERE id = ?
	`, commentID)
	if err != nil {
		return fmt.Errorf(
			"delete comment: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after deleting comment: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}
