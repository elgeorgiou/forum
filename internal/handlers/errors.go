package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"forum/internal/models"
)

func RenderErrorPage(
	w http.ResponseWriter,
	statusCode int,
) {
	title := ""
	message := ""
	buttonURL := "/"
	buttonText := "Back to Home"

	switch statusCode {
	case http.StatusBadRequest:
		title = "Oops! Bad Request."
		message = "Something in this request doesn't look right."

	case http.StatusUnauthorized:
		title = "Login Required."
		message = "You need to log in before you can continue."
		buttonURL = "/auth?mode=login"
		buttonText = "Log In"

	case http.StatusForbidden:
		title = "Access Denied."
		message = "You don't have permission to enter this part of the forum."

	case http.StatusNotFound:
		title = "Oops! Page Not Found."
		message = "Looks like this page wandered off into another dimension."

	case http.StatusMethodNotAllowed:
		title = "Method Not Allowed."
		message = "That action isn't available here."

	case http.StatusRequestEntityTooLarge:
		title = "Upload Too Large."
		message = "Your image exceeds the 20 MB upload limit."
		buttonURL = "/"
		buttonText = "Back to Home"

	case http.StatusTooManyRequests:
		title = "Slow Down, Player!"
		message = "Too many requests. Give it a moment and try again."

	default:
		statusCode = http.StatusInternalServerError
		title = "Game Crash!"
		message = "Something went wrong on our side. Please try again."
	}

	data := models.PageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
		ButtonURL:  buttonURL,
		ButtonText: buttonText,
	}

	files := []string{
		filepath.Join(
			"templates",
			"layouts",
			"base.html",
		),
		filepath.Join(
			"templates",
			"partials",
			"navbar.html",
		),
		filepath.Join(
			"templates",
			"partials",
			"footer.html",
		),
		filepath.Join(
			"templates",
			"pages",
			"error.html",
		),
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(statusCode)

	_ = tmpl.ExecuteTemplate(
		w,
		"base",
		data,
	)
}
