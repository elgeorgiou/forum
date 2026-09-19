package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/database"
	"forum/internal/middleware"
)

type NotificationHandler struct {
	db *sql.DB
}

func NewNotificationHandler(
	db *sql.DB,
) *NotificationHandler {
	return &NotificationHandler{
		db: db,
	}
}

func (h *NotificationHandler) Read(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	user := middleware.UserFromContext(
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

	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			"Invalid notification",
			http.StatusBadRequest,
		)
		return
	}

	notificationID, err := strconv.ParseInt(
		r.FormValue("notification_id"),
		10,
		64,
	)
	if err != nil || notificationID <= 0 {
		http.Error(
			w,
			"Invalid notification",
			http.StatusBadRequest,
		)
		return
	}

	err = database.MarkNotificationAsRead(
		h.db,
		notificationID,
		user.ID,
	)
	if errors.Is(
		err,
		database.ErrNotificationNotFound,
	) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		http.Error(
			w,
			http.StatusText(
				http.StatusInternalServerError,
			),
			http.StatusInternalServerError,
		)
		return
	}

	redirectURL := strings.TrimSpace(
		r.FormValue("redirect"),
	)

	if !strings.HasPrefix(
		redirectURL,
		"/posts/",
	) {
		redirectURL = "/notifications"
	}

	http.Redirect(
		w,
		r,
		redirectURL,
		http.StatusSeeOther,
	)
}

func (h *NotificationHandler) ReadAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	user := middleware.UserFromContext(
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

	err := database.MarkAllNotificationsAsRead(
		h.db,
		user.ID,
	)
	if err != nil {
		http.Error(
			w,
			http.StatusText(
				http.StatusInternalServerError,
			),
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/notifications",
		http.StatusSeeOther,
	)
}
