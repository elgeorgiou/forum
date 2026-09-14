package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrReactionNotFound = errors.New("reaction not found")

const (
	ReactionDislike = -1
	ReactionLike    = 1
)

func SetPostReaction(db *sql.DB, userID, postID int64, reaction int) error {
	if reaction != ReactionLike && reaction != ReactionDislike {
		return errors.New("invalid post reaction")
	}

	_, err := db.Exec(`
		INSERT INTO post_reactions (user_id, post_id, reaction)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, post_id)
		DO UPDATE SET
			reaction = excluded.reaction,
			created_at = CURRENT_TIMESTAMP
	`, userID, postID, reaction)
	if err != nil {
		return fmt.Errorf("set post reaction: %w", err)
	}

	return nil
}

func RemovePostReaction(db *sql.DB, userID, postID int64) error {
	result, err := db.Exec(`
		DELETE FROM post_reactions
		WHERE user_id = ? AND post_id = ?
	`, userID, postID)
	if err != nil {
		return fmt.Errorf("remove post reaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after removing post reaction: %w", err)
	}

	if rowsAffected == 0 {
		return ErrReactionNotFound
	}

	return nil
}

func GetPostReactionCounts(db *sql.DB, postID int64) (likes, dislikes int, err error) {
	err = db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN reaction = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN reaction = -1 THEN 1 ELSE 0 END), 0)
		FROM post_reactions
		WHERE post_id = ?
	`, postID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("get post reaction counts: %w", err)
	}

	return likes, dislikes, nil
}

func GetPostReactionByUser(db *sql.DB, userID, postID int64) (int, error) {
	var reaction int

	err := db.QueryRow(`
		SELECT reaction
		FROM post_reactions
		WHERE user_id = ? AND post_id = ?
	`, userID, postID).Scan(&reaction)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrReactionNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("get post reaction by user: %w", err)
	}

	return reaction, nil
}

func GetLikedPostsByUser(db *sql.DB, userID int64) ([]Post, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.image_path,
		       p.created_at, p.updated_at
		FROM posts AS p
		INNER JOIN post_reactions AS pr
			ON pr.post_id = p.id
		WHERE pr.user_id = ?
		  AND pr.reaction = ?
		ORDER BY pr.created_at DESC
	`, userID, ReactionLike)
	if err != nil {
		return nil, fmt.Errorf("get liked posts by user: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func GetDislikedPostsByUser(db *sql.DB, userID int64) ([]Post, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.image_path,
		       p.created_at, p.updated_at
		FROM posts AS p
		INNER JOIN post_reactions AS pr
			ON pr.post_id = p.id
		WHERE pr.user_id = ?
		  AND pr.reaction = ?
		ORDER BY pr.created_at DESC
	`, userID, ReactionDislike)
	if err != nil {
		return nil, fmt.Errorf("get disliked posts by user: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func SetCommentReaction(db *sql.DB, userID, commentID int64, reaction int) error {
	if reaction != ReactionLike && reaction != ReactionDislike {
		return errors.New("invalid comment reaction")
	}

	_, err := db.Exec(`
		INSERT INTO comment_reactions (user_id, comment_id, reaction)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, comment_id)
		DO UPDATE SET
			reaction = excluded.reaction,
			created_at = CURRENT_TIMESTAMP
	`, userID, commentID, reaction)
	if err != nil {
		return fmt.Errorf("set comment reaction: %w", err)
	}

	return nil
}

func RemoveCommentReaction(db *sql.DB, userID, commentID int64) error {
	result, err := db.Exec(`
		DELETE FROM comment_reactions
		WHERE user_id = ? AND comment_id = ?
	`, userID, commentID)
	if err != nil {
		return fmt.Errorf("remove comment reaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after removing comment reaction: %w", err)
	}

	if rowsAffected == 0 {
		return ErrReactionNotFound
	}

	return nil
}

func GetCommentReactionCounts(db *sql.DB, commentID int64) (likes, dislikes int, err error) {
	err = db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN reaction = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN reaction = -1 THEN 1 ELSE 0 END), 0)
		FROM comment_reactions
		WHERE comment_id = ?
	`, commentID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, fmt.Errorf("get comment reaction counts: %w", err)
	}

	return likes, dislikes, nil
}

func GetCommentReactionByUser(db *sql.DB, userID, commentID int64) (int, error) {
	var reaction int

	err := db.QueryRow(`
		SELECT reaction
		FROM comment_reactions
		WHERE user_id = ? AND comment_id = ?
	`, userID, commentID).Scan(&reaction)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrReactionNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("get comment reaction by user: %w", err)
	}

	return reaction, nil
}
