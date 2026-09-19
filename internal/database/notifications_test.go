package database

import (
	"database/sql"
	"testing"
)

func TestCreatePostNotification(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestUser(
		t,
		db,
		"notificationauthor",
		"notificationauthor@example.com",
	)

	actorID := createTestUser(
		t,
		db,
		"notificationreactor",
		"notificationreactor@example.com",
	)

	postID, err := CreatePost(
		db,
		authorID,
		"Notification test post",
		"Test content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	err = CreatePostNotification(
		db,
		postID,
		actorID,
		NotificationPostLike,
	)
	if err != nil {
		t.Fatalf("create post notification: %v", err)
	}

	notifications, err := GetNotificationViewsByUser(
		db,
		authorID,
	)
	if err != nil {
		t.Fatalf("get notification views: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf(
			"expected 1 notification, got %d",
			len(notifications),
		)
	}

	notification := notifications[0]

	if notification.UserID != authorID {
		t.Fatalf(
			"expected notification user %d, got %d",
			authorID,
			notification.UserID,
		)
	}

	if !notification.ActorID.Valid ||
		notification.ActorID.Int64 != actorID {
		t.Fatalf(
			"expected actor %d, got %v",
			actorID,
			notification.ActorID,
		)
	}

	if notification.ActorUsername != "notificationreactor" {
		t.Fatalf(
			"expected actor username %q, got %q",
			"notificationreactor",
			notification.ActorUsername,
		)
	}

	if notification.Type != NotificationPostLike {
		t.Fatalf(
			"expected notification type %q, got %q",
			NotificationPostLike,
			notification.Type,
		)
	}

	if !notification.PostID.Valid ||
		notification.PostID.Int64 != postID {
		t.Fatalf(
			"expected post %d, got %v",
			postID,
			notification.PostID,
		)
	}

	if notification.PostTitle != "Notification test post" {
		t.Fatalf(
			"expected post title %q, got %q",
			"Notification test post",
			notification.PostTitle,
		)
	}

	if notification.IsRead {
		t.Fatal("expected notification to be unread")
	}
}

func TestCreateCommentNotification(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestUser(
		t,
		db,
		"commentnotificationauthor",
		"commentnotificationauthor@example.com",
	)

	actorID := createTestUser(
		t,
		db,
		"commentnotificationactor",
		"commentnotificationactor@example.com",
	)

	postID, err := CreatePost(
		db,
		authorID,
		"Comment notification post",
		"Test content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	commentID, err := CreateComment(
		db,
		postID,
		actorID,
		"Notification comment",
	)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	err = CreateCommentNotification(
		db,
		postID,
		commentID,
		actorID,
	)
	if err != nil {
		t.Fatalf(
			"create comment notification: %v",
			err,
		)
	}

	notifications, err := GetNotificationViewsByUser(
		db,
		authorID,
	)
	if err != nil {
		t.Fatalf("get notification views: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf(
			"expected 1 notification, got %d",
			len(notifications),
		)
	}

	notification := notifications[0]

	if notification.Type != NotificationComment {
		t.Fatalf(
			"expected type %q, got %q",
			NotificationComment,
			notification.Type,
		)
	}

	if !notification.CommentID.Valid ||
		notification.CommentID.Int64 != commentID {
		t.Fatalf(
			"expected comment %d, got %v",
			commentID,
			notification.CommentID,
		)
	}

	if notification.ActorUsername !=
		"commentnotificationactor" {
		t.Fatalf(
			"expected actor username %q, got %q",
			"commentnotificationactor",
			notification.ActorUsername,
		)
	}
}

func TestOwnPostDoesNotCreateNotification(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"ownpostuser",
		"ownpostuser@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Own post",
		"Own content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	err = CreatePostNotification(
		db,
		postID,
		userID,
		NotificationPostLike,
	)
	if err != nil {
		t.Fatalf(
			"create own post notification: %v",
			err,
		)
	}

	notifications, err := GetNotificationsByUser(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get notifications: %v", err)
	}

	if len(notifications) != 0 {
		t.Fatalf(
			"expected 0 notifications, got %d",
			len(notifications),
		)
	}
}

func TestOwnCommentDoesNotCreateNotification(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"owncommentuser",
		"owncommentuser@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Own comment post",
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
		"My own comment",
	)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	err = CreateCommentNotification(
		db,
		postID,
		commentID,
		userID,
	)
	if err != nil {
		t.Fatalf(
			"create own comment notification: %v",
			err,
		)
	}

	notifications, err := GetNotificationsByUser(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get notifications: %v", err)
	}

	if len(notifications) != 0 {
		t.Fatalf(
			"expected 0 notifications, got %d",
			len(notifications),
		)
	}
}

func TestNotificationReadState(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"readstateuser",
		"readstateuser@example.com",
	)

	actorID := createTestUser(
		t,
		db,
		"readstateactor",
		"readstateactor@example.com",
	)

	postID, err := CreatePost(
		db,
		userID,
		"Read state post",
		"Test content",
		sql.NullString{},
	)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	firstID, err := CreateNotification(
		db,
		userID,
		sql.NullInt64{
			Int64: actorID,
			Valid: true,
		},
		NotificationPostLike,
		sql.NullInt64{
			Int64: postID,
			Valid: true,
		},
		sql.NullInt64{},
	)
	if err != nil {
		t.Fatalf(
			"create first notification: %v",
			err,
		)
	}

	_, err = CreateNotification(
		db,
		userID,
		sql.NullInt64{
			Int64: actorID,
			Valid: true,
		},
		NotificationPostDislike,
		sql.NullInt64{
			Int64: postID,
			Valid: true,
		},
		sql.NullInt64{},
	)
	if err != nil {
		t.Fatalf(
			"create second notification: %v",
			err,
		)
	}

	count, err := GetUnreadNotificationCount(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get unread count: %v", err)
	}

	if count != 2 {
		t.Fatalf(
			"expected 2 unread notifications, got %d",
			count,
		)
	}

	err = MarkNotificationAsRead(
		db,
		firstID,
		userID,
	)
	if err != nil {
		t.Fatalf("mark notification as read: %v", err)
	}

	count, err = GetUnreadNotificationCount(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get unread count: %v", err)
	}

	if count != 1 {
		t.Fatalf(
			"expected 1 unread notification, got %d",
			count,
		)
	}

	err = MarkAllNotificationsAsRead(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf(
			"mark all notifications as read: %v",
			err,
		)
	}

	count, err = GetUnreadNotificationCount(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get unread count: %v", err)
	}

	if count != 0 {
		t.Fatalf(
			"expected 0 unread notifications, got %d",
			count,
		)
	}
}
