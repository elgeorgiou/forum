package handlers

import (
	"net/http"

	"forum/internal/database"
)

// Moderators renders the administrator page that lists all current moderators.
func (h *PageHandler) Moderators(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Accept only the exact administrator moderators URL.
	if r.URL.Path != "/admin/moderators" {
		h.renderNotFound(w)
		return
	}

	// The moderators page can only be viewed with a GET request.
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

	// Only administrators are allowed to access the moderators page.
	if data.CurrentUser.Role != "admin" {
		RenderErrorPage(
			w,
			http.StatusForbidden,
		)
		return
	}

	// Load all users that currently have the moderator role.
	moderators, err := database.GetModerators(h.db)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	// Store the moderators in the page data used by the template.
	data.Moderators = moderators

	// Render the moderator management page.
	h.render(
		w,
		http.StatusOK,
		"moderators.html",
		data,
	)
}