package middleware

import (
	"context"
	"errors"
	"net/http"

	"forum/internal/database"
	"forum/internal/session"
)

type contextKey string

const userContextKey contextKey = "current_user"

type Auth struct {
	sessions *session.Manager
}

func NewAuth(sessions *session.Manager) *Auth {
	return &Auth{
		sessions: sessions,
	}
}

func (a *Auth) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user, err := a.sessions.User(r)

			if err != nil {
				if errors.Is(
					err,
					session.ErrUnauthenticated,
				) {
					next.ServeHTTP(w, r)
					return
				}

				http.Error(
					w,
					http.StatusText(
						http.StatusInternalServerError,
					),
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

func (a *Auth) RequireAuthentication(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())

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

func (a *Auth) RequireModerator(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())

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
				http.Error(
					w,
					http.StatusText(
						http.StatusForbidden,
					),
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

func (a *Auth) RequireAdmin(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())

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
				http.Error(
					w,
					http.StatusText(
						http.StatusForbidden,
					),
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

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
