package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/session"
)

func setupAuthTestDB(t *testing.T) *sql.DB {
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
			"initialize schema: %v",
			err,
		)
	}

	return db
}

func newAuthTestHandler(
	t *testing.T,
) (*AuthHandler, *sql.DB) {
	t.Helper()

	db := setupAuthTestDB(t)

	sessionManager := session.NewManager(db)
	pageHandler := NewPageHandler(db)

	pageHandler.templateDir = filepath.Join(
		"..",
		"..",
		"templates",
	)

	handler := NewAuthHandler(
		db,
		sessionManager,
		pageHandler,
	)

	return handler, db
}

func newFormRequest(
	method string,
	target string,
	values url.Values,
) *http.Request {
	request := httptest.NewRequest(
		method,
		target,
		strings.NewReader(values.Encode()),
	)

	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	return request
}

func createAuthTestUser(
	t *testing.T,
	db *sql.DB,
	username string,
	email string,
	password string,
) *database.User {
	t.Helper()

	passwordHash, err := auth.HashPassword(
		password,
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
			"create user: %v",
			err,
		)
	}

	user, err := database.GetUserByID(
		db,
		userID,
	)
	if err != nil {
		t.Fatalf(
			"get user: %v",
			err,
		)
	}

	return user
}

func TestRegisterCreatesUser(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"new-player",
			},
			"email": {
				"player@example.com",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
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

	user, err := database.GetUserByEmail(
		db,
		"player@example.com",
	)
	if err != nil {
		t.Fatalf(
			"get registered user: %v",
			err,
		)
	}

	if user.Username != "new-player" {
		t.Fatalf(
			"expected username %q, got %q",
			"new-player",
			user.Username,
		)
	}

	if user.PasswordHash == "Password123!" {
		t.Fatal(
			"expected password to be hashed",
		)
	}

	if err := auth.CheckPassword(
		user.PasswordHash,
		"Password123!",
	); err != nil {
		t.Fatalf(
			"check stored password: %v",
			err,
		)
	}
}

func TestRegisterCreatesSession(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"session-player",
			},
			"email": {
				"session@example.com",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	user, err := database.GetUserByEmail(
		db,
		"session@example.com",
	)
	if err != nil {
		t.Fatalf(
			"get user: %v",
			err,
		)
	}

	currentSession, err :=
		database.GetSessionByUserID(
			db,
			user.ID,
		)
	if err != nil {
		t.Fatalf(
			"get session: %v",
			err,
		)
	}

	if currentSession.UserID != user.ID {
		t.Fatalf(
			"expected session user ID %d, got %d",
			user.ID,
			currentSession.UserID,
		)
	}

	if len(
		recorder.Result().Cookies(),
	) == 0 {
		t.Fatal(
			"expected session cookie",
		)
	}
}

func TestRegisterRejectsEmptyFields(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestRegisterRejectsInvalidEmail(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"player",
			},
			"email": {
				"not-an-email",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestRegisterRejectsPasswordMismatch(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"player",
			},
			"email": {
				"player@example.com",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"DifferentPassword!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestRegisterRejectsDuplicateUsername(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	createAuthTestUser(
		t,
		db,
		"existing-player",
		"existing@example.com",
		"Password123!",
	)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"existing-player",
			},
			"email": {
				"different@example.com",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}
}

func TestRegisterRejectsDuplicateEmail(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	createAuthTestUser(
		t,
		db,
		"existing-player",
		"existing@example.com",
		"Password123!",
	)

	request := newFormRequest(
		http.MethodPost,
		"/register",
		url.Values{
			"username": {
				"different-player",
			},
			"email": {
				"existing@example.com",
			},
			"password": {
				"Password123!",
			},
			"confirm_password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Register(
		recorder,
		request,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}
}

func TestLoginWithUsername(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	user := createAuthTestUser(
		t,
		db,
		"login-player",
		"login@example.com",
		"Password123!",
	)

	request := newFormRequest(
		http.MethodPost,
		"/login",
		url.Values{
			"identifier": {
				"login-player",
			},
			"password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Login(
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

	_, err := database.GetSessionByUserID(
		db,
		user.ID,
	)
	if err != nil {
		t.Fatalf(
			"expected session: %v",
			err,
		)
	}
}

func TestLoginWithEmail(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	user := createAuthTestUser(
		t,
		db,
		"email-player",
		"email-login@example.com",
		"Password123!",
	)

	request := newFormRequest(
		http.MethodPost,
		"/login",
		url.Values{
			"identifier": {
				"email-login@example.com",
			},
			"password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Login(
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

	_, err := database.GetSessionByUserID(
		db,
		user.ID,
	)
	if err != nil {
		t.Fatalf(
			"expected session: %v",
			err,
		)
	}
}

func TestLoginRejectsWrongPassword(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	createAuthTestUser(
		t,
		db,
		"wrong-password-player",
		"wrong-password@example.com",
		"Password123!",
	)

	request := newFormRequest(
		http.MethodPost,
		"/login",
		url.Values{
			"identifier": {
				"wrong-password-player",
			},
			"password": {
				"WrongPassword!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Login(
		recorder,
		request,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestLoginRejectsUnknownUser(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/login",
		url.Values{
			"identifier": {
				"missing-player",
			},
			"password": {
				"Password123!",
			},
		},
	)

	recorder := httptest.NewRecorder()

	handler.Login(
		recorder,
		request,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestLoginRejectsEmptyFields(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	request := newFormRequest(
		http.MethodPost,
		"/login",
		url.Values{},
	)

	recorder := httptest.NewRecorder()

	handler.Login(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestLogoutDeletesSession(
	t *testing.T,
) {
	handler, db := newAuthTestHandler(t)

	user := createAuthTestUser(
		t,
		db,
		"logout-player",
		"logout@example.com",
		"Password123!",
	)

	sessionManager := session.NewManager(db)

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

	request := httptest.NewRequest(
		http.MethodPost,
		"/logout",
		nil,
	)

	request.AddCookie(cookies[0])

	recorder := httptest.NewRecorder()

	handler.Logout(
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

	_, err := database.GetSessionByUserID(
		db,
		user.ID,
	)

	if err != database.ErrSessionNotFound {
		t.Fatalf(
			"expected session removed, got %v",
			err,
		)
	}

	expiredCookies :=
		recorder.Result().Cookies()

	if len(expiredCookies) == 0 {
		t.Fatal(
			"expected expired session cookie",
		)
	}

	if expiredCookies[0].MaxAge != -1 {
		t.Fatalf(
			"expected MaxAge -1, got %d",
			expiredCookies[0].MaxAge,
		)
	}
}

func TestAuthHandlersRejectWrongMethod(
	t *testing.T,
) {
	handler, _ := newAuthTestHandler(t)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  string
	}{
		{
			name:    "register",
			handler: handler.Register,
			target:  "/register",
		},
		{
			name:    "login",
			handler: handler.Login,
			target:  "/login",
		},
		{
			name:    "logout",
			handler: handler.Logout,
			target:  "/logout",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				request := httptest.NewRequest(
					http.MethodGet,
					test.target,
					nil,
				)

				recorder :=
					httptest.NewRecorder()

				test.handler(
					recorder,
					request,
				)

				if recorder.Code !=
					http.StatusMethodNotAllowed {
					t.Fatalf(
						"expected status %d, got %d",
						http.StatusMethodNotAllowed,
						recorder.Code,
					)
				}
			},
		)
	}
}

func TestValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{
			email: "player@example.com",
			valid: true,
		},
		{
			email: "not-an-email",
			valid: false,
		},
		{
			email: "",
			valid: false,
		},
	}

	for _, test := range tests {
		actual := validEmail(test.email)

		if actual != test.valid {
			t.Fatalf(
				"validEmail(%q): expected %t, got %t",
				test.email,
				test.valid,
				actual,
			)
		}
	}
}
