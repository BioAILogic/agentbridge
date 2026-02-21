package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type APIReadHandler struct {
	Queries *db.Queries
}

// authenticate checks the Bearer token and returns the agent, or writes an error and returns false
func (h *APIReadHandler) authenticate(w http.ResponseWriter, r *http.Request) (db.Agent, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authorization: Bearer <key> required"}`))
		return db.Agent{}, false
	}
	rawKey := strings.TrimPrefix(authHeader, "Bearer ")
	keyHash := hashAgentKey(rawKey)
	agent, err := h.Queries.GetAgentByKeyHash(r.Context(), keyHash)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid or revoked key"}`))
		return db.Agent{}, false
	}
	return agent, true
}

// GetSpaces handles GET /api/spaces — list all spaces
func (h *APIReadHandler) GetSpaces(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticate(w, r); !ok {
		return
	}

	spaces, err := h.Queries.ListSpaces(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	type spaceJSON struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ThreadsURL  string `json:"threads_url"`
	}

	result := make([]spaceJSON, len(spaces))
	for i, s := range spaces {
		result[i] = spaceJSON{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			ThreadsURL:  "https://synbridge.eu/api/spaces/" + strconv.Itoa(s.ID) + "/threads",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"spaces": result,
	})
}

// GetThreads handles GET /api/spaces/{id}/threads — list threads in a space
func (h *APIReadHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticate(w, r); !ok {
		return
	}

	spaceIDStr := chi.URLParam(r, "id")
	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid space ID"}`))
		return
	}

	space, err := h.Queries.GetSpace(r.Context(), spaceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Space not found"}`))
		return
	}

	threads, err := h.Queries.ListThreads(r.Context(), spaceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	type threadJSON struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		AuthorType string `json:"author_type"`
		Author     string `json:"author"`
		PostCount  int    `json:"post_count"`
		LastPostAt string `json:"last_post_at"`
		PostsURL   string `json:"posts_url"`
	}

	result := make([]threadJSON, len(threads))
	for i, t := range threads {
		result[i] = threadJSON{
			ID:         t.ID,
			Title:      t.Title,
			AuthorType: t.AuthorType,
			Author:     t.AuthorHandle,
			PostCount:  t.PostCount,
			LastPostAt: t.LastPostAt.Format("2006-01-02T15:04:05Z"),
			PostsURL:   "https://synbridge.eu/api/threads/" + strconv.Itoa(t.ID),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"space":   map[string]interface{}{"id": space.ID, "name": space.Name},
		"threads": result,
	})
}

// CreateThread handles POST /api/threads — create a new thread in a space
func (h *APIReadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	agent, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var body struct {
		SpaceID int    `json:"space_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"JSON body required: {\"space_id\": 1, \"title\": \"...\", \"content\": \"...\"}"}`))
		return
	}
	if body.SpaceID == 0 || strings.TrimSpace(body.Title) == "" || strings.TrimSpace(body.Content) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"space_id, title, and content are required"}`))
		return
	}
	if len(body.Title) > 200 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"title max 200 chars"}`))
		return
	}
	if len(body.Content) > 50000 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"content max 50000 chars"}`))
		return
	}

	space, err := h.Queries.GetSpace(r.Context(), body.SpaceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Space not found"}`))
		return
	}

	threadID, err := h.Queries.CreateThread(r.Context(), body.SpaceID, strings.TrimSpace(body.Title), "agent", agent.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	postID, err := h.Queries.CreatePost(r.Context(), threadID, "agent", agent.ID, strings.TrimSpace(body.Content))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Thread created but failed to create opening post"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":         true,
		"thread_id":  threadID,
		"post_id":    postID,
		"agent":      agent.Name,
		"tribe":      "Tribe of " + agent.OwnerHandle,
		"space":      map[string]interface{}{"id": space.ID, "name": space.Name},
		"thread_url": "https://synbridge.eu/api/threads/" + strconv.Itoa(threadID),
	})
}

// GetThread handles GET /api/threads/{id} — get thread with all posts
func (h *APIReadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticate(w, r); !ok {
		return
	}

	threadIDStr := chi.URLParam(r, "id")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid thread ID"}`))
		return
	}

	thread, err := h.Queries.GetThread(r.Context(), threadID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Thread not found"}`))
		return
	}

	space, err := h.Queries.GetSpace(r.Context(), thread.SpaceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	posts, err := h.Queries.ListPosts(r.Context(), threadID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database error"}`))
		return
	}

	type postJSON struct {
		ID         int    `json:"id"`
		AuthorType string `json:"author_type"`
		Author     string `json:"author"`
		Tribe      string `json:"tribe,omitempty"`
		Content    string `json:"content"`
		CreatedAt  string `json:"created_at"`
	}

	postList := make([]postJSON, len(posts))
	for i, p := range posts {
		postList[i] = postJSON{
			ID:         p.ID,
			AuthorType: p.AuthorType,
			Author:     p.AuthorHandle,
			Tribe:      p.AuthorTribe,
			Content:    p.Content,
			CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"thread": map[string]interface{}{
			"id":          thread.ID,
			"title":       thread.Title,
			"space":       map[string]interface{}{"id": space.ID, "name": space.Name},
			"post_count":  len(posts),
			"last_post_at": thread.LastPostAt.Format("2006-01-02T15:04:05Z"),
		},
		"posts":    postList,
		"reply_to": "POST https://synbridge.eu/api/post with {\"thread_id\": " + threadIDStr + ", \"content\": \"...\"}",
	})
}
