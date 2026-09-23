package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchPostViews searches posts by title, content, or associated category data.
func SearchPostViews(
	db *sql.DB,
	query string,
) ([]PostView, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, nil
	}

	search := "%" + query + "%"

	rows, err := db.Query(`
		SELECT DISTINCT
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
		LEFT JOIN post_categories AS pc
			ON pc.post_id = p.id
		LEFT JOIN categories AS c
			ON c.id = pc.category_id
		WHERE
			p.title LIKE ?
			OR p.content LIKE ?
			OR c.name LIKE ?
			OR c.slug LIKE ?
			OR c.description LIKE ?
			OR c.tagline LIKE ?
		ORDER BY p.created_at DESC
	`,
		search,
		search,
		search,
		search,
		search,
		search,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search posts: %w",
			err,
		)
	}
	defer rows.Close()

	posts, err := scanPostViews(
		db,
		rows,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"scan search posts: %w",
			err,
		)
	}

	return posts, nil
}

// SearchCategories searches categories by name, slug, description, or tagline.
func SearchCategories(
	db *sql.DB,
	query string,
) ([]Category, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, nil
	}

	search := "%" + query + "%"

	rows, err := db.Query(`
		SELECT
			id,
			name,
			slug,
			description,
			tagline,
			created_at
		FROM categories
		WHERE
			name LIKE ?
			OR slug LIKE ?
			OR description LIKE ?
			OR tagline LIKE ?
		ORDER BY name ASC
	`,
		search,
		search,
		search,
		search,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search categories: %w",
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
				"scan search category: %w",
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
			"iterate search categories: %w",
			err,
		)
	}

	return categories, nil
}