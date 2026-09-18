package database

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestPostPreviewShortContent(t *testing.T) {
	content := "This is a short post."

	preview := postPreview(content)

	if preview != content {
		t.Fatalf(
			"expected %q, got %q",
			content,
			preview,
		)
	}
}

func TestPostPreviewTrimsWhitespace(t *testing.T) {
	preview := postPreview(
		"   Hello players!   ",
	)

	if preview != "Hello players!" {
		t.Fatalf(
			"expected trimmed preview, got %q",
			preview,
		)
	}
}

func TestPostPreviewLongContent(t *testing.T) {
	content := strings.Repeat(
		"a",
		200,
	)

	preview := postPreview(content)

	if !strings.HasSuffix(
		preview,
		"...",
	) {
		t.Fatalf(
			"expected preview to end with ellipsis, got %q",
			preview,
		)
	}

	runes := []rune(
		strings.TrimSuffix(
			preview,
			"...",
		),
	)

	if len(runes) != 180 {
		t.Fatalf(
			"expected 180 characters before ellipsis, got %d",
			len(runes),
		)
	}
}

func TestPostPreviewUnicode(t *testing.T) {
	content := strings.Repeat(
		"🎮",
		200,
	)

	preview := postPreview(content)

	withoutEllipsis := strings.TrimSuffix(
		preview,
		"...",
	)

	if len([]rune(withoutEllipsis)) != 180 {
		t.Fatalf(
			"expected 180 unicode characters, got %d",
			len([]rune(withoutEllipsis)),
		)
	}
}

func TestGetPostViewByID(t *testing.T) {
	db := openPostViewTestDatabase(t)
	defer db.Close()

	userID := createPostViewTestUser(
		t,
		db,
		"player-one",
		"player@example.com",
	)

	categoryID := createPostViewTestCategory(
		t,
		db,
		"RPG",
		"rpg",
	)

	postID, err := CreatePost(
		db,
		userID,
		"My RPG Post",
		"This is the full post content.",
		sql.NullString{
			String: "/static/uploads/test.png",
			Valid:  true,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := AddPostCategory(
		db,
		postID,
		categoryID,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := CreateComment(
		db,
		postID,
		userID,
		"First comment",
	); err != nil {
		t.Fatal(err)
	}

	if err := SetPostReaction(
		db,
		userID,
		postID,
		1,
	); err != nil {
		t.Fatal(err)
	}

	post, err := GetPostViewByID(
		db,
		postID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if post.ID != postID {
		t.Fatalf(
			"expected post ID %d, got %d",
			postID,
			post.ID,
		)
	}

	if post.UserID != userID {
		t.Fatalf(
			"expected user ID %d, got %d",
			userID,
			post.UserID,
		)
	}

	if post.Username != "player-one" {
		t.Fatalf(
			"expected username player-one, got %q",
			post.Username,
		)
	}

	if post.Title != "My RPG Post" {
		t.Fatalf(
			"expected title My RPG Post, got %q",
			post.Title,
		)
	}

	if post.Content != "This is the full post content." {
		t.Fatalf(
			"unexpected content %q",
			post.Content,
		)
	}

	if post.ImagePath != "/static/uploads/test.png" {
		t.Fatalf(
			"unexpected image path %q",
			post.ImagePath,
		)
	}

	if post.CommentCount != 1 {
		t.Fatalf(
			"expected 1 comment, got %d",
			post.CommentCount,
		)
	}

	if post.LikeCount != 1 {
		t.Fatalf(
			"expected 1 like, got %d",
			post.LikeCount,
		)
	}

	if post.DislikeCount != 0 {
		t.Fatalf(
			"expected 0 dislikes, got %d",
			post.DislikeCount,
		)
	}

	if post.Score != 1 {
		t.Fatalf(
			"expected score 1, got %d",
			post.Score,
		)
	}

	if len(post.Categories) != 1 {
		t.Fatalf(
			"expected 1 category, got %d",
			len(post.Categories),
		)
	}

	if post.CategoryName != "RPG" {
		t.Fatalf(
			"expected category RPG, got %q",
			post.CategoryName,
		)
	}

	if post.CategorySlug != "rpg" {
		t.Fatalf(
			"expected category slug rpg, got %q",
			post.CategorySlug,
		)
	}
}

func TestGetPostViewByIDNotFound(t *testing.T) {
	db := openPostViewTestDatabase(t)
	defer db.Close()

	_, err := GetPostViewByID(
		db,
		999999,
	)

	if !errors.Is(
		err,
		ErrPostNotFound,
	) {
		t.Fatalf(
			"expected ErrPostNotFound, got %v",
			err,
		)
	}
}

func openPostViewTestDatabase(
	t *testing.T,
) *sql.DB {
	t.Helper()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			tagline TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_path TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE post_categories (
			post_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, category_id)
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

	CREATE TABLE post_reactions (
	user_id INTEGER NOT NULL,
	post_id INTEGER NOT NULL,
	reaction INTEGER NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, post_id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatal(err)
	}

	return db
}

func createPostViewTestUser(
	t *testing.T,
	db *sql.DB,
	username string,
	email string,
) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO users (
			username,
			email,
			password_hash
		)
		VALUES (?, ?, ?)
	`, username, email, "test-hash")
	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func createPostViewTestCategory(
	t *testing.T,
	db *sql.DB,
	name string,
	slug string,
) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO categories (
			name,
			slug,
			description,
			tagline
		)
		VALUES (?, ?, ?, ?)
	`, name, slug, "", "")
	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	return id
}
