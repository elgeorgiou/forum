package database

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "forum-test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	schemaPath := filepath.Join("schema", "schema.sql")

	if err := InitSchema(db, schemaPath); err != nil {
		t.Fatalf("initialize test schema: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
	t.Helper()

	id, err := CreateUser(
		db,
		username,
		email,
		"test-password-hash",
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return id
}

func createTestCategory(t *testing.T, db *sql.DB, name, slug string) int64 {
	t.Helper()

	id, err := CreateCategory(
		db,
		name,
		slug,
		"Test category",
		"Test category tagline",
	)
	if err != nil {
		t.Fatalf("create test category: %v", err)
	}

	return id
}

func TestOpenAndInitSchema(t *testing.T) {
	db := setupTestDB(t)

	var tableName string

	err := db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name = 'users'
	`).Scan(&tableName)
	if err != nil {
		t.Fatalf("find users table: %v", err)
	}

	if tableName != "users" {
		t.Fatalf("expected users table, got %q", tableName)
	}
}

func TestUsers(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"eleana",
		"eleana@example.com",
	)

	user, err := GetUserByID(db, userID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}

	if user.Username != "eleana" {
		t.Fatalf("expected username %q, got %q", "eleana", user.Username)
	}

	if user.Email != "eleana@example.com" {
		t.Fatalf(
			"expected email %q, got %q",
			"eleana@example.com",
			user.Email,
		)
	}

	userByEmail, err := GetUserByEmail(db, "eleana@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}

	if userByEmail.ID != userID {
		t.Fatalf("expected user id %d, got %d", userID, userByEmail.ID)
	}

	userByUsername, err := GetUserByUsername(db, "eleana")
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}

	if userByUsername.ID != userID {
		t.Fatalf("expected user id %d, got %d", userID, userByUsername.ID)
	}

	if err := UpdateUserRole(db, userID, "moderator"); err != nil {
		t.Fatalf("update user role: %v", err)
	}

	updatedUser, err := GetUserByID(db, userID)
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}

	if updatedUser.Role != "moderator" {
		t.Fatalf(
			"expected role %q, got %q",
			"moderator",
			updatedUser.Role,
		)
	}
}

func TestCategories(t *testing.T) {
	db := setupTestDB(t)

	categoryID := createTestCategory(
		t,
		db,
		"Horror",
		"horror",
	)

	category, err := GetCategoryByID(db, categoryID)
	if err != nil {
		t.Fatalf("get category by id: %v", err)
	}

	if category.Name != "Horror" {
		t.Fatalf("expected category %q, got %q", "Horror", category.Name)
	}

	bySlug, err := GetCategoryBySlug(db, "horror")
	if err != nil {
		t.Fatalf("get category by slug: %v", err)
	}

	if bySlug.ID != categoryID {
		t.Fatalf(
			"expected category id %d, got %d",
			categoryID,
			bySlug.ID,
		)
	}

	categories, err := GetAllCategories(db)
	if err != nil {
		t.Fatalf("get all categories: %v", err)
	}

	if len(categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(categories))
	}

	if err := DeleteCategory(db, categoryID); err != nil {
		t.Fatalf("delete category: %v", err)
	}

	_, err = GetCategoryByID(db, categoryID)
	if err != ErrCategoryNotFound {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestPosts(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"postuser",
		"postuser@example.com",
	)

	categoryID := createTestCategory(
		t,
		db,
		"RPG",
		"rpg",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Best RPG games",
		"What is your favorite RPG?",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := AddPostCategory(db, postID, categoryID); err != nil {
		t.Fatalf("add post category: %v", err)
	}

	post, err := GetPostByID(db, postID)
	if err != nil {
		t.Fatalf("get post by id: %v", err)
	}

	if post.Title != "Best RPG games" {
		t.Fatalf(
			"expected title %q, got %q",
			"Best RPG games",
			post.Title,
		)
	}

	postsByUser, err := GetPostsByUser(db, userID)
	if err != nil {
		t.Fatalf("get posts by user: %v", err)
	}

	if len(postsByUser) != 1 {
		t.Fatalf("expected 1 user post, got %d", len(postsByUser))
	}

	postsByCategory, err := GetPostsByCategory(db, categoryID)
	if err != nil {
		t.Fatalf("get posts by category: %v", err)
	}

	if len(postsByCategory) != 1 {
		t.Fatalf(
			"expected 1 category post, got %d",
			len(postsByCategory),
		)
	}

	err = UpdatePost(
		db,
		postID,
		"Updated RPG title",
		"Updated content",
		sql.NullString{
			String: "/static/uploads/test.png",
			Valid:  true,
		},
	)
	if err != nil {
		t.Fatalf("update post: %v", err)
	}

	updatedPost, err := GetPostByID(db, postID)
	if err != nil {
		t.Fatalf("get updated post: %v", err)
	}

	if updatedPost.Title != "Updated RPG title" {
		t.Fatalf(
			"expected updated title %q, got %q",
			"Updated RPG title",
			updatedPost.Title,
		)
	}

	if !updatedPost.ImagePath.Valid {
		t.Fatal("expected image path to be valid")
	}

	if err := DeletePost(db, postID); err != nil {
		t.Fatalf("delete post: %v", err)
	}

	_, err = GetPostByID(db, postID)
	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestComments(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"commentuser",
		"commentuser@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Test post",
		"Test content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	commentID, err := CreateComment(
		db,
		postID,
		userID,
		"First comment",
	)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	comment, err := GetCommentByID(db, commentID)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	if comment.Content != "First comment" {
		t.Fatalf(
			"expected comment %q, got %q",
			"First comment",
			comment.Content,
		)
	}

	comments, err := GetCommentsByPost(db, postID)
	if err != nil {
		t.Fatalf("get comments by post: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	userComments, err := GetCommentsByUser(db, userID)
	if err != nil {
		t.Fatalf("get comments by user: %v", err)
	}

	if len(userComments) != 1 {
		t.Fatalf(
			"expected 1 user comment, got %d",
			len(userComments),
		)
	}

	if err := UpdateComment(
		db,
		commentID,
		"Updated comment",
	); err != nil {
		t.Fatalf("update comment: %v", err)
	}

	updatedComment, err := GetCommentByID(db, commentID)
	if err != nil {
		t.Fatalf("get updated comment: %v", err)
	}

	if updatedComment.Content != "Updated comment" {
		t.Fatalf(
			"expected updated comment %q, got %q",
			"Updated comment",
			updatedComment.Content,
		)
	}

	if err := DeleteComment(db, commentID); err != nil {
		t.Fatalf("delete comment: %v", err)
	}

	_, err = GetCommentByID(db, commentID)
	if err != ErrCommentNotFound {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestPostReactions(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestUser(
		t,
		db,
		"author",
		"author@example.com",
	)

	reviewerID := createTestUser(
		t,
		db,
		"reviewer",
		"reviewer@example.com",
	)

	postID, err := CreatePost(
		db,
		authorID,
		"Reaction test",
		"Post content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := SetPostReaction(
		db,
		reviewerID,
		postID,
		ReactionLike,
	); err != nil {
		t.Fatalf("set post like: %v", err)
	}

	likes, dislikes, err := GetPostReactionCounts(db, postID)
	if err != nil {
		t.Fatalf("get post reaction counts: %v", err)
	}

	if likes != 1 || dislikes != 0 {
		t.Fatalf(
			"expected 1 like and 0 dislikes, got %d and %d",
			likes,
			dislikes,
		)
	}

	if err := SetPostReaction(
		db,
		reviewerID,
		postID,
		ReactionDislike,
	); err != nil {
		t.Fatalf("change post reaction: %v", err)
	}

	reaction, err := GetPostReactionByUser(
		db,
		reviewerID,
		postID,
	)
	if err != nil {
		t.Fatalf("get post reaction by user: %v", err)
	}

	if reaction != ReactionDislike {
		t.Fatalf(
			"expected reaction %d, got %d",
			ReactionDislike,
			reaction,
		)
	}

	likedPosts, err := GetLikedPostsByUser(db, reviewerID)
	if err != nil {
		t.Fatalf("get liked posts: %v", err)
	}

	if len(likedPosts) != 0 {
		t.Fatalf(
			"expected 0 liked posts, got %d",
			len(likedPosts),
		)
	}

	dislikedPosts, err := GetDislikedPostsByUser(db, reviewerID)
	if err != nil {
		t.Fatalf("get disliked posts: %v", err)
	}

	if len(dislikedPosts) != 1 {
		t.Fatalf(
			"expected 1 disliked post, got %d",
			len(dislikedPosts),
		)
	}

	if err := RemovePostReaction(
		db,
		reviewerID,
		postID,
	); err != nil {
		t.Fatalf("remove post reaction: %v", err)
	}
}

func TestCommentReactions(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"commentauthor",
		"commentauthor@example.com",
	)

	reactorID := createTestUser(
		t,
		db,
		"commentreactor",
		"commentreactor@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Comment reaction post",
		"Content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	commentID, err := CreateComment(
		db,
		postID,
		userID,
		"React to this comment",
	)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	if err := SetCommentReaction(
		db,
		reactorID,
		commentID,
		ReactionLike,
	); err != nil {
		t.Fatalf("set comment reaction: %v", err)
	}

	likes, dislikes, err := GetCommentReactionCounts(
		db,
		commentID,
	)
	if err != nil {
		t.Fatalf("get comment reaction counts: %v", err)
	}

	if likes != 1 || dislikes != 0 {
		t.Fatalf(
			"expected 1 like and 0 dislikes, got %d and %d",
			likes,
			dislikes,
		)
	}

	reaction, err := GetCommentReactionByUser(
		db,
		reactorID,
		commentID,
	)
	if err != nil {
		t.Fatalf("get comment reaction by user: %v", err)
	}

	if reaction != ReactionLike {
		t.Fatalf(
			"expected reaction %d, got %d",
			ReactionLike,
			reaction,
		)
	}

	if err := RemoveCommentReaction(
		db,
		reactorID,
		commentID,
	); err != nil {
		t.Fatalf("remove comment reaction: %v", err)
	}
}

func TestSessions(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"sessionuser",
		"sessionuser@example.com",
	)

	err := CreateSession(
		db,
		"session-one",
		userID,
		"2099-01-01 00:00:00",
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	session, err := GetSessionByID(db, "session-one")
	if err != nil {
		t.Fatalf("get session by id: %v", err)
	}

	if session.UserID != userID {
		t.Fatalf(
			"expected user id %d, got %d",
			userID,
			session.UserID,
		)
	}

	err = CreateSession(
		db,
		"session-two",
		userID,
		"2099-01-02 00:00:00",
	)
	if err != nil {
		t.Fatalf("replace session: %v", err)
	}

	_, err = GetSessionByID(db, "session-one")
	if err != ErrSessionNotFound {
		t.Fatalf(
			"expected old session to be removed, got %v",
			err,
		)
	}

	session, err = GetSessionByUserID(db, userID)
	if err != nil {
		t.Fatalf("get session by user id: %v", err)
	}

	if session.ID != "session-two" {
		t.Fatalf(
			"expected session id %q, got %q",
			"session-two",
			session.ID,
		)
	}

	if err := DeleteSessionByUserID(db, userID); err != nil {
		t.Fatalf("delete session by user id: %v", err)
	}
}

func TestNotifications(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"notificationuser",
		"notification@example.com",
	)

	actorID := createTestUser(
		t,
		db,
		"notificationactor",
		"actor@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Notification post",
		"Content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	notificationID, err := CreateNotification(
		db,
		userID,
		sql.NullInt64{
			Int64: actorID,
			Valid: true,
		},
		"post_like",
		sql.NullInt64{
			Int64: postID,
			Valid: true,
		},
		sql.NullInt64{},
	)
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	count, err := GetUnreadNotificationCount(db, userID)
	if err != nil {
		t.Fatalf("get unread notification count: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 unread notification, got %d", count)
	}

	notifications, err := GetNotificationsByUser(db, userID)
	if err != nil {
		t.Fatalf("get notifications by user: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf(
			"expected 1 notification, got %d",
			len(notifications),
		)
	}

	if err := MarkNotificationAsRead(
		db,
		notificationID,
		userID,
	); err != nil {
		t.Fatalf("mark notification as read: %v", err)
	}

	count, err = GetUnreadNotificationCount(db, userID)
	if err != nil {
		t.Fatalf("get unread notification count: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected 0 unread notifications, got %d", count)
	}

	if err := DeleteNotification(
		db,
		notificationID,
		userID,
	); err != nil {
		t.Fatalf("delete notification: %v", err)
	}
}

func TestModeratorRequests(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"moderatorcandidate",
		"candidate@example.com",
	)

	adminID := createTestUser(
		t,
		db,
		"admin",
		"admin@example.com",
	)

	requestID, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf("create moderator request: %v", err)
	}

	requests, err := GetPendingModeratorRequests(db)
	if err != nil {
		t.Fatalf("get pending moderator requests: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"expected 1 moderator request, got %d",
			len(requests),
		)
	}

	if err := ReviewModeratorRequest(
		db,
		requestID,
		"approved",
		adminID,
	); err != nil {
		t.Fatalf("review moderator request: %v", err)
	}

	requests, err = GetPendingModeratorRequests(db)
	if err != nil {
		t.Fatalf("get pending moderator requests: %v", err)
	}

	if len(requests) != 0 {
		t.Fatalf(
			"expected 0 pending moderator requests, got %d",
			len(requests),
		)
	}
}

func TestReports(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestUser(
		t,
		db,
		"reportauthor",
		"reportauthor@example.com",
	)

	reporterID := createTestUser(
		t,
		db,
		"reporter",
		"reporter@example.com",
	)

	adminID := createTestUser(
		t,
		db,
		"reportadmin",
		"reportadmin@example.com",
	)

	postID, err := CreatePost(
		db,
		authorID,
		"Report test",
		"Reported content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	reportID, err := CreatePostReport(
		db,
		reporterID,
		postID,
		"Test report reason",
	)
	if err != nil {
		t.Fatalf("create post report: %v", err)
	}

	reports, err := GetPendingReports(db)
	if err != nil {
		t.Fatalf("get pending reports: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf(
			"expected 1 pending report, got %d",
			len(reports),
		)
	}

	if err := ReviewReport(
		db,
		reportID,
		"resolved",
		adminID,
		"Report reviewed and resolved",
	); err != nil {
		t.Fatalf("review report: %v", err)
	}

	reports, err = GetPendingReports(db)
	if err != nil {
		t.Fatalf("get pending reports: %v", err)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected 0 pending reports, got %d",
			len(reports),
		)
	}

	notifications, err := GetNotificationViewsByUser(
		db,
		reporterID,
	)
	if err != nil {
		t.Fatalf(
			"get reporter notifications: %v",
			err,
		)
	}

	if len(notifications) != 1 {
		t.Fatalf(
			"expected 1 report notification, got %d",
			len(notifications),
		)
	}

	if notifications[0].Type != NotificationReport {
		t.Fatalf(
			"expected notification type %q, got %q",
			NotificationReport,
			notifications[0].Type,
		)
	}

	if notifications[0].ReportStatus != "resolved" {
		t.Fatalf(
			"expected report status %q, got %q",
			"resolved",
			notifications[0].ReportStatus,
		)
	}

	if notifications[0].ReportResponse !=
		"Report reviewed and resolved" {
		t.Fatalf(
			"expected report response %q, got %q",
			"Report reviewed and resolved",
			notifications[0].ReportResponse,
		)
	}
}

func TestOAuthAccounts(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"oauthuser",
		"oauth@example.com",
	)

	accountID, err := CreateOAuthAccount(
		db,
		userID,
		"github",
		"github-user-123",
	)
	if err != nil {
		t.Fatalf("create oauth account: %v", err)
	}

	if accountID == 0 {
		t.Fatal("expected oauth account id")
	}

	account, err := GetOAuthAccount(
		db,
		"github",
		"github-user-123",
	)
	if err != nil {
		t.Fatalf("get oauth account: %v", err)
	}

	if account.UserID != userID {
		t.Fatalf(
			"expected user id %d, got %d",
			userID,
			account.UserID,
		)
	}

	accounts, err := GetOAuthAccountsByUser(db, userID)
	if err != nil {
		t.Fatalf("get oauth accounts by user: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("expected 1 oauth account, got %d", len(accounts))
	}

	if err := DeleteOAuthAccount(
		db,
		userID,
		"github",
	); err != nil {
		t.Fatalf("delete oauth account: %v", err)
	}

	_, err = GetOAuthAccount(
		db,
		"github",
		"github-user-123",
	)
	if err != ErrOAuthAccountNotFound {
		t.Fatalf(
			"expected ErrOAuthAccountNotFound, got %v",
			err,
		)
	}
}
