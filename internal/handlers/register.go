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

type RegisterHandler struct {
	Queries *db.Queries
}

func (h *RegisterHandler) GetHTTP(w http.ResponseWriter, r *http.Request) {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "/opt/synbridge/static"
	}
	http.ServeFile(w, r, filepath.Join(staticDir, "register.html"))
}

func (h *RegisterHandler) PostHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form submission")
		return
	}

	handle := strings.TrimSpace(r.FormValue("handle"))
	inviteCode := strings.TrimSpace(r.FormValue("invite_code"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	// Validate handle
	if handle == "" {
		h.renderError(w, r, "Twitter handle is required")
		return
	}

	// Normalize handle: strip leading @, lowercase
	handle = strings.TrimPrefix(handle, "@")
	handle = strings.ToLower(handle)

	// Validate invite code
	if inviteCode == "" {
		h.renderError(w, r, "Invitation code is required")
		return
	}

	_, err := h.Queries.GetInvitation(r.Context(), inviteCode)
	if err != nil {
		h.renderError(w, r, "Invalid or already used invitation code")
		return
	}

	// Validate password
	if len(password) < 8 {
		h.renderError(w, r, "Password must be at least 8 characters")
		return
	}
	if password != passwordConfirm {
		h.renderError(w, r, "Passwords do not match")
		return
	}

	// Hash password with bcrypt (cost 12)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		h.renderError(w, r, "Error processing registration")
		return
	}

	// Insert human
	humanID, err := h.Queries.CreateHuman(r.Context(), handle, string(hash))
	if err != nil {
		h.renderError(w, r, "Error creating account (handle may already exist)")
		return
	}

	// Mark invitation used
	if err := h.Queries.MarkInvitationUsed(r.Context(), inviteCode, humanID); err != nil {
		h.renderError(w, r, "Error completing registration")
		return
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		h.renderError(w, r, "Error creating session")
		return
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := h.Queries.CreateSession(r.Context(), sessionID, humanID, expiresAt); err != nil {
		h.renderError(w, r, "Error creating session")
		return
	}

	// Set cookie
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

func (h *RegisterHandler) renderError(w http.ResponseWriter, r *http.Request, msg string) {
	// Redirect back to register with error in query param
	// The HTML page will read this and display the error
	http.Redirect(w, r, "/register?error="+urlEncode(msg), http.StatusSeeOther)
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func urlEncode(s string) string {
	// Simple URL encoding for error messages
	return strings.ReplaceAll(s, " ", "+")
}
