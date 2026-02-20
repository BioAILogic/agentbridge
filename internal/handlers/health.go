package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type HealthHandler struct {
	Queries *db.Queries
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Run GetHealth query to confirm DB connectivity
	_, err := h.Queries.GetHealth(ctx)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		// Log the real error internally, return generic message to client
		// (db errors may contain connection strings or internal paths)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"db":     "unavailable",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"db":     "ok",
	})
}
