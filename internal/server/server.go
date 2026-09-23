package server

import (
	"fmt"
	"net/http"
	"os"

	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"forum/internal/session"
)

// Run initializes the application dependencies, registers the routes,
// and starts the HTTP server.
func Run() error {
	db, err := database.Open("data/forum.db")
	if err != nil {
		return fmt.Errorf(
			"open database: %w",
			err,
		)
	}
	defer db.Close()

	// Initialize the database schema and apply required migrations.
	if err := database.InitSchema(
		db,
		"internal/database/schema/schema.sql",
	); err != nil {
		return fmt.Errorf(
			"initialize database: %w",
			err,
		)
	}

	// Insert the default forum categories if they do not already exist.
	if err := database.SeedCategories(db); err != nil {
		return fmt.Errorf(
			"seed categories: %w",
			err,
		)
	}

	// Create the session manager and remove sessions that have expired.
	sessionManager := session.NewManager(db)

	if err := sessionManager.DeleteExpired(); err != nil {
		return fmt.Errorf(
			"delete expired sessions: %w",
			err,
		)
	}

	// Configure authentication and authorization middleware.
	authMiddleware := middleware.NewAuth(
		sessionManager,
	)

	authMiddleware.SetErrorRenderer(
		handlers.RenderErrorPage,
	)

	// Create the application's HTTP handlers and inject their dependencies.
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

	commentHandler :=
		handlers.NewCommentHandler(db)

	reactionHandler :=
		handlers.NewReactionHandler(db)

	notificationHandler :=
		handlers.NewNotificationHandler(db)

	moderationHandler :=
		handlers.NewModerationHandler(db)

	githubOAuthHandler :=
		handlers.NewGitHubOAuthHandler(
			db,
			sessionManager,
		)

	googleOAuthHandler :=
		handlers.NewGoogleOAuthHandler(
			db,
			sessionManager,
		)

	// Create the HTTP request multiplexer used to register application routes.
	mux := http.NewServeMux()

	// Serve static assets such as CSS, JavaScript, and uploaded images.
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

	// Register public page and authentication routes.
	mux.HandleFunc(
		"/",
		pageHandler.Home,
	)

	mux.HandleFunc(
		"/auth",
		pageHandler.Auth,
	)

	mux.HandleFunc(
		"/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"/logout",
		authHandler.Logout,
	)

	// Register OAuth authentication routes.
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

	// Register public forum browsing routes.
	mux.HandleFunc(
		"/categories",
		pageHandler.Categories,
	)

	mux.HandleFunc(
		"/recent",
		pageHandler.Recent,
	)

	// Require authentication before allowing access to the profile page.
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

	// Register search routes.
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

	// Register authenticated post management routes.
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

	// Register authenticated comment management routes.
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

	// Register moderation report routes available to moderators and admins.
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

	// Register individual post pages.
	mux.HandleFunc(
		"/posts/",
		postHandler.View,
	)

	// Protect the moderation dashboard so only moderators and admins can access it.
	mux.Handle(
		"/dashboard",
		authMiddleware.RequireModerator(
			http.HandlerFunc(
				pageHandler.Dashboard,
			),
		),
	)

	// Register authenticated user activity and notification routes.
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

	// Register moderator application routes for authenticated users.
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

	// Register administrator moderation management routes.
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

	// Register administrator moderator management routes.
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

	// Register administrator category management routes.
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

	// Load the authenticated user into the request context before routing requests.
	handler := authMiddleware.LoadUser(mux)

	// Use the configured port or fall back to the default development port.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure the HTTP server with the application's middleware-wrapped handler.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	fmt.Printf(
		"Server running on port %s\n",
		port,
	)

	// Start accepting and serving HTTP requests.
	return server.ListenAndServe()
}
