package server

import (
	"fmt"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"forum/internal/session"
)

func Run() error {
	db, err := database.Open("data/forum.db")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.InitSchema(
		db,
		"internal/database/schema/schema.sql",
	); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	if err := database.SeedCategories(db); err != nil {
		return fmt.Errorf("seed categories: %w", err)
	}

	sessionManager := session.NewManager(db)

	if err := sessionManager.DeleteExpired(); err != nil {
		return fmt.Errorf(
			"delete expired sessions: %w",
			err,
		)
	}

	authMiddleware := middleware.NewAuth(sessionManager)
	pageHandler := handlers.NewPageHandler(db)
	authHandler := handlers.NewAuthHandler(
		db,
		sessionManager,
		pageHandler,
	)
	postHandler := handlers.NewPostHandler(
		db,
		pageHandler,
	)
	commentHandler := handlers.NewCommentHandler(db)
	reactionHandler := handlers.NewReactionHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	moderationHandler := handlers.NewModerationHandler(db)
	githubOAuthHandler := handlers.NewGitHubOAuthHandler(
		db,
		sessionManager,
	)
	googleOAuthHandler := handlers.NewGoogleOAuthHandler(
		db,
		sessionManager,
	)

	mux := http.NewServeMux()

	staticFiles := http.FileServer(
		http.Dir("static"),
	)

	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			staticFiles,
		),
	)

	mux.HandleFunc("/", pageHandler.Home)
	mux.HandleFunc("/auth", pageHandler.Auth)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	mux.HandleFunc(
		"/auth/github",
		githubOAuthHandler.Login,
	)
	mux.HandleFunc(
		"/auth/github/callback",
		githubOAuthHandler.Callback,
	)
	mux.HandleFunc(
		"/auth/google",
		googleOAuthHandler.Login,
	)
	mux.HandleFunc(
		"/auth/google/callback",
		googleOAuthHandler.Callback,
	)

	mux.HandleFunc(
		"/categories",
		pageHandler.Categories,
	)
	mux.HandleFunc(
		"/recent",
		pageHandler.Recent,
	)
	mux.Handle(
		"/profile",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				pageHandler.Profile,
			),
		),
	)
	mux.HandleFunc(
		"/about",
		pageHandler.About,
	)
	mux.HandleFunc(
		"/search",
		pageHandler.Search,
	)
	mux.HandleFunc(
		"/search/suggestions",
		pageHandler.SearchSuggestions,
	)
	mux.HandleFunc(
		"/categories/",
		pageHandler.Category,
	)

	mux.Handle(
		"/posts/create",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				postHandler.Create,
			),
		),
	)
	mux.Handle(
		"/posts/update",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				postHandler.Update,
			),
		),
	)
	mux.Handle(
		"/posts/delete",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				postHandler.Delete,
			),
		),
	)
	mux.Handle(
		"/posts/react",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				reactionHandler.Post,
			),
		),
	)
	mux.Handle(
		"/comments/update",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				commentHandler.Update,
			),
		),
	)
	mux.Handle(
		"/comments/delete",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				commentHandler.Delete,
			),
		),
	)
	mux.Handle(
		"/comments/react",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				reactionHandler.Comment,
			),
		),
	)
	mux.Handle(
		"/posts/{id}/comments",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				commentHandler.Create,
			),
		),
	)

	mux.Handle(
		"POST /moderation/reports/posts",
		authMiddleware.RequireModerator(
			http.HandlerFunc(
				moderationHandler.ReportPost,
			),
		),
	)
	mux.Handle(
		"POST /moderation/reports/comments",
		authMiddleware.RequireModerator(
			http.HandlerFunc(
				moderationHandler.ReportComment,
			),
		),
	)

	mux.HandleFunc(
		"/posts/",
		postHandler.View,
	)

	mux.Handle(
		"/dashboard",
		authMiddleware.RequireModerator(
			http.HandlerFunc(
				pageHandler.Dashboard,
			),
		),
	)
	mux.Handle(
		"/activity",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				pageHandler.Activity,
			),
		),
	)
	mux.Handle(
		"/notifications",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				pageHandler.Notifications,
			),
		),
	)
	mux.Handle(
		"/notifications/read",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				notificationHandler.Read,
			),
		),
	)
	mux.Handle(
		"/notifications/read-all",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				notificationHandler.ReadAll,
			),
		),
	)

	mux.Handle(
		"GET /moderation/request",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				pageHandler.ModeratorRequest,
			),
		),
	)
	mux.Handle(
		"POST /moderation/request",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				moderationHandler.RequestModerator,
			),
		),
	)
	mux.Handle(
		"POST /admin/moderation/requests/review",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.ReviewModeratorRequest,
			),
		),
	)
	mux.Handle(
		"POST /admin/moderation/reports/review",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.ReviewReport,
			),
		),
	)

	mux.Handle(
		"GET /admin/moderators",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				pageHandler.Moderators,
			),
		),
	)
	mux.Handle(
		"POST /admin/moderators/demote",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.DemoteModerator,
			),
		),
	)

	mux.Handle(
		"GET /admin/categories",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				pageHandler.ManageCategories,
			),
		),
	)
	mux.Handle(
		"POST /admin/categories/create",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.CreateCategory,
			),
		),
	)
	mux.Handle(
		"POST /admin/categories/update",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.UpdateCategory,
			),
		),
	)
	mux.Handle(
		"POST /admin/categories/delete",
		authMiddleware.RequireAdmin(
			http.HandlerFunc(
				moderationHandler.DeleteCategory,
			),
		),
	)

	handler := authMiddleware.LoadUser(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println(
		"Server running at http://localhost:8080",
	)

	return server.ListenAndServe()
}
