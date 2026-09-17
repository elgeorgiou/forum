package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/session"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(
		t.TempDir(),
		"forum-test.db",
	)

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf(
			"open test database: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf(
				"close test database: %v",
				err,
			)
		}
	})

	schemaPath := filepath.Join(
		"..",
		"database",
		"schema",
		"schema.sql",
	)

	if err := database.InitSchema(
		db,
		schemaPath,
	); err != nil {
		t.Fatalf(
			"initialize test schema: %v",
			err,
		)
	}

	return db
}

func createTestUser(
	t *testing.T,
	db *sql.DB,
	username string,
	email string,
) *database.User {
	t.Helper()

	passwordHash, err := auth.HashPassword(
		"Password123!",
	)
	if err != nil {
		t.Fatalf(
			"hash password: %v",
			err,
		)
	}

	userID, err := database.CreateUser(
		db,
		username,
		email,
		passwordHash,
	)
	if err != nil {
		t.Fatalf(
			"create test user: %v",
			err,
		)
	}

	user, err := database.GetUserByID(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf(
			"get test user: %v",
			err,
		)
	}

	return user
}

func TestLoadUserWithoutSession(
	t *testing.T,
) {
	db := setupTestDB(t)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	var receivedUser *database.User

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			receivedUser = UserFromContext(
				r.Context(),
			)

			w.WriteHeader(http.StatusOK)
		},
	)

	handler := authMiddleware.LoadUser(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if receivedUser != nil {
		t.Fatal(
			"expected no authenticated user",
		)
	}
}

func TestLoadUserAddsUserToContext(
	t *testing.T,
) {
	db := setupTestDB(t)

	user := createTestUser(
		t,
		db,
		"middleware-user",
		"middleware@example.com",
	)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	sessionRecorder := httptest.NewRecorder()

	if err := sessionManager.Create(
		sessionRecorder,
		user.ID,
	); err != nil {
		t.Fatalf(
			"create session: %v",
			err,
		)
	}

	cookies :=
		sessionRecorder.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal(
			"expected session cookie",
		)
	}

	var receivedUser *database.User

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			receivedUser = UserFromContext(
				r.Context(),
			)

			w.WriteHeader(http.StatusOK)
		},
	)

	handler := authMiddleware.LoadUser(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(cookies[0])

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if receivedUser == nil {
		t.Fatal(
			"expected authenticated user",
		)
	}

	if receivedUser.ID != user.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			user.ID,
			receivedUser.ID,
		)
	}
}

func TestRequireAuthenticationAllowsUser(
	t *testing.T,
) {
	db := setupTestDB(t)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	user := &database.User{
		ID:       1,
		Username: "user",
		Role:     "user",
	}

	nextCalled := false

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		},
	)

	handler :=
		authMiddleware.RequireAuthentication(
			next,
		)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request = request.WithContext(
		WithUser(
			request.Context(),
			user,
		),
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if !nextCalled {
		t.Fatal(
			"expected next handler to be called",
		)
	}
}

func TestRequireAuthenticationRedirectsGuest(
	t *testing.T,
) {
	db := setupTestDB(t)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			t.Fatal(
				"next handler should not be called",
			)
		},
	)

	handler :=
		authMiddleware.RequireAuthentication(
			next,
		)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusSeeOther,
			recorder.Code,
		)
	}

	location := recorder.Header().Get(
		"Location",
	)

	if location != "/auth?mode=login" {
		t.Fatalf(
			"expected login redirect, got %q",
			location,
		)
	}
}

func TestRequireModeratorAllowsModerator(
	t *testing.T,
) {
	testRoleAllowed(
		t,
		"moderator",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireModerator(
				next,
			)
		},
	)
}

func TestRequireModeratorAllowsAdmin(
	t *testing.T,
) {
	testRoleAllowed(
		t,
		"admin",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireModerator(
				next,
			)
		},
	)
}

func TestRequireModeratorRejectsUser(
	t *testing.T,
) {
	testRoleForbidden(
		t,
		"user",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireModerator(
				next,
			)
		},
	)
}

func TestRequireAdminAllowsAdmin(
	t *testing.T,
) {
	testRoleAllowed(
		t,
		"admin",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireAdmin(
				next,
			)
		},
	)
}

func TestRequireAdminRejectsModerator(
	t *testing.T,
) {
	testRoleForbidden(
		t,
		"moderator",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireAdmin(
				next,
			)
		},
	)
}

func TestRequireAdminRejectsUser(
	t *testing.T,
) {
	testRoleForbidden(
		t,
		"user",
		func(
			middleware *Auth,
			next http.Handler,
		) http.Handler {
			return middleware.RequireAdmin(
				next,
			)
		},
	)
}

func TestUserFromContextWithoutUser(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	user := UserFromContext(
		request.Context(),
	)

	if user != nil {
		t.Fatal("expected nil user")
	}
}

func TestWithUser(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	expected := &database.User{
		ID:       42,
		Username: "context-user",
		Email:    "context@example.com",
		Role:     "user",
	}

	ctx := WithUser(
		request.Context(),
		expected,
	)

	actual := UserFromContext(ctx)

	if actual == nil {
		t.Fatal("expected user")
	}

	if actual.ID != expected.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			expected.ID,
			actual.ID,
		)
	}
}

func testRoleAllowed(
	t *testing.T,
	role string,
	wrap func(*Auth, http.Handler) http.Handler,
) {
	t.Helper()

	db := setupTestDB(t)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	user := &database.User{
		ID:       1,
		Username: "role-user",
		Role:     role,
	}

	nextCalled := false

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := wrap(
		authMiddleware,
		next,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request = request.WithContext(
		WithUser(
			request.Context(),
			user,
		),
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"role %q: expected status %d, got %d",
			role,
			http.StatusOK,
			recorder.Code,
		)
	}

	if !nextCalled {
		t.Fatalf(
			"role %q: expected next handler",
			role,
		)
	}
}

func testRoleForbidden(
	t *testing.T,
	role string,
	wrap func(*Auth, http.Handler) http.Handler,
) {
	t.Helper()

	db := setupTestDB(t)

	sessionManager := session.NewManager(db)
	authMiddleware := NewAuth(sessionManager)

	user := &database.User{
		ID:       1,
		Username: "role-user",
		Role:     role,
	}

	next := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			t.Fatal(
				"next handler should not be called",
			)
		},
	)

	handler := wrap(
		authMiddleware,
		next,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request = request.WithContext(
		WithUser(
			request.Context(),
			user,
		),
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"role %q: expected status %d, got %d",
			role,
			http.StatusForbidden,
			recorder.Code,
		)
	}
}
