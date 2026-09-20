package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/database"
	"forum/internal/middleware"
)

func (h *PageHandler) ManageCategories(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/admin/categories" {
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

	if data.CurrentUser.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	categories, err := database.GetAllCategories(h.db)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data.Categories = categories

	h.render(
		w,
		http.StatusOK,
		"categories_admin.html",
		data,
	)
}

func (h *ModerationHandler) CreateCategory(
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

	admin := middleware.UserFromContext(
		r.Context(),
	)
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
		RenderErrorPage(
			w,
			http.StatusForbidden,
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

	name := strings.TrimSpace(
		r.FormValue("name"),
	)
	slug := strings.TrimSpace(
		r.FormValue("slug"),
	)
	description := strings.TrimSpace(
		r.FormValue("description"),
	)
	tagline := strings.TrimSpace(
		r.FormValue("tagline"),
	)

	if name == "" || slug == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	_, err := database.CreateCategory(
		h.db,
		name,
		slug,
		description,
		tagline,
	)
	if err != nil {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/admin/categories",
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) UpdateCategory(
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

	admin := middleware.UserFromContext(
		r.Context(),
	)
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
		RenderErrorPage(
			w,
			http.StatusForbidden,
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

	categoryID, err := strconv.ParseInt(
		r.FormValue("category_id"),
		10,
		64,
	)
	if err != nil || categoryID <= 0 {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	name := strings.TrimSpace(
		r.FormValue("name"),
	)
	slug := strings.TrimSpace(
		r.FormValue("slug"),
	)
	description := strings.TrimSpace(
		r.FormValue("description"),
	)
	tagline := strings.TrimSpace(
		r.FormValue("tagline"),
	)

	if name == "" || slug == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	err = database.UpdateCategory(
		h.db,
		categoryID,
		name,
		slug,
		description,
		tagline,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrCategoryNotFound,
		) {
			RenderErrorPage(
				w,
				http.StatusNotFound,
			)
			return
		}

		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/admin/categories",
		http.StatusSeeOther,
	)
}

func (h *ModerationHandler) DeleteCategory(
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

	admin := middleware.UserFromContext(
		r.Context(),
	)
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
		RenderErrorPage(
			w,
			http.StatusForbidden,
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

	categoryID, err := strconv.ParseInt(
		r.FormValue("category_id"),
		10,
		64,
	)
	if err != nil || categoryID <= 0 {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	err = database.DeleteCategory(
		h.db,
		categoryID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrCategoryNotFound,
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
		"/admin/categories",
		http.StatusSeeOther,
	)
}
