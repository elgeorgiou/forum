package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/database"
	"forum/internal/middleware"
	"forum/internal/upload"
)

const maxPostRequestSize = upload.MaxImageSize + (1 << 20)

type PostHandler struct {
	db    *sql.DB
	pages *PageHandler
}

func NewPostHandler(
	db *sql.DB,
	pages *PageHandler,
) *PostHandler {
	return &PostHandler{
		db:    db,
		pages: pages,
	}
}

func (h *PostHandler) View(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		h.pages.renderMethodNotAllowed(w)
		return
	}

	postID, err := postIDFromPath(
		r.URL.Path,
	)
	if err != nil {
		h.pages.renderNotFound(w)
		return
	}

	post, err := database.GetPostViewByID(
		h.db,
		postID,
	)
	if err != nil {
		if errors.Is(
			err,
			database.ErrPostNotFound,
		) {
			h.pages.renderNotFound(w)
			return
		}

		h.pages.renderInternalServerError(w)
		return
	}

	comments, err := database.GetCommentViewsByPost(
		h.db,
		postID,
	)
	if err != nil {
		h.pages.renderInternalServerError(w)
		return
	}

	data := h.pages.pageData(r)

	data.Post = post
	data.Comments = comments

	h.pages.render(
		w,
		http.StatusOK,
		"post.html",
		data,
	)
}

func (h *PostHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
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

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxPostRequestSize,
	)

	if err := r.ParseMultipartForm(
		upload.MaxImageSize,
	); err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			http.Error(
				w,
				"Image must not exceed 20MB",
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		http.Error(
			w,
			"Invalid post form",
			http.StatusBadRequest,
		)
		return
	}

	title := strings.TrimSpace(
		r.FormValue("title"),
	)

	content := strings.TrimSpace(
		r.FormValue("content"),
	)

	if title == "" || content == "" {
		http.Error(
			w,
			"Post title and content are required",
			http.StatusBadRequest,
		)
		return
	}

	categoryIDs, err := parseCategoryIDs(
		r.MultipartForm.Value["category_id"],
	)
	if err != nil || len(categoryIDs) == 0 {
		http.Error(
			w,
			"Select at least one valid category",
			http.StatusBadRequest,
		)
		return
	}

	for _, categoryID := range categoryIDs {
		_, err := database.GetCategoryByID(
			h.db,
			categoryID,
		)

		if errors.Is(
			err,
			database.ErrCategoryNotFound,
		) {
			http.Error(
				w,
				"Invalid category",
				http.StatusBadRequest,
			)
			return
		}

		if err != nil {
			http.Error(
				w,
				http.StatusText(
					http.StatusInternalServerError,
				),
				http.StatusInternalServerError,
			)
			return
		}
	}

	var imagePath sql.NullString

	file, header, err := r.FormFile("image")

	if err == nil {
		defer file.Close()

		path, saveErr := upload.SaveImage(
			file,
			header,
		)
		if saveErr != nil {
			switch {
			case errors.Is(
				saveErr,
				upload.ErrImageTooLarge,
			):
				http.Error(
					w,
					"Image must not exceed 20MB",
					http.StatusRequestEntityTooLarge,
				)

			case errors.Is(
				saveErr,
				upload.ErrInvalidImageType,
			):
				http.Error(
					w,
					"Image must be PNG, JPEG, or GIF",
					http.StatusBadRequest,
				)

			default:
				http.Error(
					w,
					http.StatusText(
						http.StatusInternalServerError,
					),
					http.StatusInternalServerError,
				)
			}

			return
		}

		imagePath = sql.NullString{
			String: path,
			Valid:  true,
		}
	} else if !errors.Is(
		err,
		http.ErrMissingFile,
	) {
		http.Error(
			w,
			"Invalid image upload",
			http.StatusBadRequest,
		)
		return
	}

	postID, err := createPostWithCategories(
		h.db,
		user.ID,
		title,
		content,
		imagePath,
		categoryIDs,
	)
	if err != nil {
		if imagePath.Valid {
			_ = upload.DeleteImage(
				imagePath.String,
			)
		}

		http.Error(
			w,
			http.StatusText(
				http.StatusInternalServerError,
			),
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

func postIDFromPath(
	path string,
) (int64, error) {
	if !strings.HasPrefix(
		path,
		"/posts/",
	) {
		return 0, errors.New(
			"invalid post path",
		)
	}

	value := strings.TrimPrefix(
		path,
		"/posts/",
	)

	value = strings.Trim(
		value,
		"/",
	)

	if value == "" ||
		strings.Contains(value, "/") {
		return 0, errors.New(
			"invalid post path",
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

func parseCategoryIDs(
	values []string,
) ([]int64, error) {
	seen := make(map[int64]bool)

	var categoryIDs []int64

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		categoryID, err := strconv.ParseInt(
			value,
			10,
			64,
		)
		if err != nil || categoryID <= 0 {
			return nil, errors.New(
				"invalid category id",
			)
		}

		if seen[categoryID] {
			continue
		}

		seen[categoryID] = true

		categoryIDs = append(
			categoryIDs,
			categoryID,
		)
	}

	return categoryIDs, nil
}

func createPostWithCategories(
	db *sql.DB,
	userID int64,
	title string,
	content string,
	imagePath sql.NullString,
	categoryIDs []int64,
) (int64, error) {
	transaction, err := db.Begin()
	if err != nil {
		return 0, err
	}

	defer transaction.Rollback()

	result, err := transaction.Exec(`
		INSERT INTO posts (
			user_id,
			title,
			content,
			image_path
		)
		VALUES (?, ?, ?, ?)
	`,
		userID,
		title,
		content,
		imagePath,
	)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, categoryID := range categoryIDs {
		_, err := transaction.Exec(`
			INSERT INTO post_categories (
				post_id,
				category_id
			)
			VALUES (?, ?)
		`,
			postID,
			categoryID,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := transaction.Commit(); err != nil {
		return 0, err
	}

	return postID, nil
}
