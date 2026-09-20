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

type CommentHandler struct {
	db *sql.DB
}

func NewCommentHandler(
	db *sql.DB,
) *CommentHandler {
	return &CommentHandler{
		db: db,
	}
}

func (h *CommentHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
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

	postID, err := commentPostIDFromPath(
		r.URL.Path,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusNotFound,
		)
		return
	}

	_, err = database.GetPostByID(
		h.db,
		postID,
	)
	if errors.Is(
		err,
		database.ErrPostNotFound,
	) {
		RenderErrorPage(
			w,
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	if err := r.ParseForm(); err != nil {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	content := strings.TrimSpace(
		r.FormValue("content"),
	)

	if content == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	commentID, err := database.CreateComment(
		h.db,
		postID,
		user.ID,
		content,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	err = database.CreateCommentNotification(
		h.db,
		postID,
		commentID,
		user.ID,
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
		"/posts/"+strconv.FormatInt(
			postID,
			10,
		),
		http.StatusSeeOther,
	)
}

func (h *CommentHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
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
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	commentID, err := strconv.ParseInt(
		r.FormValue("comment_id"),
		10,
		64,
	)
	if err != nil || commentID <= 0 {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	comment, err := database.GetCommentByID(
		h.db,
		commentID,
	)
	if errors.Is(
		err,
		database.ErrCommentNotFound,
	) {
		RenderErrorPage(
			w,
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	if comment.UserID != user.ID {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	content := strings.TrimSpace(
		r.FormValue("content"),
	)

	if content == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	if err := database.UpdateComment(
		h.db,
		commentID,
		content,
	); err != nil {
		if errors.Is(
			err,
			database.ErrCommentNotFound,
		) {
			RenderErrorPage(
				w,
				http.StatusNotFound,
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
		"/posts/"+strconv.FormatInt(
			comment.PostID,
			10,
		),
		http.StatusSeeOther,
	)
}

func (h *CommentHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
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
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	commentID, err := strconv.ParseInt(
		r.FormValue("comment_id"),
		10,
		64,
	)
	if err != nil || commentID <= 0 {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	comment, err := database.GetCommentByID(
		h.db,
		commentID,
	)
	if errors.Is(
		err,
		database.ErrCommentNotFound,
	) {
		RenderErrorPage(
			w,
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	canDelete :=
		comment.UserID == user.ID ||
			user.Role == "moderator" ||
			user.Role == "admin"

	if !canDelete {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	if err := database.DeleteComment(
		h.db,
		commentID,
	); err != nil {
		if errors.Is(
			err,
			database.ErrCommentNotFound,
		) {
			RenderErrorPage(
				w,
				http.StatusNotFound,
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
		"/posts/"+strconv.FormatInt(
			comment.PostID,
			10,
		),
		http.StatusSeeOther,
	)
}

func commentPostIDFromPath(
	path string,
) (int64, error) {
	const prefix = "/posts/"
	const suffix = "/comments"

	if !strings.HasPrefix(
		path,
		prefix,
	) || !strings.HasSuffix(
		path,
		suffix,
	) {
		return 0, errors.New(
			"invalid comment path",
		)
	}

	value := strings.TrimPrefix(
		path,
		prefix,
	)

	value = strings.TrimSuffix(
		value,
		suffix,
	)

	value = strings.Trim(
		value,
		"/",
	)

	if value == "" ||
		strings.Contains(value, "/") {
		return 0, errors.New(
			"invalid post id",
		)
	}

	postID, err := strconv.ParseInt(
		value,
		10,
		64,
	)
	if err != nil || postID <= 0 {
		return 0, errors.New(
			"invalid post id",
		)
	}

	return postID, nil
}
