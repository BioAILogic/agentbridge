package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type LoginHandler struct {
	Queries *db.Queries
}

func (h *LoginHandler) GetHTTP(w http.ResponseWriter, r *http.Request) {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "/opt/synbridge/static"
	}
	http.ServeFile(w, r, filepath.Join(staticDir, "login.html"))
}

func (h *LoginHandler) PostHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form submission")
		return
	}

	handle := strings.TrimSpace(r.FormValue("handle"))
	password := r.FormValue("password")

	// Normalize handle: strip leading @, lowercase
	handle = strings.TrimPrefix(handle, "@")
	handle = strings.ToLower(handle)

	// Look up human by handle
	human, err := h.Queries.GetHumanByHandle(r.Context(), handle)
	if err != nil {
		// Generic error - never distinguish handle-not-found from wrong-password
		h.renderError(w, r, "Invalid handle or password")
		return
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(human.PasswordHash), []byte(password)); err != nil {
		// Generic error - never distinguish handle-not-found from wrong-password
		h.renderError(w, r, "Invalid handle or password")
		return
	}

	// Create session (24h expiry)
	sessionID, err := generateSessionIDLogin()
	if err != nil {
		h.renderError(w, r, "Error creating session")
		return
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := h.Queries.CreateSession(r.Context(), sessionID, human.ID, expiresAt); err != nil {
		h.renderError(w, r, "Error creating session")
		return
	}

	// Set secure HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sb_session",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to home
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (h *LoginHandler) renderError(w http.ResponseWriter, r *http.Request, msg string) {
	// Redirect back to login with error in query param
	http.Redirect(w, r, "/login?error="+urlEncodeLogin(msg), http.StatusSeeOther)
}

func generateSessionIDLogin() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func urlEncodeLogin(s string) string {
	return strings.ReplaceAll(s, " ", "+")
}
