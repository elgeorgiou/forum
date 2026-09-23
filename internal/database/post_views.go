package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// PostView represents a post together with its author, engagement,
// category, and display-related data.
type PostView struct {
	ID           int64
	UserID       int64
	Username     string
	Title        string
	Content      string
	Preview      string
	ImagePath    string
	CreatedAt    string
	UpdatedAt    string
	CommentCount int
	LikeCount    int
	DislikeCount int
	Score        int
	ViewCount    int
	Categories   []Category
	CategoryName string
	CategorySlug string
}

// GetPostViewByID retrieves a post with its author, reaction,
// comment, and category data by its database ID.
func GetPostViewByID(
	db *sql.DB,
	postID int64,
) (PostView, error) {
	var post PostView

	err := db.QueryRow(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Username,
		&post.Title,
		&post.Content,
		&post.ImagePath,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.CommentCount,
		&post.LikeCount,
		&post.DislikeCount,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return PostView{}, ErrPostNotFound
	}

	if err != nil {
		return PostView{}, fmt.Errorf(
			"get post view by id: %w",
			err,
		)
	}

	post.Preview = postPreview(post.Content)
	post.Score = post.LikeCount - post.DislikeCount

	categories, err := GetPostCategories(
		db,
		post.ID,
	)
	if err != nil {
		return PostView{}, err
	}

	post.Categories = categories

	if len(categories) > 0 {
		post.CategoryName = categories[0].Name
		post.CategorySlug = categories[0].Slug
	}

	return post, nil
}

// GetAllPostViews retrieves all posts with their associated display data.
func GetAllPostViews(
	db *sql.DB,
) ([]PostView, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf(
			"get all post views: %w",
			err,
		)
	}
	defer rows.Close()

	return scanPostViews(db, rows)
}

// GetPostViewsByCategory retrieves all posts belonging to a category.
func GetPostViewsByCategory(
	db *sql.DB,
	categoryID int64,
) ([]PostView, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		INNER JOIN post_categories AS pc
			ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf(
			"get post views by category: %w",
			err,
		)
	}
	defer rows.Close()

	return scanPostViews(db, rows)
}

// GetPostViewsByUser retrieves all posts created by a user.
func GetPostViewsByUser(
	db *sql.DB,
	userID int64,
) ([]PostView, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS pr
				WHERE pr.post_id = p.id
				  AND pr.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get post views by user: %w",
			err,
		)
	}
	defer rows.Close()

	return scanPostViews(db, rows)
}

// GetLikedPostViewsByUser retrieves all posts liked by a user.
func GetLikedPostViewsByUser(
	db *sql.DB,
	userID int64,
) ([]PostView, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS counts
				WHERE counts.post_id = p.id
				  AND counts.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS counts
				WHERE counts.post_id = p.id
				  AND counts.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		INNER JOIN post_reactions AS pr
			ON pr.post_id = p.id
		WHERE pr.user_id = ?
		  AND pr.reaction = 1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get liked post views by user: %w",
			err,
		)
	}
	defer rows.Close()

	return scanPostViews(db, rows)
}

// GetDislikedPostViewsByUser retrieves all posts disliked by a user.
func GetDislikedPostViewsByUser(
	db *sql.DB,
	userID int64,
) ([]PostView, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			p.user_id,
			u.username,
			p.title,
			p.content,
			COALESCE(p.image_path, ''),
			p.created_at,
			p.updated_at,
			(
				SELECT COUNT(*)
				FROM comments AS c
				WHERE c.post_id = p.id
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS counts
				WHERE counts.post_id = p.id
				  AND counts.reaction = 1
			),
			(
				SELECT COUNT(*)
				FROM post_reactions AS counts
				WHERE counts.post_id = p.id
				  AND counts.reaction = -1
			)
		FROM posts AS p
		INNER JOIN users AS u
			ON u.id = p.user_id
		INNER JOIN post_reactions AS pr
			ON pr.post_id = p.id
		WHERE pr.user_id = ?
		  AND pr.reaction = -1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"get disliked post views by user: %w",
			err,
		)
	}
	defer rows.Close()

	return scanPostViews(db, rows)
}

// GetPostCategories retrieves all categories associated with a post.
func GetPostCategories(
	db *sql.DB,
	postID int64,
) ([]Category, error) {
	rows, err := db.Query(`
		SELECT
			c.id,
			c.name,
			c.slug,
			c.description,
			c.tagline,
			c.created_at
		FROM categories AS c
		INNER JOIN post_categories AS pc
			ON pc.category_id = c.id
		WHERE pc.post_id = ?
		ORDER BY c.name ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf(
			"get post categories: %w",
			err,
		)
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.Tagline,
			&category.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan post category: %w",
				err,
			)
		}

		categories = append(
			categories,
			category,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate post categories: %w",
			err,
		)
	}

	return categories, nil
}

// scanPostViews scans post rows and enriches each post with its categories.
func scanPostViews(
	db *sql.DB,
	rows *sql.Rows,
) ([]PostView, error) {
	var posts []PostView

	for rows.Next() {
		var post PostView

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.CommentCount,
			&post.LikeCount,
			&post.DislikeCount,
		); err != nil {
			return nil, fmt.Errorf(
				"scan post view: %w",
				err,
			)
		}

		post.Preview = postPreview(
			post.Content,
		)

		post.Score =
			post.LikeCount -
				post.DislikeCount

		posts = append(
			posts,
			post,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate post views: %w",
			err,
		)
	}

	for index := range posts {
		categories, err := GetPostCategories(
			db,
			posts[index].ID,
		)
		if err != nil {
			return nil, err
		}

		posts[index].Categories = categories

		if len(categories) > 0 {
			posts[index].CategoryName =
				categories[0].Name

			posts[index].CategorySlug =
				categories[0].Slug
		}
	}

	return posts, nil
}

// postPreview creates a shortened preview of post content.
func postPreview(
	content string,
) string {
	content = strings.TrimSpace(content)

	const maximumLength = 180

	runes := []rune(content)

	if len(runes) <= maximumLength {
		return content
	}

	return strings.TrimSpace(
		string(runes[:maximumLength]),
	) + "..."
}