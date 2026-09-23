package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"forum/internal/database"
	"forum/internal/middleware"
)

// ReactionHandler handles HTTP requests for post and comment reactions.
type ReactionHandler struct {
	db *sql.DB
}

// reactionResponse represents the JSON response returned after a reaction update.
type reactionResponse struct {
	Reaction int `json:"reaction"`
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}

// NewReactionHandler creates a new ReactionHandler.
func NewReactionHandler(
	db *sql.DB,
) *ReactionHandler {
	return &ReactionHandler{
		db: db,
	}
}

// Post handles adding, changing, or removing a reaction on a post.
func (h *ReactionHandler) Post(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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

	if err := r.ParseMultipartForm(1 << 20); err != nil {
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

	reaction, err := parseReaction(
		r.FormValue("reaction"),
	)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
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
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	currentReaction, err :=
		database.GetPostReactionByUser(
			h.db,
			user.ID,
			postID,
		)

	switch {
	case err == nil &&
		currentReaction == reaction:

		err = database.RemovePostReaction(
			h.db,
			user.ID,
			postID,
		)

		reaction = 0

	case err == nil:

		err = database.SetPostReaction(
			h.db,
			user.ID,
			postID,
			reaction,
		)

	case errors.Is(
		err,
		database.ErrReactionNotFound,
	):

		err = database.SetPostReaction(
			h.db,
			user.ID,
			postID,
			reaction,
		)

	default:
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if reaction != 0 {
		notificationType :=
			database.NotificationPostLike

		if reaction == database.ReactionDislike {
			notificationType =
				database.NotificationPostDislike
		}

		err = database.CreatePostNotification(
			h.db,
			postID,
			user.ID,
			notificationType,
		)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusInternalServerError,
			)
			return
		}
	}

	if isFetchRequest(r) {
		likes, dislikes, err :=
			database.GetPostReactionCounts(
				h.db,
				postID,
			)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusInternalServerError,
			)
			return
		}

		writeReactionJSON(
			w,
			reaction,
			likes,
			dislikes,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/posts/"+
			strconv.FormatInt(postID, 10),
		http.StatusSeeOther,
	)
}

// Comment handles adding, changing, or removing a reaction on a comment.
func (h *ReactionHandler) Comment(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		RenderErrorPage(w, http.StatusMethodNotAllowed)
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

	if err := r.ParseMultipartForm(1 << 20); err != nil {
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

	reaction, err := parseReaction(
		r.FormValue("reaction"),
	)
	if err != nil {
		RenderErrorPage(w, http.StatusBadRequest)
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
		RenderErrorPage(w, http.StatusNotFound)
		return
	}

	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	currentReaction, err :=
		database.GetCommentReactionByUser(
			h.db,
			user.ID,
			commentID,
		)

	switch {
	case err == nil &&
		currentReaction == reaction:

		err = database.RemoveCommentReaction(
			h.db,
			user.ID,
			commentID,
		)

		reaction = 0

	case err == nil:

		err = database.SetCommentReaction(
			h.db,
			user.ID,
			commentID,
			reaction,
		)

	case errors.Is(
		err,
		database.ErrReactionNotFound,
	):

		err = database.SetCommentReaction(
			h.db,
			user.ID,
			commentID,
			reaction,
		)

	default:
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError)
		return
	}

	if isFetchRequest(r) {
		likes, dislikes, err :=
			database.GetCommentReactionCounts(
				h.db,
				commentID,
			)
		if err != nil {
			RenderErrorPage(
				w,
				http.StatusInternalServerError,
			)
			return
		}

		writeReactionJSON(
			w,
			reaction,
			likes,
			dislikes,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/posts/"+
			strconv.FormatInt(
				comment.PostID,
				10,
			),
		http.StatusSeeOther,
	)
}

// parseReaction converts and validates a reaction value.
func parseReaction(
	value string,
) (int, error) {
	reaction, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(
			"invalid reaction",
		)
	}

	if reaction != database.ReactionLike &&
		reaction != database.ReactionDislike {
		return 0, errors.New(
			"invalid reaction",
		)
	}

	return reaction, nil
}

// isFetchRequest reports whether the request was sent using the fetch API.
func isFetchRequest(
	r *http.Request,
) bool {
	return r.Header.Get(
		"X-Requested-With",
	) == "fetch"
}

// writeReactionJSON writes the updated reaction state as a JSON response.
func writeReactionJSON(
	w http.ResponseWriter,
	reaction int,
	likes int,
	dislikes int,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(
		reactionResponse{
			Reaction: reaction,
			Likes:    likes,
			Dislikes: dislikes,
		},
	)
}