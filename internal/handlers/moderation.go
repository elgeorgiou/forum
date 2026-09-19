package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"forum/internal/database"
	"forum/internal/middleware"
)

type ModerationHandler struct {
	db *sql.DB
}

func NewModerationHandler(
	db *sql.DB,
) *ModerationHandler {
	return &ModerationHandler{
		db: db,
	}
}

func (h *ModerationHandler) RequestModerator(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	if user.Role != "user" {
		http.Error(
			w,
			"Forbidden",
			http.StatusForbidden,
		)
		return
	}

	_, err := database.CreateModeratorRequest(
		h.db,
		user.ID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrModeratorRequestExists,
		) {
			http.Redirect(
				w,
				r,
				"/moderation/request",
				http.StatusSeeOther,
			)
			return
		}

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/moderation/request",
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) ReviewModeratorRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	admin := middleware.UserFromContext(r.Context())
	if admin == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	if admin.Role != "admin" {
		http.Error(
			w,
			"Forbidden",
			http.StatusForbidden,
		)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)
		return
	}

	requestID, err := strconv.ParseInt(
		r.FormValue("request_id"),
		10,
		64,
	)
	if err != nil || requestID <= 0 {
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)
		return
	}

	var status string

	switch r.FormValue("action") {
	case "approve":
		status = "approved"

	case "reject":
		status = "rejected"

	default:
		http.Error(
			w,
			"Bad Request",
			http.StatusBadRequest,
		)
		return
	}

	err = database.ReviewModeratorRequest(
		h.db,
		requestID,
		status,
		admin.ID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrModeratorRequestNotFound,
		) {
			http.Error(
				w,
				"Not Found",
				http.StatusNotFound,
			)
			return
		}

		http.Error(
			w,
			"Internal Server Error",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/dashboard#moderator-requests",
		http.StatusSeeOther,
	)
}
