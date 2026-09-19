package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"forum/internal/database"
)

type searchSuggestionPost struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

type searchSuggestionCategory struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Tagline     string `json:"tagline"`
}

type searchSuggestionsResponse struct {
	Posts      []searchSuggestionPost     `json:"posts"`
	Categories []searchSuggestionCategory `json:"categories"`
}

func (h *PageHandler) SearchSuggestions(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method Not Allowed",
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
			http.Error(
				w,
				"Internal Server Error",
				http.StatusInternalServerError,
			)
			return
		}

		categories, err := database.SearchCategories(
			h.db,
			query,
		)
		if err != nil {
			http.Error(
				w,
				"Internal Server Error",
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

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
