package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrPostNotFound = errors.New("post not found")

type Post struct {
	ID        int64
	UserID    int64
	Title     string
	Content   string
	ImagePath sql.NullString
	CreatedAt string
	UpdatedAt string
}

func CreatePost(
	db *sql.DB,
	userID int64,
	title string,
	content string,
	imagePath sql.NullString,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO posts (user_id, title, content, image_path)
		VALUES (?, ?, ?, ?)
	`, userID, title, content, imagePath)
	if err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created post id: %w", err)
	}

	return id, nil
}

func GetPostByID(db *sql.DB, id int64) (Post, error) {
	var post Post

	err := db.QueryRow(`
		SELECT id, user_id, title, content, image_path, created_at, updated_at
		FROM posts
		WHERE id = ?
	`, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.ImagePath,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrPostNotFound
	}

	if err != nil {
		return Post{}, fmt.Errorf("get post by id: %w", err)
	}

	return post, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`
		SELECT id, user_id, title, content, image_path, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("get all posts: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func GetPostsByUser(db *sql.DB, userID int64) ([]Post, error) {
	rows, err := db.Query(`
		SELECT id, user_id, title, content, image_path, created_at, updated_at
		FROM posts
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get posts by user: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func GetPostsByCategory(db *sql.DB, categoryID int64) ([]Post, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.image_path,
		       p.created_at, p.updated_at
		FROM posts AS p
		INNER JOIN post_categories AS pc
			ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get posts by category: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func AddPostCategory(db *sql.DB, postID, categoryID int64) error {
	_, err := db.Exec(`
		INSERT INTO post_categories (post_id, category_id)
		VALUES (?, ?)
	`, postID, categoryID)
	if err != nil {
		return fmt.Errorf("add category to post: %w", err)
	}

	return nil
}

func UpdatePost(
	db *sql.DB,
	postID int64,
	title string,
	content string,
	imagePath sql.NullString,
) error {
	result, err := db.Exec(`
		UPDATE posts
		SET title = ?,
		    content = ?,
		    image_path = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, content, imagePath, postID)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after updating post: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	return nil
}

func DeletePost(db *sql.DB, postID int64) error {
	result, err := db.Exec(`
		DELETE FROM posts
		WHERE id = ?
	`, postID)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows after deleting post: %w", err)
	}

	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	return nil
}

func scanPosts(rows *sql.Rows) ([]Post, error) {
	var posts []Post

	for rows.Next() {
		var post Post

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	return posts, nil
}
