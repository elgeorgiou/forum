package handlers

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"forum/internal/database"
	"forum/internal/middleware"
	"forum/internal/models"
)

type PageHandler struct {
	db          *sql.DB
	templateDir string
}

func NewPageHandler(db *sql.DB) *PageHandler {
	return &PageHandler{
		db:          db,
		templateDir: "templates",
	}
}

func (h *PageHandler) Home(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/" {
		h.renderNotFound(w)
		return
	}

	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	categories, err := database.GetAllCategories(
		h.db,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data := h.pageData(r)

	filter := r.URL.Query().Get("filter")

	var posts []database.PostView

	switch filter {
	case "":
		posts, err = database.GetAllPostViews(
			h.db,
		)

	case "mine":
		if !data.IsAuthenticated {
			http.Redirect(
				w,
				r,
				"/auth?mode=login",
				http.StatusSeeOther,
			)
			return
		}

		posts, err = database.GetPostViewsByUser(
			h.db,
			data.CurrentUser.ID,
		)

	case "liked":
		if !data.IsAuthenticated {
			http.Redirect(
				w,
				r,
				"/auth?mode=login",
				http.StatusSeeOther,
			)
			return
		}

		posts, err = database.GetLikedPostViewsByUser(
			h.db,
			data.CurrentUser.ID,
		)

	default:
		h.renderBadRequest(w)
		return
	}

	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data.Categories = categories
	data.RecentPosts = posts
	data.PostFilter = filter

	h.render(
		w,
		http.StatusOK,
		"home.html",
		data,
	)
}

func (h *PageHandler) Auth(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	mode := r.URL.Query().Get("mode")

	if mode == "" {
		mode = "login"
	}

	if mode != "login" &&
		mode != "signup" {
		h.renderBadRequest(w)
		return
	}

	data := h.pageData(r)

	data.AuthMode = mode

	h.render(
		w,
		http.StatusOK,
		"auth.html",
		data,
	)
}

func (h *PageHandler) Categories(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/categories" {
		h.renderNotFound(w)
		return
	}

	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	categories, err := database.GetAllCategories(
		h.db,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data := h.pageData(r)

	data.Categories = categories

	h.render(
		w,
		http.StatusOK,
		"home.html",
		data,
	)
}

func (h *PageHandler) Category(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	slug := strings.TrimPrefix(
		r.URL.Path,
		"/categories/",
	)

	slug = strings.Trim(slug, "/")

	if slug == "" ||
		strings.Contains(slug, "/") {
		h.renderNotFound(w)
		return
	}

	category, err := database.GetCategoryBySlug(
		h.db,
		slug,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrCategoryNotFound,
		) {
			h.renderNotFound(w)
			return
		}

		h.renderInternalServerError(w)
		return
	}

	posts, err := database.GetPostViewsByCategory(
		h.db,
		category.ID,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data := h.pageData(r)

	data.Category = models.CategoryPageData{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Tagline:     category.Tagline,
		PostCount:   len(posts),
	}

	data.Posts = posts

	h.render(
		w,
		http.StatusOK,
		"category.html",
		data,
	)
}

func (h *PageHandler) Dashboard(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/dashboard" {
		h.renderNotFound(w)
		return
	}

	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	data := h.pageData(r)

	h.render(
		w,
		http.StatusOK,
		"dashboard.html",
		data,
	)
}

func (h *PageHandler) Activity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/activity" {
		h.renderNotFound(w)
		return
	}

	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	data := h.pageData(r)

	if data.CurrentUser == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	userID := data.CurrentUser.ID

	createdPosts, err := database.GetPostViewsByUser(
		h.db,
		userID,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	likedPosts, err := database.GetLikedPostViewsByUser(
		h.db,
		userID,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	dislikedPosts, err := database.GetDislikedPostViewsByUser(
		h.db,
		userID,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	comments, err := database.GetActivityCommentsByUser(
		h.db,
		userID,
	)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data.CreatedPosts = createdPosts
	data.LikedPosts = likedPosts
	data.DislikedPosts = dislikedPosts
	data.ActivityComments = comments

	h.render(
		w,
		http.StatusOK,
		"activity.html",
		data,
	)
}

func (h *PageHandler) pageData(
	r *http.Request,
) models.PageData {
	data := models.PageData{}

	user := middleware.UserFromContext(
		r.Context(),
	)

	if user == nil {
		return data
	}

	data.IsAuthenticated = true
	data.CurrentUser = user

	unreadCount, err :=
		database.GetUnreadNotificationCount(
			h.db,
			user.ID,
		)

	if err == nil {
		data.UnreadNotificationCount =
			unreadCount
	}

	return data
}

func (h *PageHandler) renderBadRequest(
	w http.ResponseWriter,
) {
	h.renderError(
		w,
		http.StatusBadRequest,
		"Oops! Bad Request.",
		"Something in this request doesn't look right.",
		"/",
		"Back to Home",
	)
}

func (h *PageHandler) renderNotFound(
	w http.ResponseWriter,
) {
	h.renderError(
		w,
		http.StatusNotFound,
		"Oops! Page Not Found.",
		"Looks like this page wandered off into another dimension.",
		"/",
		"Back to Home",
	)
}

func (h *PageHandler) renderMethodNotAllowed(
	w http.ResponseWriter,
) {
	h.renderError(
		w,
		http.StatusMethodNotAllowed,
		"Method Not Allowed.",
		"That action isn't available here.",
		"/",
		"Back to Home",
	)
}

func (h *PageHandler) renderInternalServerError(
	w http.ResponseWriter,
) {
	h.renderError(
		w,
		http.StatusInternalServerError,
		"Game Crash!",
		"Something went wrong on our side. Please try again.",
		"/",
		"Back to Home",
	)
}

func (h *PageHandler) renderError(
	w http.ResponseWriter,
	statusCode int,
	title string,
	message string,
	buttonURL string,
	buttonText string,
) {
	data := models.PageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
		ButtonURL:  buttonURL,
		ButtonText: buttonText,
	}

	h.render(
		w,
		statusCode,
		"error.html",
		data,
	)
}

func (h *PageHandler) render(
	w http.ResponseWriter,
	statusCode int,
	page string,
	data models.PageData,
) {
	files := []string{
		filepath.Join(
			h.templateDir,
			"layouts",
			"base.html",
		),
		filepath.Join(
			h.templateDir,
			"partials",
			"navbar.html",
		),
		filepath.Join(
			h.templateDir,
			"partials",
			"footer.html",
		),
		filepath.Join(
			h.templateDir,
			"pages",
			page,
		),
	}

	tmpl, err := template.ParseFiles(
		files...,
	)
	if err != nil {
		http.Error(
			w,
			"failed to load page",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(statusCode)

	if err := tmpl.ExecuteTemplate(
		w,
		"base",
		data,
	); err != nil {
		return
	}
}
