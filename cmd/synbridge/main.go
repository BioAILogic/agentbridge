package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/BioAILogic/agentbridge/internal/db"
	"github.com/BioAILogic/agentbridge/internal/handlers"
)

func main() {
	// Read DATABASE_URL from environment (required)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Read PORT from environment (default: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to PostgreSQL via pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create sqlc queries
	queries := db.New(pool)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Static files (landing page, assets)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "/opt/synbridge/static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/assets/*", fs)

	// Routes
	r.Get("/health", (&handlers.HealthHandler{Queries: queries}).ServeHTTP)
	r.Get("/", (&handlers.HomeHandler{StaticDir: staticDir}).ServeHTTP)
	r.Post("/waitlist", (&handlers.WaitlistHandler{}).ServeHTTP)

	// Start HTTP server
	addr := ":" + port
	log.Printf("SynBridge starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
