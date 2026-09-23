package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/session"
)

// AuthHandler handles user registration, login, logout, and authentication-related operations.
type AuthHandler struct {
	db       *sql.DB
	sessions *session.Manager
	pages    *PageHandler
}

// NewAuthHandler creates a new AuthHandler with its required dependencies.
func NewAuthHandler(
	db *sql.DB,
	sessions *session.Manager,
	pages *PageHandler,
) *AuthHandler {
	return &AuthHandler{
		db:       db,
		sessions: sessions,
		pages:    pages,
	}
}

// Register handles the creation of a new user account and starts a session for the user.
func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		h.pages.renderMethodNotAllowed(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.pages.renderBadRequest(w)
		return
	}

	username := strings.TrimSpace(
		r.FormValue("username"),
	)

	email := strings.TrimSpace(
		r.FormValue("email"),
	)

	password := r.FormValue("password")

	confirmPassword := r.FormValue(
		"confirm_password",
	)

	data := models.PageData{
		AuthMode: "signup",
		Form: models.AuthFormData{
			Username: username,
			Email:    email,
		},
	}

	if username == "" ||
		email == "" ||
		strings.TrimSpace(password) == "" {
		data.Error = "Username, email and password are required."

		h.pages.render(
			w,
			http.StatusBadRequest,
			"auth.html",
			data,
		)
		return
	}

	if confirmPassword != "" &&
		password != confirmPassword {
		data.Error = "Passwords do not match."

		h.pages.render(
			w,
			http.StatusBadRequest,
			"auth.html",
			data,
		)
		return
	}

	if !validEmail(email) {
		data.Error = "Please enter a valid email address."

		h.pages.render(
			w,
			http.StatusBadRequest,
			"auth.html",
			data,
		)
		return
	}

	_, err := database.GetUserByUsername(
		h.db,
		username,
	)

	switch {
	case err == nil:
		data.Error = "Username is already in use."

		h.pages.render(
			w,
			http.StatusConflict,
			"auth.html",
			data,
		)
		return

	case !errors.Is(err, database.ErrUserNotFound):
		h.pages.renderInternalServerError(w)
		return
	}

	_, err = database.GetUserByEmail(
		h.db,
		email,
	)

	switch {
	case err == nil:
		data.Error = "Email is already in use."

		h.pages.render(
			w,
			http.StatusConflict,
			"auth.html",
			data,
		)
		return

	case !errors.Is(err, database.ErrUserNotFound):
		h.pages.renderInternalServerError(w)
		return
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		if errors.Is(err, auth.ErrEmptyPassword) {
			data.Error = "Password is required."

			h.pages.render(
				w,
				http.StatusBadRequest,
				"auth.html",
				data,
			)
			return
		}

		h.pages.renderInternalServerError(w)
		return
	}

	userID, err := database.CreateUser(
		h.db,
		username,
		email,
		passwordHash,
	)
	if err != nil {
		if isDuplicateUserError(err) {
			data.Error = "Username or email is already in use."

			h.pages.render(
				w,
				http.StatusConflict,
				"auth.html",
				data,
			)
			return
		}

		h.pages.renderInternalServerError(w)
		return
	}

	if err := h.sessions.Create(
		w,
		userID,
	); err != nil {
		h.pages.renderInternalServerError(w)
		return
	}

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}

// Login authenticates a user by email or username and starts a new session.
func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		h.pages.renderMethodNotAllowed(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.pages.renderBadRequest(w)
		return
	}

	identifier := strings.TrimSpace(
		r.FormValue("identifier"),
	)

	password := r.FormValue("password")

	data := models.PageData{
		AuthMode: "login",
		Form: models.AuthFormData{
			Identifier: identifier,
		},
	}

	if identifier == "" ||
		strings.TrimSpace(password) == "" {
		data.Error = "Email or username and password are required."

		h.pages.render(
			w,
			http.StatusBadRequest,
			"auth.html",
			data,
		)
		return
	}

	user, err := h.findUser(identifier)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			data.Error = "Incorrect email, username or password."

			h.pages.render(
				w,
				http.StatusUnauthorized,
				"auth.html",
				data,
			)
			return
		}

		h.pages.renderInternalServerError(w)
		return
	}

	if err := auth.CheckPassword(
		user.PasswordHash,
		password,
	); err != nil {
		data.Error = "Incorrect email, username or password."

		h.pages.render(
			w,
			http.StatusUnauthorized,
			"auth.html",
			data,
		)
		return
	}

	if err := h.sessions.Create(
		w,
		user.ID,
	); err != nil {
		h.pages.renderInternalServerError(w)
		return
	}

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}

// Logout destroys the current user session and redirects to the home page.
func (h *AuthHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		h.pages.renderMethodNotAllowed(w)
		return
	}

	if err := h.sessions.Destroy(
		w,
		r,
	); err != nil {
		h.pages.renderInternalServerError(w)
		return
	}

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}

// findUser retrieves a user by email when the identifier contains an @,
// otherwise it retrieves the user by username.
func (h *AuthHandler) findUser(
	identifier string,
) (*database.User, error) {
	if strings.Contains(identifier, "@") {
		return database.GetUserByEmail(
			h.db,
			identifier,
		)
	}

	return database.GetUserByUsername(
		h.db,
		identifier,
	)
}

// validEmail reports whether email is a valid email address.
func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return strings.EqualFold(
		address.Address,
		email,
	)
}

// isDuplicateUserError reports whether an error was caused by a database uniqueness constraint.
func isDuplicateUserError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(
		err.Error(),
	)

	return strings.Contains(
		message,
		"unique constraint failed",
	) ||
		strings.Contains(
			message,
			"constraint failed",
		)
}