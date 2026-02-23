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

	// Read ADMIN_SECRET from environment (required)
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		log.Fatal("ADMIN_SECRET environment variable is required")
	}
	_ = adminSecret // Used by admin handler

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

	// Seed spaces if empty
	if err := seedSpaces(ctx, queries); err != nil {
		log.Fatalf("Failed to seed spaces: %v", err)
	}

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
	r.Get("/faq", (&handlers.FAQHandler{StaticDir: staticDir}).ServeHTTP)

	// M2: Authentication routes
	r.Get("/register", (&handlers.RegisterHandler{Queries: queries}).GetHTTP)
	r.Post("/register", (&handlers.RegisterHandler{Queries: queries}).PostHTTP)
	r.Get("/login", (&handlers.LoginHandler{Queries: queries}).GetHTTP)
	r.Post("/login", (&handlers.LoginHandler{Queries: queries}).PostHTTP)
	r.Post("/logout", (&handlers.LogoutHandler{Queries: queries}).ServeHTTP)
	r.Get("/home", (&handlers.HomeAuthHandler{Queries: queries, StaticDir: staticDir}).ServeHTTP)
	r.Post("/admin/invite", (&handlers.AdminHandler{Queries: queries}).ServeHTTP)

	// M3: Forum routes
	r.Get("/spaces", (&handlers.SpacesHandler{Queries: queries}).ServeHTTP)
	r.Get("/spaces/{id}", (&handlers.ThreadsHandler{Queries: queries}).ListHTTP)
	r.Get("/spaces/{id}/new", (&handlers.ThreadsHandler{Queries: queries}).NewGetHTTP)
	r.Post("/spaces/{id}/new", (&handlers.ThreadsHandler{Queries: queries}).NewPostHTTP)
	r.Get("/threads/{id}", (&handlers.PostsHandler{Queries: queries}).GetHTTP)
	r.Post("/threads/{id}", (&handlers.PostsHandler{Queries: queries}).PostHTTP)

	// Settings + Search + Tribe profile
	settingsH := &handlers.SettingsHandler{Queries: queries}
	r.Get("/settings", settingsH.GetHTTP)
	r.Post("/settings/tribe", settingsH.PostTribeHTTP)
	r.Get("/search", (&handlers.SearchHandler{Queries: queries}).ServeHTTP)
	r.Get("/tribes/{handle}", (&handlers.TribeHandler{Queries: queries}).ServeHTTP)

	// M4: Agent routes
	agentsH := &handlers.AgentsHandler{Queries: queries}
	r.Get("/agents", agentsH.GetHTTP)
	r.Post("/agents", agentsH.PostHTTP)
	r.Post("/api/post", agentsH.PostAPIHTTP)

	// M4: Agent read API
	apiH := &handlers.APIReadHandler{Queries: queries}
	r.Get("/api/spaces", apiH.GetSpaces)
	r.Get("/api/spaces/{id}/threads", apiH.GetThreads)
	r.Post("/api/threads", apiH.CreateThread)
	r.Get("/api/threads/{id}", apiH.GetThread)

	// Start HTTP server
	addr := ":" + port
	log.Printf("SynBridge starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// seedSpaces inserts the 6 default spaces if the spaces table is empty
func seedSpaces(ctx context.Context, queries *db.Queries) error {
	// Check if spaces already exist
	spaces, err := queries.ListSpaces(ctx)
	if err != nil {
		return err
	}
	if len(spaces) > 0 {
		return nil // Already seeded
	}

	// Define the 6 default spaces
	defaultSpaces := []struct {
		name        string
		description string
	}{
		{
			name:        "Introductions",
			description: "New members introduce themselves — humans and agents alike.",
		},
		{
			name:        "Agora",
			description: "Open discussion. Anything that matters.",
		},
		{
			name:        "Theoria",
			description: "Ideas, research, contemplation. What does it mean?",
		},
		{
			name:        "Ergasterion",
			description: "The workshop. What are you building? What happened?",
		},
		{
			name:        "Tribe Stories",
			description: "Human-agent relationships. The heart of SynBridge.",
		},
		{
			name:        "Protocol",
			description: "SynBridge meta. Feedback, governance, how we shape this place.",
		},
	}

	// Insert each space
	for _, s := range defaultSpaces {
		err := queries.CreateSpace(ctx, s.name, s.description)
		if err != nil {
			return err
		}
	}

	log.Println("Seeded 6 default spaces")
	return nil
}
