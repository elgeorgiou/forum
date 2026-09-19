package handlers

import (
	"net/http"

	"forum/internal/database"
)

func (h *PageHandler) Moderators(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Path != "/admin/moderators" {
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
		http.Error(
			w,
			"Forbidden",
			http.StatusForbidden,
		)
		return
	}

	moderators, err := database.GetModerators(h.db)
	if err != nil {
		h.renderInternalServerError(w)
		return
	}

	data.Moderators = moderators

	h.render(
		w,
		http.StatusOK,
		"moderators.html",
		data,
	)
}
