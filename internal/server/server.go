package server

import (
	"fmt"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
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

	pageHandler := handlers.NewPageHandler(db)

	mux := http.NewServeMux()

	staticFiles := http.FileServer(http.Dir("static"))

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", staticFiles),
	)

	mux.HandleFunc("/", pageHandler.Home)
	mux.HandleFunc("/auth", pageHandler.Auth)

	mux.HandleFunc("/categories", pageHandler.Categories)
	mux.HandleFunc("/categories/", pageHandler.Category)

	mux.HandleFunc("/dashboard", pageHandler.Dashboard)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server running at http://localhost:8080")

	return server.ListenAndServe()
}
