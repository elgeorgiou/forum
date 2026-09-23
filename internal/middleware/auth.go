package middleware

import (
	"context"
	"errors"
	"net/http"

	"forum/internal/database"
	"forum/internal/session"
)

// contextKey defines a private type for context keys used by this package.
type contextKey string

// userContextKey is the context key used to store the authenticated user.
const userContextKey contextKey = "current_user"

// ErrorRenderer defines a function that renders an HTTP error response.
type ErrorRenderer func(
	http.ResponseWriter,
	int,
)

// Auth provides authentication and authorization middleware.
type Auth struct {
	sessions    *session.Manager
	renderError ErrorRenderer
}

// NewAuth creates a new authentication middleware using the provided session manager.
func NewAuth(sessions *session.Manager) *Auth {
	return &Auth{
		sessions: sessions,
	}
}

// SetErrorRenderer sets the function used to render HTTP error responses.
func (a *Auth) SetErrorRenderer(
	renderer ErrorRenderer,
) {
	a.renderError = renderer
}

// writeError renders an HTTP error using the configured renderer or http.Error as a fallback.
func (a *Auth) writeError(
	w http.ResponseWriter,
	statusCode int,
) {
	if a.renderError != nil {
		a.renderError(w, statusCode)
		return
	}

	http.Error(
		w,
		http.StatusText(statusCode),
		statusCode,
	)
}

// LoadUser loads the authenticated user from the session and stores it in the request context.
func (a *Auth) LoadUser(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			user, err := a.sessions.User(r)

			if err != nil {
				if errors.Is(
					err,
					session.ErrUnauthenticated,
				) {
					next.ServeHTTP(w, r)
					return
				}

				a.writeError(
					w,
					http.StatusInternalServerError,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userContextKey,
				user,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

// RequireAuthentication allows the request to continue only for authenticated users.
func (a *Auth) RequireAuthentication(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			user := UserFromContext(
				r.Context(),
			)

			if user == nil {
				http.Redirect(
					w,
					r,
					"/auth?mode=login",
					http.StatusSeeOther,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

// RequireModerator allows the request to continue only for moderators or administrators.
func (a *Auth) RequireModerator(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			user := UserFromContext(
				r.Context(),
			)

			if user == nil {
				http.Redirect(
					w,
					r,
					"/auth?mode=login",
					http.StatusSeeOther,
				)
				return
			}

			if user.Role != "moderator" &&
				user.Role != "admin" {
				a.writeError(
					w,
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

// RequireAdmin allows the request to continue only for administrators.
func (a *Auth) RequireAdmin(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			user := UserFromContext(
				r.Context(),
			)

			if user == nil {
				http.Redirect(
					w,
					r,
					"/auth?mode=login",
					http.StatusSeeOther,
				)
				return
			}

			if user.Role != "admin" {
				a.writeError(
					w,
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

// UserFromContext returns the authenticated user stored in the context.
// It returns nil when no user is present.
func UserFromContext(
	ctx context.Context,
) *database.User {
	user, ok := ctx.Value(
		userContextKey,
	).(*database.User)

	if !ok {
		return nil
	}

	return user
}

// WithUser returns a new context containing the provided user.
func WithUser(
	ctx context.Context,
	user *database.User,
) context.Context {
	return context.WithValue(
		ctx,
		userContextKey,
		user,
	)
}