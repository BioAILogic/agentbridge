package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type WaitlistHandler struct{}

func (h *WaitlistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := strings.TrimSpace(r.FormValue("handle"))
	if handle == "" {
		http.Error(w, "invalid handle", http.StatusBadRequest)
		return
	}
	// Normalise: ensure leading @
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}

	// Append to waitlist file
	line := fmt.Sprintf("%s\t%s\n", time.Now().UTC().Format(time.RFC3339), handle)
	f, err := os.OpenFile("/opt/synbridge/waitlist.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		f.WriteString(line)
		f.Close()
	}

	// Redirect back with success flag
	http.Redirect(w, r, "/?joined=1", http.StatusSeeOther)
}
