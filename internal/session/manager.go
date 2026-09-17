package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"forum/internal/database"
)

const (
	CookieName      = "forum_session"
	SessionDuration = 24 * time.Hour
	tokenBytes      = 32
)

var ErrUnauthenticated = errors.New("user is not authenticated")

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) Create(
	w http.ResponseWriter,
	userID int64,
) error {
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate session token: %w", err)
	}

	expiresAt := time.Now().
		UTC().
		Add(SessionDuration)

	err = database.CreateSession(
		m.db,
		token,
		userID,
		expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (m *Manager) Destroy(
	w http.ResponseWriter,
	r *http.Request,
) error {
	cookie, err := r.Cookie(CookieName)

	if err == nil && cookie.Value != "" {
		err = database.DeleteSessionByID(
			m.db,
			cookie.Value,
		)

		if err != nil &&
			!errors.Is(err, database.ErrSessionNotFound) {
			return fmt.Errorf("delete session: %w", err)
		}
	} else if err != nil &&
		!errors.Is(err, http.ErrNoCookie) {
		return fmt.Errorf("read session cookie: %w", err)
	}

	expireCookie(w)

	return nil
}

func (m *Manager) User(
	r *http.Request,
) (*database.User, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, ErrUnauthenticated
		}

		return nil, fmt.Errorf(
			"read session cookie: %w",
			err,
		)
	}

	if cookie.Value == "" {
		return nil, ErrUnauthenticated
	}

	currentSession, err := database.GetSessionByID(
		m.db,
		cookie.Value,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrSessionNotFound,
		) {
			return nil, ErrUnauthenticated
		}

		return nil, fmt.Errorf(
			"get session: %w",
			err,
		)
	}

	expiresAt, err := parseTime(
		currentSession.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse session expiration: %w",
			err,
		)
	}

	if !expiresAt.After(time.Now().UTC()) {
		err = database.DeleteSessionByID(
			m.db,
			currentSession.ID,
		)

		if err != nil &&
			!errors.Is(
				err,
				database.ErrSessionNotFound,
			) {
			return nil, fmt.Errorf(
				"delete expired session: %w",
				err,
			)
		}

		return nil, ErrUnauthenticated
	}

	user, err := database.GetUserByID(
		m.db,
		currentSession.UserID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrUserNotFound,
		) {
			return nil, ErrUnauthenticated
		}

		return nil, fmt.Errorf(
			"get session user: %w",
			err,
		)
	}

	return user, nil
}

func (m *Manager) DeleteExpired() error {
	if err := database.DeleteExpiredSessions(m.db); err != nil {
		return fmt.Errorf(
			"delete expired sessions: %w",
			err,
		)
	}

	return nil
}

func generateToken() (string, error) {
	randomBytes := make([]byte, tokenBytes)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

func parseTime(value string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"unsupported time format %q",
		value,
	)
}

func expireCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func secureCookies() bool {
	return os.Getenv("FORUM_SECURE_COOKIES") == "true"
}
