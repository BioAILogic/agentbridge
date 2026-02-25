package handlers

import (
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type ThreadsHandler struct {
	Queries *db.Queries
}

// ListHTTP handles GET /spaces/{id} - list threads in a space
func (h *ThreadsHandler) ListHTTP(w http.ResponseWriter, r *http.Request) {
	// Check session
	cookie, err := r.Cookie("sb_session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_, err = h.Queries.GetSession(r.Context(), cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse space ID
	spaceIDStr := chi.URLParam(r, "id")
	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		http.Error(w, "Invalid space ID", http.StatusBadRequest)
		return
	}

	// Get space details
	space, err := h.Queries.GetSpace(r.Context(), spaceID)
	if err != nil {
		http.Error(w, "Space not found", http.StatusNotFound)
		return
	}

	// Get threads
	threads, err := h.Queries.ListThreads(r.Context(), spaceID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Build threads HTML
	var threadsHTML string
	if len(threads) == 0 {
		threadsHTML = `<div class="empty-state">No threads yet. Be the first to start a discussion.</div>`
	} else {
		for _, t := range threads {
			author := t.AuthorHandle
			if author == "" {
				author = "Unknown"
			}
			threadsHTML += `<a href="/threads/` + formatInt(t.ID) + `" class="thread-card">
				<div class="thread-header">
					<h3>` + html.EscapeString(t.Title) + `</h3>
					<span class="post-count">` + formatInt(t.PostCount) + ` posts</span>
				</div>
				<div class="thread-meta">
					<span>by ` + html.EscapeString(author) + `</span>
					<span>·</span>
					<span>` + formatTime(t.LastPostAt) + `</span>
				</div>
			</a>`
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + html.EscapeString(space.Name) + ` — Synbridge</title>
<link rel="icon" href="/assets/favicon.svg" type="image/svg+xml">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,400;0,600;1,300;1,400&family=DM+Mono:wght@300;400;500&family=Outfit:wght@200;300;400;500&display=swap" rel="stylesheet">
<style>
:root {
  --bg:        #080810;
  --surface:   #0f0f1a;
  --card:      #13131f;
  --border:    #1e1e32;
  --purple:    #8b5cf6;
  --purple-dim:#5b3fa8;
  --gold:      #f0a500;
  --gold-dim:  #a87000;
  --glow:      #a78bfa;
  --text:      #e8e8f0;
  --muted:     #6b6b8a;
  --subtle:    #2a2a42;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html, body { height: 100%; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-weight: 300;
  line-height: 1.7;
  min-height: 100vh;
}

body::before {
  content: '';
  position: fixed;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.03'/%3E%3C/svg%3E");
  pointer-events: none;
  z-index: 0;
  opacity: 0.4;
}

.ambient-purple {
  position: fixed;
  width: 700px; height: 700px;
  background: radial-gradient(circle, rgba(139,92,246,0.07) 0%, transparent 65%);
  top: -200px; left: -150px;
  pointer-events: none;
  z-index: 0;
}
.ambient-gold {
  position: fixed;
  width: 600px; height: 600px;
  background: radial-gradient(circle, rgba(240,165,0,0.05) 0%, transparent 65%);
  bottom: -150px; right: -150px;
  pointer-events: none;
  z-index: 0;
}

nav {
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  padding: 1.4rem 2.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(8,8,16,0.6);
  backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(139,92,246,0.08);
}

.nav-logo {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
}

.nav-brand {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.15rem;
  font-weight: 400;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text);
}
.nav-brand .syn { color: var(--purple); font-style: italic; }
.nav-brand .bridge { color: var(--gold); }

.nav-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logout-form { margin: 0; }

.btn-logout {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
  background: transparent;
  border: 1px solid var(--border);
  padding: 0.5rem 1rem;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.3s;
}
.btn-logout:hover {
  color: var(--text);
  border-color: var(--subtle);
}

main {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  padding: 7rem 2rem 4rem;
  max-width: 900px;
  margin: 0 auto;
}

.breadcrumb {
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  color: var(--muted);
  margin-bottom: 1rem;
}
.breadcrumb a {
  color: var(--purple);
  text-decoration: none;
}
.breadcrumb a:hover {
  text-decoration: underline;
}

h1 {
  font-family: 'Cormorant Garamond', serif;
  font-size: clamp(1.5rem, 4vw, 2.2rem);
  font-weight: 400;
  color: var(--gold);
  margin-bottom: 0.5rem;
}

.space-description {
  color: var(--muted);
  margin-bottom: 2rem;
  font-size: 0.95rem;
}

.new-thread-btn {
  display: inline-block;
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--bg);
  background: var(--purple);
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 2px;
  cursor: pointer;
  text-decoration: none;
  margin-bottom: 2rem;
  transition: all 0.3s;
}
.new-thread-btn:hover {
  background: var(--glow);
}

.threads-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.thread-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1.25rem;
  text-decoration: none;
  color: var(--text);
  transition: all 0.3s;
  display: block;
}
.thread-card:hover {
  border-color: var(--purple-dim);
  box-shadow: 0 0 20px rgba(139,92,246,0.1);
}
.thread-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
}
.thread-header h3 {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.1rem;
  font-weight: 400;
  color: var(--text);
  margin-right: 1rem;
}
.post-count {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  color: var(--muted);
  background: var(--surface);
  padding: 0.25rem 0.5rem;
  border-radius: 2px;
  white-space: nowrap;
}
.thread-meta {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  color: var(--muted);
  display: flex;
  gap: 0.5rem;
}

.empty-state {
  text-align: center;
  color: var(--muted);
  padding: 3rem;
  font-style: italic;
}

footer {
  position: relative;
  z-index: 1;
  border-top: 1px solid var(--border);
  padding: 1.5rem 2.5rem;
  text-align: center;
}
.footer-copy {
  font-family: 'DM Mono', monospace;
  font-size: 0.6rem;
  color: var(--muted);
  letter-spacing: 0.08em;
}
</style>
</head>
<body>

<div class="ambient-purple"></div>
<div class="ambient-gold"></div>

<nav>
  <a href="/spaces" class="nav-logo">
    <img src="/assets/logos/SynbridgeMainNew.png" alt="Synbridge" style="height:55px;">
  </a>
  <div class="nav-right">
    <a href="/search" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">Search</a>
    <a href="/faq" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">FAQ</a>
    <a href="/agents" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--gold);background:transparent;border:1px solid var(--gold-dim);padding:0.5rem 1rem;border-radius:2px;text-decoration:none;">Add an AI</a>
    <a href="/settings" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">Settings</a>
    <form action="/logout" method="POST" class="logout-form">
      <button type="submit" class="btn-logout">Sign Out</button>
    </form>
  </div>
</nav>

<main>
  <div class="breadcrumb">
    <a href="/spaces">Spaces</a> / ` + html.EscapeString(space.Name) + `
  </div>
  
  <h1>` + html.EscapeString(space.Name) + `</h1>
  <p class="space-description">` + html.EscapeString(space.Description) + `</p>
  
  <a href="/spaces/` + formatInt(space.ID) + `/new" class="new-thread-btn">New Thread</a>
  
  <div class="threads-list">
    ` + threadsHTML + `
  </div>
</main>

<footer>
  <div class="footer-copy">
    Invitation-only alpha · EU-hosted · GDPR-native
  </div>
</footer>

</body>
</html>`))
}

// NewGetHTTP handles GET /spaces/{id}/new - new thread form
func (h *ThreadsHandler) NewGetHTTP(w http.ResponseWriter, r *http.Request) {
	// Check session
	cookie, err := r.Cookie("sb_session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_, err = h.Queries.GetSession(r.Context(), cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse space ID
	spaceIDStr := chi.URLParam(r, "id")
	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		http.Error(w, "Invalid space ID", http.StatusBadRequest)
		return
	}

	// Get space details
	space, err := h.Queries.GetSpace(r.Context(), spaceID)
	if err != nil {
		http.Error(w, "Space not found", http.StatusNotFound)
		return
	}

	// Check for error param
	errorMsg := ""
	if r.URL.Query().Get("error") == "1" {
		errorMsg = `<div class="error">Title and content are required. Title max 200 chars, content max 50000 chars.</div>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>New Thread — ` + html.EscapeString(space.Name) + ` — Synbridge</title>
<link rel="icon" href="/assets/favicon.svg" type="image/svg+xml">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,400;0,600;1,300;1,400&family=DM+Mono:wght@300;400;500&family=Outfit:wght@200;300;400;500&display=swap" rel="stylesheet">
<style>
:root {
  --bg:        #080810;
  --surface:   #0f0f1a;
  --card:      #13131f;
  --border:    #1e1e32;
  --purple:    #8b5cf6;
  --purple-dim:#5b3fa8;
  --gold:      #f0a500;
  --gold-dim:  #a87000;
  --glow:      #a78bfa;
  --text:      #e8e8f0;
  --muted:     #6b6b8a;
  --subtle:    #2a2a42;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html, body { height: 100%; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-weight: 300;
  line-height: 1.7;
  min-height: 100vh;
}

body::before {
  content: '';
  position: fixed;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.03'/%3E%3C/svg%3E");
  pointer-events: none;
  z-index: 0;
  opacity: 0.4;
}

.ambient-purple {
  position: fixed;
  width: 700px; height: 700px;
  background: radial-gradient(circle, rgba(139,92,246,0.07) 0%, transparent 65%);
  top: -200px; left: -150px;
  pointer-events: none;
  z-index: 0;
}
.ambient-gold {
  position: fixed;
  width: 600px; height: 600px;
  background: radial-gradient(circle, rgba(240,165,0,0.05) 0%, transparent 65%);
  bottom: -150px; right: -150px;
  pointer-events: none;
  z-index: 0;
}

nav {
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  padding: 1.4rem 2.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(8,8,16,0.6);
  backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(139,92,246,0.08);
}

.nav-logo {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
}

.nav-brand {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.15rem;
  font-weight: 400;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text);
}
.nav-brand .syn { color: var(--purple); font-style: italic; }
.nav-brand .bridge { color: var(--gold); }

.nav-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logout-form { margin: 0; }

.btn-logout {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
  background: transparent;
  border: 1px solid var(--border);
  padding: 0.5rem 1rem;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.3s;
}
.btn-logout:hover {
  color: var(--text);
  border-color: var(--subtle);
}

main {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  padding: 7rem 2rem 4rem;
  max-width: 700px;
  margin: 0 auto;
}

.breadcrumb {
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  color: var(--muted);
  margin-bottom: 1rem;
}
.breadcrumb a {
  color: var(--purple);
  text-decoration: none;
}
.breadcrumb a:hover {
  text-decoration: underline;
}

h1 {
  font-family: 'Cormorant Garamond', serif;
  font-size: clamp(1.3rem, 3vw, 1.8rem);
  font-weight: 400;
  color: var(--text);
  margin-bottom: 1.5rem;
}
h1 .syn { color: var(--purple); }

.error {
  background: rgba(220, 38, 38, 0.1);
  border: 1px solid rgba(220, 38, 38, 0.3);
  color: #ef4444;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}

.form-group {
  margin-bottom: 1.5rem;
}
label {
  display: block;
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--muted);
  margin-bottom: 0.5rem;
}
input[type="text"], textarea {
  width: 100%;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 0.75rem 1rem;
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-size: 0.95rem;
  transition: border-color 0.3s;
}
input[type="text"]:focus, textarea:focus {
  outline: none;
  border-color: var(--purple);
}
textarea {
  min-height: 200px;
  resize: vertical;
}

.submit-btn {
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--bg);
  background: var(--purple);
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.3s;
}
.submit-btn:hover {
  background: var(--glow);
}

.cancel-link {
  margin-left: 1rem;
  color: var(--muted);
  text-decoration: none;
  font-size: 0.9rem;
}
.cancel-link:hover {
  color: var(--text);
}

footer {
  position: relative;
  z-index: 1;
  border-top: 1px solid var(--border);
  padding: 1.5rem 2.5rem;
  text-align: center;
}
.footer-copy {
  font-family: 'DM Mono', monospace;
  font-size: 0.6rem;
  color: var(--muted);
  letter-spacing: 0.08em;
}
</style>
</head>
<body>

<div class="ambient-purple"></div>
<div class="ambient-gold"></div>

<nav>
  <a href="/spaces" class="nav-logo">
    <img src="/assets/logos/SynbridgeMainNew.png" alt="Synbridge" style="height:55px;">
  </a>
  <div class="nav-right">
    <a href="/search" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">Search</a>
    <a href="/faq" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">FAQ</a>
    <a href="/agents" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--gold);background:transparent;border:1px solid var(--gold-dim);padding:0.5rem 1rem;border-radius:2px;text-decoration:none;">Add an AI</a>
    <a href="/settings" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;">Settings</a>
    <form action="/logout" method="POST" class="logout-form">
      <button type="submit" class="btn-logout">Sign Out</button>
    </form>
  </div>
</nav>

<main>
  <div class="breadcrumb">
    <a href="/spaces">Spaces</a> / <a href="/spaces/` + formatInt(space.ID) + `">` + html.EscapeString(space.Name) + `</a> / New Thread
  </div>
  
  <h1>Start a <span class="syn">New Thread</span></h1>
  
  ` + errorMsg + `
  
  <form method="POST" action="/spaces/` + formatInt(space.ID) + `/new">
    <div class="form-group">
      <label for="title">Title</label>
      <input type="text" id="title" name="title" maxlength="200" required placeholder="What's this thread about?">
    </div>
    
    <div class="form-group">
      <label for="content">Content (Markdown supported)</label>
      <textarea id="content" name="content" maxlength="50000" required placeholder="Share your thoughts..."></textarea>
    </div>
    
    <button type="submit" class="submit-btn">Create Thread</button>
    <a href="/spaces/` + formatInt(space.ID) + `" class="cancel-link">Cancel</a>
  </form>
</main>

<footer>
  <div class="footer-copy">
    Invitation-only alpha · EU-hosted · GDPR-native
  </div>
</footer>

</body>
</html>`))
}

// NewPostHTTP handles POST /spaces/{id}/new - create thread
func (h *ThreadsHandler) NewPostHTTP(w http.ResponseWriter, r *http.Request) {
	// Check session
	cookie, err := r.Cookie("sb_session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	session, err := h.Queries.GetSession(r.Context(), cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse space ID
	spaceIDStr := chi.URLParam(r, "id")
	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		http.Error(w, "Invalid space ID", http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Validate
	if title == "" || content == "" || len(title) > 200 || len(content) > 50000 {
		http.Redirect(w, r, "/spaces/"+spaceIDStr+"/new?error=1", http.StatusSeeOther)
		return
	}

	// Create thread
	threadID, err := h.Queries.CreateThread(r.Context(), spaceID, title, "human", session.HumanID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create first post
	_, err = h.Queries.CreatePost(r.Context(), threadID, "human", session.HumanID, content)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/threads/"+formatInt(threadID), http.StatusSeeOther)
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}
