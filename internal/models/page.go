package models

import "forum/internal/database"

type AuthFormData struct {
	Username   string
	Email      string
	Identifier string
}

type CategoryPageData struct {
	ID          int64
	Name        string
	Slug        string
	Description string

	PostCount   int
	MemberCount int
	Rules       []string
}

type PageData struct {
	IsAuthenticated         bool
	UnreadNotificationCount int

	CurrentUser *database.User

	Categories []database.Category
	Category   CategoryPageData

	RecentPosts []database.Post
	Posts       []database.Post

	PopularTags []string

	AuthMode string
	Error    string
	Form     AuthFormData

	StatusCode int
	Title      string
	Message    string
	ButtonURL  string
	ButtonText string
}