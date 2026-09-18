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
		return fmt.Errorf("delete expired sessions: %w", err)
	}

	authMiddleware := middleware.NewAuth(
		sessionManager,
	)

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

	commentHandler := handlers.NewCommentHandler(
		db,
	)

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
		"/posts/{id}/comments",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				commentHandler.Create,
			),
		),
	)

	mux.HandleFunc(
		"/posts/",
		postHandler.View,
	)

	mux.Handle(
		"/dashboard",
		authMiddleware.RequireAuthentication(
			http.HandlerFunc(
				pageHandler.Dashboard,
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
