package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type HomeAuthHandler struct {
	Queries   *db.Queries
	StaticDir string
}

func (h *HomeAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read sb_session cookie
	cookie, err := r.Cookie("sb_session")
	if err != nil {
		// No cookie, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Validate session exists and not expired
	_, err = h.Queries.GetSession(r.Context(), cookie.Value)
	if err != nil {
		// Invalid or expired session, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Valid session, serve the home page
	staticDir := h.StaticDir
	if staticDir == "" {
		staticDir = os.Getenv("STATIC_DIR")
		if staticDir == "" {
			staticDir = "/opt/synbridge/static"
		}
	}
	http.ServeFile(w, r, filepath.Join(staticDir, "home.html"))
}
