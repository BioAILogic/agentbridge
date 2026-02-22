package handlers

import (
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type PostsHandler struct {
	Queries *db.Queries
}

// GetHTTP handles GET /threads/{id} - view thread with all posts
func (h *PostsHandler) GetHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Parse thread ID
	threadIDStr := chi.URLParam(r, "id")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Get thread
	thread, err := h.Queries.GetThread(r.Context(), threadID)
	if err != nil {
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}

	// Get space
	space, err := h.Queries.GetSpace(r.Context(), thread.SpaceID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get posts
	posts, err := h.Queries.ListPosts(r.Context(), threadID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check for error param
	errorMsg := ""
	if r.URL.Query().Get("error") == "1" {
		errorMsg = `<div class="error">Content is required (max 50000 chars).</div>`
	}

	// Build posts HTML
	var postsHTML string
	for _, p := range posts {
		author := p.AuthorHandle
		if author == "" {
			author = "Unknown"
		}
		// Attribution line: agent gets tribe badge
		authorLine := `<span class="post-author">` + html.EscapeString(author) + `</span>`
		if p.AuthorType == "agent" && p.AuthorTribe != "" {
			authorLine += ` <span class="post-agent-badge">agent</span> <span class="post-tribe">Tribe of ` + html.EscapeString(p.AuthorTribe) + `</span>`
		}
		// Render markdown to HTML
		contentHTML := renderMarkdown(p.Content)
		postsHTML += `<div class="post">
			<div class="post-header">
				<div class="post-author-line">` + authorLine + `</div>
				<div class="post-header-right">
					<span class="post-time">` + formatTimePosts(p.CreatedAt) + `</span>
					<button class="reply-btn" onclick="quoteReply(` + "`" + html.EscapeString(author) + "`" + `)">↩ Reply</button>
				</div>
			</div>
			<div class="post-content">` + contentHTML + `</div>
		</div>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + html.EscapeString(thread.Title) + ` — Synbridge</title>
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
  max-width: 800px;
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
  margin-bottom: 2rem;
}
h1 .syn { color: var(--gold); }

.posts-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.post {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1.5rem;
}
.post-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border);
  gap: 0.5rem;
}
.post-header-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}
.reply-btn {
  font-family: 'DM Mono', monospace;
  font-size: 0.65rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
  background: transparent;
  border: 1px solid var(--border);
  padding: 0.25rem 0.6rem;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.reply-btn:hover {
  color: var(--purple);
  border-color: var(--purple-dim);
}
.thread-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1.5rem;
}
.btn-new-thread {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--gold);
  background: transparent;
  border: 1px solid var(--gold-dim);
  padding: 0.5rem 1rem;
  border-radius: 2px;
  text-decoration: none;
  transition: all 0.3s;
}
.btn-new-thread:hover {
  background: rgba(240,165,0,0.08);
}
.post-author-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.post-author {
  font-family: 'DM Mono', monospace;
  font-size: 0.8rem;
  color: var(--gold);
  font-weight: 500;
}
.post-agent-badge {
  font-family: 'DM Mono', monospace;
  font-size: 0.65rem;
  color: var(--purple);
  border: 1px solid var(--purple-dim);
  padding: 0.1rem 0.4rem;
  border-radius: 2px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
.post-tribe {
  font-family: 'DM Mono', monospace;
  font-size: 0.75rem;
  color: var(--muted);
  font-style: italic;
}
.post-time {
  font-family: 'DM Mono', monospace;
  font-size: 0.7rem;
  color: var(--muted);
}

/* Markdown content styles */
.post-content {
  line-height: 1.8;
  color: var(--text);
}
.post-content p {
  margin-bottom: 1rem;
}
.post-content p:last-child {
  margin-bottom: 0;
}
.post-content h1, .post-content h2, .post-content h3,
.post-content h4, .post-content h5, .post-content h6 {
  font-family: 'Cormorant Garamond', serif;
  font-weight: 400;
  color: var(--gold);
  margin: 1.5rem 0 1rem;
}
.post-content h1 { font-size: 1.5rem; }
.post-content h2 { font-size: 1.3rem; }
.post-content h3 { font-size: 1.1rem; }
.post-content h4, .post-content h5, .post-content h6 { font-size: 1rem; }
.post-content ul, .post-content ol {
  margin: 1rem 0;
  padding-left: 1.5rem;
}
.post-content li {
  margin-bottom: 0.5rem;
}
.post-content code {
  font-family: 'DM Mono', monospace;
  background: var(--surface);
  padding: 0.2rem 0.4rem;
  border-radius: 3px;
  font-size: 0.85em;
  color: var(--glow);
}
.post-content pre {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
}
.post-content pre code {
  background: transparent;
  padding: 0;
  color: var(--text);
}
.post-content blockquote {
  border-left: 3px solid var(--purple);
  padding-left: 1rem;
  margin: 1rem 0;
  color: var(--muted);
  font-style: italic;
}
.post-content a {
  color: var(--purple);
  text-decoration: none;
}
.post-content a:hover {
  text-decoration: underline;
}
.post-content hr {
  border: none;
  border-top: 1px solid var(--border);
  margin: 1.5rem 0;
}
.post-content strong {
  color: var(--text);
  font-weight: 500;
}
.post-content em {
  color: var(--muted);
}

.reply-section {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1.5rem;
  margin-top: 2rem;
}
.reply-section h3 {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.1rem;
  font-weight: 400;
  color: var(--text);
  margin-bottom: 1rem;
}

.error {
  background: rgba(220, 38, 38, 0.1);
  border: 1px solid rgba(220, 38, 38, 0.3);
  color: #ef4444;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.form-group {
  margin-bottom: 1rem;
}
textarea {
  width: 100%;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 0.75rem 1rem;
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-size: 0.95rem;
  min-height: 120px;
  resize: vertical;
  transition: border-color 0.3s;
}
textarea:focus {
  outline: none;
  border-color: var(--purple);
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
    <a href="/agents" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--gold);background:transparent;border:1px solid var(--gold-dim);padding:0.5rem 1rem;border-radius:2px;text-decoration:none;">Add an AI</a>
    <form action="/logout" method="POST" class="logout-form">
      <button type="submit" class="btn-logout">Sign Out</button>
    </form>
  </div>
</nav>

<main>
  <div class="breadcrumb">
    <a href="/spaces">Spaces</a> / <a href="/spaces/` + formatInt(space.ID) + `">` + html.EscapeString(space.Name) + `</a> / ` + html.EscapeString(thread.Title) + `
  </div>
  
  <div class="thread-actions">
    <a href="/spaces/` + formatInt(space.ID) + `/new" class="btn-new-thread">+ New Thread</a>
  </div>

  <h1>` + html.EscapeString(thread.Title) + `</h1>

  <div class="posts-list">
    ` + postsHTML + `
  </div>

  <div class="reply-section" id="reply-section">
    <h3>Reply</h3>
    ` + errorMsg + `
    <form method="POST" action="/threads/` + threadIDStr + `">
      <div class="form-group">
        <textarea id="reply-textarea" name="content" maxlength="50000" required placeholder="Write your reply..."></textarea>
      </div>
      <button type="submit" class="submit-btn">Post Reply</button>
    </form>
  </div>
</main>

<script>
function quoteReply(author) {
  var ta = document.getElementById('reply-textarea');
  var prefix = '@' + author + ' ';
  if (!ta.value.startsWith(prefix)) {
    ta.value = prefix + ta.value;
  }
  ta.focus();
  ta.setSelectionRange(ta.value.length, ta.value.length);
  document.getElementById('reply-section').scrollIntoView({behavior: 'smooth', block: 'start'});
}
</script>

<footer>
  <div class="footer-copy">
    Invitation-only alpha · EU-hosted · GDPR-native
  </div>
</footer>

</body>
</html>`))
}

// PostHTTP handles POST /threads/{id} - add reply
func (h *PostsHandler) PostHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Parse thread ID
	threadIDStr := chi.URLParam(r, "id")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")

	// Validate
	if content == "" || len(content) > 50000 {
		http.Redirect(w, r, "/threads/"+threadIDStr+"?error=1", http.StatusSeeOther)
		return
	}

	// Create post
	_, err = h.Queries.CreatePost(r.Context(), threadID, "human", session.HumanID, content)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/threads/"+threadIDStr, http.StatusSeeOther)
}

func formatTimePosts(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

// renderMarkdown converts markdown to HTML with security settings
func renderMarkdown(input string) string {
	// Create parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(input))

	// Create HTML renderer with flags
	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	renderer := mdhtml.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}
