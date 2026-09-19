package database

import (
	"database/sql"
	"errors"
	"testing"
)

func TestUpdateCategory(t *testing.T) {
	db := setupTestDB(t)

	categoryID := createTestCategory(
		t,
		db,
		"Action",
		"action",
	)

	err := UpdateCategory(
		db,
		categoryID,
		"Action Games",
		"action-games",
		"Fast-paced action games.",
		"Fight. Survive. Win.",
	)
	if err != nil {
		t.Fatalf("update category: %v", err)
	}

	category, err := GetCategoryByID(
		db,
		categoryID,
	)
	if err != nil {
		t.Fatalf("get updated category: %v", err)
	}

	if category.Name != "Action Games" {
		t.Fatalf(
			"expected name %q, got %q",
			"Action Games",
			category.Name,
		)
	}

	if category.Slug != "action-games" {
		t.Fatalf(
			"expected slug %q, got %q",
			"action-games",
			category.Slug,
		)
	}

	if category.Description != "Fast-paced action games." {
		t.Fatalf(
			"unexpected description: %q",
			category.Description,
		)
	}

	if category.Tagline != "Fight. Survive. Win." {
		t.Fatalf(
			"unexpected tagline: %q",
			category.Tagline,
		)
	}
}

func TestUpdateCategoryNotFound(t *testing.T) {
	db := setupTestDB(t)

	err := UpdateCategory(
		db,
		999999,
		"Missing",
		"missing",
		"Missing category.",
		"Missing.",
	)

	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf(
			"expected ErrCategoryNotFound, got %v",
			err,
		)
	}
}

func TestDeleteCategoryRemovesPostCategoryRelation(
	t *testing.T,
) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"category_admin_test",
		"category_admin_test@example.com",
	)

	categoryID := createTestCategory(
		t,
		db,
		"Strategy",
		"strategy",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Strategy games",
		"Strategy discussion",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := AddPostCategory(
		db,
		postID,
		categoryID,
	); err != nil {
		t.Fatalf("add post category: %v", err)
	}

	if err := DeleteCategory(
		db,
		categoryID,
	); err != nil {
		t.Fatalf("delete category: %v", err)
	}

	var count int

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM post_categories
		WHERE post_id = ?
		  AND category_id = ?
	`,
		postID,
		categoryID,
	).Scan(&count)
	if err != nil {
		t.Fatalf(
			"count post category relation: %v",
			err,
		)
	}

	if count != 0 {
		t.Fatalf(
			"expected relation to be deleted, got %d",
			count,
		)
	}
}
