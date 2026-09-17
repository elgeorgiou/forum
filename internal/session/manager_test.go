package session

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(
		t.TempDir(),
		"forum-test.db",
	)

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
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
) *database.User {
	t.Helper()

	passwordHash, err := auth.HashPassword(
		"Password123!",
	)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userID, err := database.CreateUser(
		db,
		"session-user",
		"session@example.com",
		passwordHash,
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	user, err := database.GetUserByID(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf("get test user: %v", err)
	}

	return user
}

func TestNewManager(t *testing.T) {
	db := setupTestDB(t)

	manager := NewManager(db)

	if manager == nil {
		t.Fatal("expected manager, got nil")
	}

	if manager.db != db {
		t.Fatal(
			"manager does not contain expected database",
		)
	}
}

func TestCreateSessionSetsCookie(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	manager := NewManager(db)
	recorder := httptest.NewRecorder()

	err := manager.Create(
		recorder,
		user.ID,
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	response := recorder.Result()
	defer response.Body.Close()

	cookies := response.Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 cookie, got %d",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if cookie.Name != CookieName {
		t.Fatalf(
			"expected cookie name %q, got %q",
			CookieName,
			cookie.Name,
		)
	}

	if cookie.Value == "" {
		t.Fatal(
			"expected non-empty session token",
		)
	}

	if cookie.Path != "/" {
		t.Fatalf(
			"expected cookie path /, got %q",
			cookie.Path,
		)
	}

	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf(
			"expected SameSiteLaxMode, got %v",
			cookie.SameSite,
		)
	}

	if cookie.MaxAge !=
		int(SessionDuration.Seconds()) {
		t.Fatalf(
			"expected MaxAge %d, got %d",
			int(SessionDuration.Seconds()),
			cookie.MaxAge,
		)
	}
}

func TestCreateSessionStoresSession(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	manager := NewManager(db)
	recorder := httptest.NewRecorder()

	err := manager.Create(
		recorder,
		user.ID,
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	cookies := recorder.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	storedSession, err :=
		database.GetSessionByID(
			db,
			cookies[0].Value,
		)
	if err != nil {
		t.Fatalf(
			"get stored session: %v",
			err,
		)
	}

	if storedSession.UserID != user.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			user.ID,
			storedSession.UserID,
		)
	}

	if storedSession.ID != cookies[0].Value {
		t.Fatalf(
			"expected session ID %q, got %q",
			cookies[0].Value,
			storedSession.ID,
		)
	}
}

func TestCreateSessionReplacesExistingSession(
	t *testing.T,
) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	manager := NewManager(db)

	firstRecorder := httptest.NewRecorder()

	if err := manager.Create(
		firstRecorder,
		user.ID,
	); err != nil {
		t.Fatalf(
			"create first session: %v",
			err,
		)
	}

	firstCookies :=
		firstRecorder.Result().Cookies()

	if len(firstCookies) == 0 {
		t.Fatal(
			"expected first session cookie",
		)
	}

	firstToken := firstCookies[0].Value

	secondRecorder := httptest.NewRecorder()

	if err := manager.Create(
		secondRecorder,
		user.ID,
	); err != nil {
		t.Fatalf(
			"create second session: %v",
			err,
		)
	}

	secondCookies :=
		secondRecorder.Result().Cookies()

	if len(secondCookies) == 0 {
		t.Fatal(
			"expected second session cookie",
		)
	}

	secondToken := secondCookies[0].Value

	if firstToken == secondToken {
		t.Fatal(
			"expected new session token",
		)
	}

	_, err := database.GetSessionByID(
		db,
		firstToken,
	)

	if !errors.Is(
		err,
		database.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected old session removed, got %v",
			err,
		)
	}

	currentSession, err :=
		database.GetSessionByID(
			db,
			secondToken,
		)
	if err != nil {
		t.Fatalf(
			"get new session: %v",
			err,
		)
	}

	if currentSession.UserID != user.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			user.ID,
			currentSession.UserID,
		)
	}
}

func TestUserReturnsAuthenticatedUser(
	t *testing.T,
) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	manager := NewManager(db)
	recorder := httptest.NewRecorder()

	if err := manager.Create(
		recorder,
		user.ID,
	); err != nil {
		t.Fatalf(
			"create session: %v",
			err,
		)
	}

	cookies := recorder.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(cookies[0])

	currentUser, err :=
		manager.User(request)
	if err != nil {
		t.Fatalf(
			"get session user: %v",
			err,
		)
	}

	if currentUser == nil {
		t.Fatal(
			"expected authenticated user",
		)
	}

	if currentUser.ID != user.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			user.ID,
			currentUser.ID,
		)
	}

	if currentUser.Username != user.Username {
		t.Fatalf(
			"expected username %q, got %q",
			user.Username,
			currentUser.Username,
		)
	}
}

func TestUserRejectsMissingCookie(
	t *testing.T,
) {
	db := setupTestDB(t)
	manager := NewManager(db)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	user, err := manager.User(request)

	if user != nil {
		t.Fatal("expected nil user")
	}

	if !errors.Is(
		err,
		ErrUnauthenticated,
	) {
		t.Fatalf(
			"expected ErrUnauthenticated, got %v",
			err,
		)
	}
}

func TestUserRejectsUnknownSession(
	t *testing.T,
) {
	db := setupTestDB(t)
	manager := NewManager(db)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: "unknown-session",
	})

	user, err := manager.User(request)

	if user != nil {
		t.Fatal("expected nil user")
	}

	if !errors.Is(
		err,
		ErrUnauthenticated,
	) {
		t.Fatalf(
			"expected ErrUnauthenticated, got %v",
			err,
		)
	}
}

func TestUserRejectsExpiredSession(
	t *testing.T,
) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	token := "expired-session-token"

	expiresAt := time.Now().
		UTC().
		Add(-time.Hour).
		Format(time.RFC3339)

	err := database.CreateSession(
		db,
		token,
		user.ID,
		expiresAt,
	)
	if err != nil {
		t.Fatalf(
			"create expired session: %v",
			err,
		)
	}

	manager := NewManager(db)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	request.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: token,
	})

	currentUser, err :=
		manager.User(request)

	if currentUser != nil {
		t.Fatal("expected nil user")
	}

	if !errors.Is(
		err,
		ErrUnauthenticated,
	) {
		t.Fatalf(
			"expected ErrUnauthenticated, got %v",
			err,
		)
	}

	_, err = database.GetSessionByID(
		db,
		token,
	)

	if !errors.Is(
		err,
		database.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected expired session deleted, got %v",
			err,
		)
	}
}

func TestDestroyDeletesSession(
	t *testing.T,
) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	manager := NewManager(db)
	createRecorder :=
		httptest.NewRecorder()

	if err := manager.Create(
		createRecorder,
		user.ID,
	); err != nil {
		t.Fatalf(
			"create session: %v",
			err,
		)
	}

	cookies :=
		createRecorder.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/logout",
		nil,
	)

	request.AddCookie(cookies[0])

	destroyRecorder :=
		httptest.NewRecorder()

	err := manager.Destroy(
		destroyRecorder,
		request,
	)
	if err != nil {
		t.Fatalf(
			"destroy session: %v",
			err,
		)
	}

	_, err = database.GetSessionByID(
		db,
		cookies[0].Value,
	)

	if !errors.Is(
		err,
		database.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected session deleted, got %v",
			err,
		)
	}

	responseCookies :=
		destroyRecorder.Result().Cookies()

	if len(responseCookies) == 0 {
		t.Fatal("expected expired cookie")
	}

	if responseCookies[0].MaxAge != -1 {
		t.Fatalf(
			"expected MaxAge -1, got %d",
			responseCookies[0].MaxAge,
		)
	}
}

func TestDestroyWithoutSession(
	t *testing.T,
) {
	db := setupTestDB(t)
	manager := NewManager(db)

	request := httptest.NewRequest(
		http.MethodPost,
		"/logout",
		nil,
	)

	recorder := httptest.NewRecorder()

	err := manager.Destroy(
		recorder,
		request,
	)
	if err != nil {
		t.Fatalf(
			"destroy session: %v",
			err,
		)
	}

	cookies := recorder.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected expired cookie")
	}

	if cookies[0].MaxAge != -1 {
		t.Fatalf(
			"expected MaxAge -1, got %d",
			cookies[0].MaxAge,
		)
	}
}

func TestDeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	token := "expired-token"

	err := database.CreateSession(
		db,
		token,
		user.ID,
		time.Now().
			UTC().
			Add(-time.Hour).
			Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf(
			"create expired session: %v",
			err,
		)
	}

	manager := NewManager(db)

	if err := manager.DeleteExpired(); err != nil {
		t.Fatalf(
			"delete expired sessions: %v",
			err,
		)
	}

	_, err = database.GetSessionByID(
		db,
		token,
	)

	if !errors.Is(
		err,
		database.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected expired session removed, got %v",
			err,
		)
	}
}

func TestParseTimeRFC3339(t *testing.T) {
	value := "2026-09-18T20:15:29Z"

	parsed, err := parseTime(value)
	if err != nil {
		t.Fatalf(
			"parse RFC3339 time: %v",
			err,
		)
	}

	expected := time.Date(
		2026,
		time.September,
		18,
		20,
		15,
		29,
		0,
		time.UTC,
	)

	if !parsed.Equal(expected) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			parsed,
		)
	}
}

func TestParseTimeSQLiteFormat(
	t *testing.T,
) {
	value := "2026-09-18 20:15:29"

	parsed, err := parseTime(value)
	if err != nil {
		t.Fatalf(
			"parse SQLite time: %v",
			err,
		)
	}

	expected := time.Date(
		2026,
		time.September,
		18,
		20,
		15,
		29,
		0,
		time.UTC,
	)

	if !parsed.Equal(expected) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			parsed,
		)
	}
}

func TestParseTimeRejectsInvalidValue(
	t *testing.T,
) {
	_, err := parseTime("not-a-time")

	if err == nil {
		t.Fatal(
			"expected invalid time error",
		)
	}
}

func TestGenerateToken(t *testing.T) {
	firstToken, err := generateToken()
	if err != nil {
		t.Fatalf(
			"generate first token: %v",
			err,
		)
	}

	secondToken, err := generateToken()
	if err != nil {
		t.Fatalf(
			"generate second token: %v",
			err,
		)
	}

	if firstToken == "" {
		t.Fatal(
			"expected non-empty token",
		)
	}

	if len(firstToken) != tokenBytes*2 {
		t.Fatalf(
			"expected token length %d, got %d",
			tokenBytes*2,
			len(firstToken),
		)
	}

	if firstToken == secondToken {
		t.Fatal(
			"expected unique tokens",
		)
	}
}

func TestSecureCookies(t *testing.T) {
	originalValue, existed :=
		os.LookupEnv(
			"FORUM_SECURE_COOKIES",
		)

	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(
				"FORUM_SECURE_COOKIES",
				originalValue,
			); err != nil {
				t.Errorf(
					"restore environment: %v",
					err,
				)
			}

			return
		}

		if err := os.Unsetenv(
			"FORUM_SECURE_COOKIES",
		); err != nil {
			t.Errorf(
				"unset environment: %v",
				err,
			)
		}
	})

	if err := os.Setenv(
		"FORUM_SECURE_COOKIES",
		"true",
	); err != nil {
		t.Fatalf(
			"set environment: %v",
			err,
		)
	}

	if !secureCookies() {
		t.Fatal(
			"expected secure cookies",
		)
	}

	if err := os.Setenv(
		"FORUM_SECURE_COOKIES",
		"false",
	); err != nil {
		t.Fatalf(
			"set environment: %v",
			err,
		)
	}

	if secureCookies() {
		t.Fatal(
			"expected non-secure cookies",
		)
	}
}
