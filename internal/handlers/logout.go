package handlers

import (
	"net/http"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type LogoutHandler struct {
	Queries *db.Queries
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read session cookie
	cookie, err := r.Cookie("sb_session")
	if err == nil && cookie.Value != "" {
		// Delete session from DB
		_ = h.Queries.DeleteSession(r.Context(), cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sb_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
