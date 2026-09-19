package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrCategoryNotFound = errors.New("category not found")

type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	Tagline     string
	CreatedAt   string
}

func CreateCategory(
	db *sql.DB,
	name string,
	slug string,
	description string,
	tagline string,
) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO categories (
			name,
			slug,
			description,
			tagline
		)
		VALUES (?, ?, ?, ?)
	`,
		name,
		slug,
		description,
		tagline,
	)
	if err != nil {
		return 0, fmt.Errorf("create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created category id: %w", err)
	}

	return id, nil
}

func GetCategoryByID(
	db *sql.DB,
	id int64,
) (Category, error) {
	var category Category

	err := db.QueryRow(`
		SELECT
			id,
			name,
			slug,
			description,
			tagline,
			created_at
		FROM categories
		WHERE id = ?
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Tagline,
		&category.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}

	if err != nil {
		return Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return category, nil
}

func GetCategoryBySlug(
	db *sql.DB,
	slug string,
) (Category, error) {
	var category Category

	err := db.QueryRow(`
		SELECT
			id,
			name,
			slug,
			description,
			tagline,
			created_at
		FROM categories
		WHERE slug = ?
	`, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Tagline,
		&category.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}

	if err != nil {
		return Category{}, fmt.Errorf("get category by slug: %w", err)
	}

	return category, nil
}

func GetAllCategories(
	db *sql.DB,
) ([]Category, error) {
	rows, err := db.Query(`
		SELECT
			id,
			name,
			slug,
			description,
			tagline,
			created_at
		FROM categories
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
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
			return nil, fmt.Errorf("scan category: %w", err)
		}

		categories = append(
			categories,
			category,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func UpdateCategory(
	db *sql.DB,
	id int64,
	name string,
	slug string,
	description string,
	tagline string,
) error {
	result, err := db.Exec(`
		UPDATE categories
		SET
			name = ?,
			slug = ?,
			description = ?,
			tagline = ?
		WHERE id = ?
	`,
		name,
		slug,
		description,
		tagline,
		id,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after updating category: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func DeleteCategory(
	db *sql.DB,
	id int64,
) error {
	result, err := db.Exec(`
		DELETE FROM categories
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows after deleting category: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
