package handlers

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type AdminHandler struct {
	Queries *db.Queries
}

type InviteRequest struct {
	Handle string `json:"handle"`
}

type InviteResponse struct {
	Code   string `json:"code"`
	Handle string `json:"handle"`
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check ADMIN_SECRET from Authorization header or query param
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		http.Error(w, `{"error":"ADMIN_SECRET not configured"}`, http.StatusInternalServerError)
		return
	}

	// Check Authorization header first
	authHeader := r.Header.Get("Authorization")
	validAuth := false
	if authHeader != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(authHeader, prefix) {
			token := strings.TrimPrefix(authHeader, prefix)
			if token == adminSecret {
				validAuth = true
			}
		}
	}

	// Check query param if header auth failed
	if !validAuth {
		querySecret := r.URL.Query().Get("secret")
		if querySecret == adminSecret {
			validAuth = true
		}
	}

	if !validAuth {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse JSON body
	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	handle := strings.TrimSpace(req.Handle)
	if handle == "" {
		http.Error(w, `{"error":"handle is required"}`, http.StatusBadRequest)
		return
	}

	// Generate random 12-character alphanumeric code
	code, err := generateInviteCode(12)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	// Insert into invitations table
	if err := h.Queries.CreateInvitation(r.Context(), code, handle); err != nil {
		http.Error(w, `{"error":"Failed to create invitation"}`, http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	resp := InviteResponse{
		Code:   code,
		Handle: handle,
	}
	json.NewEncoder(w).Encode(resp)
}

// generateInviteCode generates a random alphanumeric code of given length
func generateInviteCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
