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
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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
		RenderErrorPage(w, http.StatusForbidden)
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

		RenderErrorPage(
			w,
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
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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
		RenderErrorPage(w, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	requestID, err := strconv.ParseInt(
		r.FormValue("request_id"),
		10,
		64,
	)
	if err != nil || requestID <= 0 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	var status string

	switch r.FormValue("action") {
	case "approve":
		status = "approved"

	case "reject":
		status = "rejected"

	default:
		RenderErrorPage(w, http.StatusBadRequest)
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
			RenderErrorPage(w, http.StatusNotFound)
			return
		}

		RenderErrorPage(
			w,
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

func (h *ModerationHandler) ReportPost(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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

	if user.Role != "moderator" &&
		user.Role != "admin" {
		RenderErrorPage(w, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(
		r.FormValue("post_id"),
		10,
		64,
	)
	if err != nil || postID <= 0 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	reason := strings.TrimSpace(
		r.FormValue("reason"),
	)

	if reason == "" {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	if len([]rune(reason)) > 1000 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	_, err = database.GetPostByID(
		h.db,
		postID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrPostNotFound,
		) {
			RenderErrorPage(w, http.StatusNotFound)
			return
		}

		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	_, err = database.CreatePostReport(
		h.db,
		user.ID,
		postID,
		reason,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/posts/"+strconv.FormatInt(postID, 10),
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) ReportComment(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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

	if user.Role != "moderator" &&
		user.Role != "admin" {
		RenderErrorPage(w, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(
		r.FormValue("comment_id"),
		10,
		64,
	)
	if err != nil || commentID <= 0 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	reason := strings.TrimSpace(
		r.FormValue("reason"),
	)

	if reason == "" {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	if len([]rune(reason)) > 1000 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	comment, err := database.GetCommentByID(
		h.db,
		commentID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrCommentNotFound,
		) {
			RenderErrorPage(w, http.StatusNotFound)
			return
		}

		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	_, err = database.CreateCommentReport(
		h.db,
		user.ID,
		commentID,
		reason,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/posts/"+
			strconv.FormatInt(comment.PostID, 10)+
			"#comment-"+
			strconv.FormatInt(commentID, 10),
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) ReviewReport(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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
		RenderErrorPage(w, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	reportID, err := strconv.ParseInt(
		r.FormValue("report_id"),
		10,
		64,
	)
	if err != nil || reportID <= 0 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	response := strings.TrimSpace(
		r.FormValue("response"),
	)

	if response == "" {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	if len([]rune(response)) > 1000 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	var status string

	switch r.FormValue("action") {
	case "resolve":
		status = "resolved"

	case "reject":
		status = "rejected"

	default:
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	err = database.ReviewReport(
		h.db,
		reportID,
		status,
		admin.ID,
		response,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrReportNotFound,
		) {
			RenderErrorPage(w, http.StatusNotFound)
			return
		}

		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/dashboard#reports",
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) DemoteModerator(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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
		RenderErrorPage(w, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(
		r.FormValue("user_id"),
		10,
		64,
	)
	if err != nil || userID <= 0 {
		RenderErrorPage(w, http.StatusBadRequest)
		return
	}

	err = database.DemoteModerator(
		h.db,
		userID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrModeratorNotFound,
		) {
			RenderErrorPage(w, http.StatusNotFound)
			return
		}

		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/admin/moderators",
		http.StatusSeeOther,
	)
}
