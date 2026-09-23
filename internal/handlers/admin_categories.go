package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/database"
	"forum/internal/middleware"
)

// ManageCategories renders the category management page for administrators.
func (h *PageHandler) ManageCategories(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Accept only the exact administrator categories URL.
	if r.URL.Path != "/admin/categories" {
		h.renderNotFound(w)
		return
	}

	// The management page can only be viewed with a GET request.
	if r.Method != http.MethodGet {
		h.renderMethodNotAllowed(w)
		return
	}

	// Build the common page data, including the currently authenticated user.
	data := h.pageData(r)

	// Redirect unauthenticated users to the login page.
	if data.CurrentUser == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	// Only administrators are allowed to manage categories.
	if data.CurrentUser.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	// Load all categories from the database for the administration page.
	categories, err := database.GetAllCategories(h.db)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	// Store the categories in the page data used by the template.
	data.Categories = categories

	// Render the category administration page.
	h.render(
		w,
		http.StatusOK,
		"categories_admin.html",
		data,
	)
}

// CreateCategory handles administrator requests to create a new category.
func (h *ModerationHandler) CreateCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Category creation is only allowed through POST requests.
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Get the authenticated user that the middleware stored in the context.
	admin := middleware.UserFromContext(
		r.Context(),
	)

	// Redirect unauthenticated users to the login page.
	if admin == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	// Only administrators can create categories.
	if admin.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	// Parse the submitted category form.
	if err := r.ParseForm(); err != nil {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Read and clean the category values submitted by the administrator.
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

	// A category must contain both a name and a slug.
	if name == "" || slug == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Insert the new category into the database.
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

	// Return to the category management page after successful creation.
	http.Redirect(
		w,
		r,
		"/admin/categories",
		http.StatusSeeOther,
	)
}

// UpdateCategory handles administrator requests to update an existing category.
func (h *ModerationHandler) UpdateCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Category updates are only allowed through POST requests.
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Get the authenticated user from the request context.
	admin := middleware.UserFromContext(
		r.Context(),
	)

	// Redirect unauthenticated users to the login page.
	if admin == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	// Only administrators can update categories.
	if admin.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	// Parse the submitted category form.
	if err := r.ParseForm(); err != nil {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Convert the submitted category ID from text to an integer.
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

	// Read and clean the updated category values.
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

	// Reject updates that do not contain the required name and slug.
	if name == "" || slug == "" {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Update the selected category in the database.
	err = database.UpdateCategory(
		h.db,
		categoryID,
		name,
		slug,
		description,
		tagline,
	)
	if err != nil {
		// Return 404 when the requested category no longer exists.
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

		// Other update failures are treated as invalid requests.
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Return to the category management page after the update.
	http.Redirect(
		w,
		r,
		"/admin/categories",
		http.StatusSeeOther,
	)
}

// DeleteCategory handles administrator requests to delete an existing category.
func (h *ModerationHandler) DeleteCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Category deletion is only allowed through POST requests.
	if r.Method != http.MethodPost {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Get the authenticated user from the request context.
	admin := middleware.UserFromContext(
		r.Context(),
	)

	// Redirect unauthenticated users to the login page.
	if admin == nil {
		http.Redirect(
			w,
			r,
			"/auth?mode=login",
			http.StatusSeeOther,
		)
		return
	}

	// Only administrators can delete categories.
	if admin.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	// Parse the submitted deletion form.
	if err := r.ParseForm(); err != nil {
		RenderErrorPage(
			w,
			http.StatusBadRequest,
		)
		return
	}

	// Convert the submitted category ID from text to an integer.
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

	// Delete the selected category from the database.
	err = database.DeleteCategory(
		h.db,
		categoryID,
	)
	if err != nil {
		// Return 404 when the requested category does not exist.
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

		// Unexpected database failures are server errors.
		RenderErrorPage(
			w,
			http.StatusInternalServerError,
		)
		return
	}

	// Return to the category management page after successful deletion.
	http.Redirect(
		w,
		r,
		"/admin/categories",
		http.StatusSeeOther,
	)
}