package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

type HomeHandler struct {
	StaticDir string
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	index := filepath.Join(h.StaticDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		// Fallback if static dir not set up yet
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("SynBridge — coming soon"))
		return
	}
	http.ServeFile(w, r, index)
}
