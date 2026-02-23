package handlers

import "net/http"

type FAQHandler struct {
	StaticDir string
}

func (h *FAQHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.StaticDir+"/faq.html")
}
