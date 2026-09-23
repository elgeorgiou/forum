package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"forum/internal/database"
)

// searchSuggestionPost represents the post data returned in search suggestions.
type searchSuggestionPost struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

// searchSuggestionCategory represents the category data returned in search suggestions.
type searchSuggestionCategory struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Tagline     string `json:"tagline"`
}

// searchSuggestionsResponse represents the complete search suggestions JSON response.
type searchSuggestionsResponse struct {
	Posts      []searchSuggestionPost     `json:"posts"`
	Categories []searchSuggestionCategory `json:"categories"`
}

// SearchSuggestions handles search suggestion requests for posts and categories.
func (h *PageHandler) SearchSuggestions(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		RenderErrorPage(
			w,
			http.StatusMethodNotAllowed,
		)
		return
	}

	query := strings.TrimSpace(
		r.URL.Query().Get("q"),
	)

	response := searchSuggestionsResponse{
		Posts:      []searchSuggestionPost{},
		Categories: []searchSuggestionCategory{},
	}

	if len(query) >= 2 {
		posts, err := database.SearchPostViews(
			h.db,
			query,
		)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusInternalServerError,
			)
			return
		}

		categories, err := database.SearchCategories(
			h.db,
			query,
		)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusInternalServerError,
			)
			return
		}

		for _, post := range posts {
			response.Posts = append(
				response.Posts,
				searchSuggestionPost{
					ID:       post.ID,
					Title:    post.Title,
					Username: post.Username,
				},
			)
		}

		for _, category := range categories {
			response.Categories = append(
				response.Categories,
				searchSuggestionCategory{
					Name:        category.Name,
					Slug:        category.Slug,
					Description: category.Description,
					Tagline:     category.Tagline,
				},
			)
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	_ = json.NewEncoder(w).Encode(response)
}