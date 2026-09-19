package database

import (
	"errors"
	"testing"
)

func TestGetModerators(t *testing.T) {
	db := setupTestDB(t)

	moderatorID := createTestUser(
		t,
		db,
		"active_moderator",
		"active_moderator@example.com",
	)

	createTestUser(
		t,
		db,
		"normal_user",
		"normal_user@example.com",
	)

	if err := UpdateUserRole(
		db,
		moderatorID,
		"moderator",
	); err != nil {
		t.Fatalf(
			"UpdateUserRole() error = %v",
			err,
		)
	}

	moderators, err := GetModerators(db)
	if err != nil {
		t.Fatalf(
			"GetModerators() error = %v",
			err,
		)
	}

	if len(moderators) != 1 {
		t.Fatalf(
			"len(moderators) = %d, want 1",
			len(moderators),
		)
	}

	if moderators[0].ID != moderatorID {
		t.Fatalf(
			"moderator ID = %d, want %d",
			moderators[0].ID,
			moderatorID,
		)
	}

	if moderators[0].Role != "moderator" {
		t.Fatalf(
			"moderator Role = %q, want moderator",
			moderators[0].Role,
		)
	}
}

func TestDemoteModerator(t *testing.T) {
	db := setupTestDB(t)

	moderatorID := createTestUser(
		t,
		db,
		"demote_moderator",
		"demote_moderator@example.com",
	)

	if err := UpdateUserRole(
		db,
		moderatorID,
		"moderator",
	); err != nil {
		t.Fatalf(
			"UpdateUserRole() error = %v",
			err,
		)
	}

	if err := DemoteModerator(
		db,
		moderatorID,
	); err != nil {
		t.Fatalf(
			"DemoteModerator() error = %v",
			err,
		)
	}

	user, err := GetUserByID(
		db,
		moderatorID,
	)
	if err != nil {
		t.Fatalf(
			"GetUserByID() error = %v",
			err,
		)
	}

	if user.Role != "user" {
		t.Fatalf(
			"user Role = %q, want user",
			user.Role,
		)
	}
}

func TestDemoteModeratorRejectsNormalUser(
	t *testing.T,
) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"not_moderator",
		"not_moderator@example.com",
	)

	err := DemoteModerator(
		db,
		userID,
	)

	if !errors.Is(
		err,
		ErrModeratorNotFound,
	) {
		t.Fatalf(
			"DemoteModerator() error = %v, want %v",
			err,
			ErrModeratorNotFound,
		)
	}
}
