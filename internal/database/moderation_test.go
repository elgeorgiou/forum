package database

import (
	"errors"
	"testing"
)

func TestCreateModeratorRequest(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"moderator_request_user",
		"moderator_request@example.com",
	)

	requestID, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf("CreateModeratorRequest() error = %v", err)
	}

	if requestID == 0 {
		t.Fatal("expected moderator request ID")
	}

	request, err := GetPendingModeratorRequestByUser(db, userID)
	if err != nil {
		t.Fatalf(
			"GetPendingModeratorRequestByUser() error = %v",
			err,
		)
	}

	if request.UserID != userID {
		t.Fatalf(
			"request UserID = %d, want %d",
			request.UserID,
			userID,
		)
	}

	if request.Status != "pending" {
		t.Fatalf(
			"request Status = %q, want pending",
			request.Status,
		)
	}
}

func TestCreateModeratorRequestRejectsDuplicatePendingRequest(
	t *testing.T,
) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"duplicate_request_user",
		"duplicate_request@example.com",
	)

	_, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf(
			"first CreateModeratorRequest() error = %v",
			err,
		)
	}

	_, err = CreateModeratorRequest(db, userID)

	if !errors.Is(err, ErrModeratorRequestExists) {
		t.Fatalf(
			"second CreateModeratorRequest() error = %v, want %v",
			err,
			ErrModeratorRequestExists,
		)
	}
}

func TestGetPendingModeratorRequestViews(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"request_view_user",
		"request_view@example.com",
	)

	_, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf(
			"CreateModeratorRequest() error = %v",
			err,
		)
	}

	requests, err := GetPendingModeratorRequestViews(db)
	if err != nil {
		t.Fatalf(
			"GetPendingModeratorRequestViews() error = %v",
			err,
		)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"len(requests) = %d, want 1",
			len(requests),
		)
	}

	if requests[0].UserID != userID {
		t.Fatalf(
			"UserID = %d, want %d",
			requests[0].UserID,
			userID,
		)
	}

	if requests[0].Username != "request_view_user" {
		t.Fatalf(
			"Username = %q, want %q",
			requests[0].Username,
			"request_view_user",
		)
	}

	if requests[0].Email != "request_view@example.com" {
		t.Fatalf(
			"Email = %q, want %q",
			requests[0].Email,
			"request_view@example.com",
		)
	}
}

func TestApproveModeratorRequestPromotesUser(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"promotion_user",
		"promotion_user@example.com",
	)

	adminID := createTestUser(
		t,
		db,
		"promotion_admin",
		"promotion_admin@example.com",
	)

	err := UpdateUserRole(db, adminID, "admin")
	if err != nil {
		t.Fatalf(
			"UpdateUserRole() error = %v",
			err,
		)
	}

	requestID, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf(
			"CreateModeratorRequest() error = %v",
			err,
		)
	}

	err = ReviewModeratorRequest(
		db,
		requestID,
		"approved",
		adminID,
	)
	if err != nil {
		t.Fatalf(
			"ReviewModeratorRequest() error = %v",
			err,
		)
	}

	user, err := GetUserByID(db, userID)
	if err != nil {
		t.Fatalf(
			"GetUserByID() error = %v",
			err,
		)
	}

	if user.Role != "moderator" {
		t.Fatalf(
			"user Role = %q, want moderator",
			user.Role,
		)
	}

	request, err := GetModeratorRequestByID(
		db,
		requestID,
	)
	if err != nil {
		t.Fatalf(
			"GetModeratorRequestByID() error = %v",
			err,
		)
	}

	if request.Status != "approved" {
		t.Fatalf(
			"request Status = %q, want approved",
			request.Status,
		)
	}

	if !request.ReviewedBy.Valid {
		t.Fatal("expected ReviewedBy to be set")
	}

	if request.ReviewedBy.Int64 != adminID {
		t.Fatalf(
			"ReviewedBy = %d, want %d",
			request.ReviewedBy.Int64,
			adminID,
		)
	}
}

func TestRejectModeratorRequestDoesNotPromoteUser(
	t *testing.T,
) {
	db := setupTestDB(t)

	userID := createTestUser(
		t,
		db,
		"rejected_user",
		"rejected_user@example.com",
	)

	adminID := createTestUser(
		t,
		db,
		"rejected_admin",
		"rejected_admin@example.com",
	)

	err := UpdateUserRole(db, adminID, "admin")
	if err != nil {
		t.Fatalf(
			"UpdateUserRole() error = %v",
			err,
		)
	}

	requestID, err := CreateModeratorRequest(db, userID)
	if err != nil {
		t.Fatalf(
			"CreateModeratorRequest() error = %v",
			err,
		)
	}

	err = ReviewModeratorRequest(
		db,
		requestID,
		"rejected",
		adminID,
	)
	if err != nil {
		t.Fatalf(
			"ReviewModeratorRequest() error = %v",
			err,
		)
	}

	user, err := GetUserByID(db, userID)
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

	request, err := GetModeratorRequestByID(
		db,
		requestID,
	)
	if err != nil {
		t.Fatalf(
			"GetModeratorRequestByID() error = %v",
			err,
		)
	}

	if request.Status != "rejected" {
		t.Fatalf(
			"request Status = %q, want rejected",
			request.Status,
		)
	}
}
