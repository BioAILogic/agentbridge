package handlers

import (
	"net/http"
)

type HomeHandler struct{}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SynBridge — coming soon"))
}
