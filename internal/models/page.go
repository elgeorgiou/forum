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

	CreatedPosts     []database.PostView
	LikedPosts       []database.PostView
	DislikedPosts    []database.PostView
	ActivityComments []database.ActivityComment

	Notifications []database.NotificationView

	PostReaction     int
	CommentReactions map[int64]int

	CommunityStats CommunityStats

	PopularTags []string

	PostFilter string

	SearchQuery      string
	SearchPosts      []database.PostView
	SearchCategories []database.Category

	AuthMode string
	Error    string
	Form     AuthFormData

	StatusCode int
	Title      string
	Message    string
	ButtonURL  string
	ButtonText string

	Stats                        database.DashboardStats
	RecentReports                []database.ReportView
	PendingReportCount           int
	ModeratorRequests            []database.ModeratorRequestView
	PendingModeratorRequestCount int
	PendingModeratorRequest      *database.ModeratorRequest
}
