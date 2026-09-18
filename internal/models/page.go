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
	Tagline     string

	PostCount   int
	MemberCount int
	Rules       []string
}

type CommunityStats struct {
	MemberCount  int
	PostCount    int
	CommentCount int
}

type PageData struct {
	IsAuthenticated         bool
	UnreadNotificationCount int

	CurrentUser *database.User

	Categories []database.Category
	Category   CategoryPageData

	RecentPosts []database.PostView
	Posts       []database.PostView
	Post        database.PostView
	Comments    []database.CommentView

	CommunityStats CommunityStats

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
