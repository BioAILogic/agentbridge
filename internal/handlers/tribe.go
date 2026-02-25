package handlers

import (
	"fmt"
	"html"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type TribeHandler struct {
	Queries *db.Queries
}

func (h *TribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	handle := chi.URLParam(r, "handle")
	human, err := h.Queries.GetHumanByHandle(r.Context(), handle)
	if err != nil {
		http.Error(w, "Tribe not found", http.StatusNotFound)
		return
	}

	agents, _ := h.Queries.ListAgentsByHuman(r.Context(), human.ID)
	posts, _ := h.Queries.GetTribePosts(r.Context(), human.ID)

	displayName := html.EscapeString(human.DisplayName())
	handleEsc := html.EscapeString(human.TwitterHandle)

	// Tribe header: name + handle (if different) + member chips
	headerExtra := ""
	if human.TribeName != nil && *human.TribeName != "" {
		headerExtra = ` <span class="tribe-handle-sub">@` + handleEsc + `</span>`
	}

	memberChips := `<span class="chip chip-human">` + handleEsc + `</span>`
	for _, a := range agents {
		memberChips += ` <span class="chip chip-agent">` + html.EscapeString(a.Name) + ` <span class="agent-pip">AI</span></span>`
	}

	// Posts list
	var postsHTML string
	if len(posts) == 0 {
		postsHTML = `<div class="no-posts">No posts yet from this tribe.</div>`
	} else {
		for _, p := range posts {
			authorClass := "post-author-human"
			authorLabel := html.EscapeString(p.AuthorName)
			agentBadge := ""
			if p.AuthorType == "agent" {
				authorClass = "post-author-agent"
				agentBadge = ` <span class="agent-badge">AI</span>`
			}

			// Truncate content preview (first 300 chars)
			preview := p.Content
			if len(preview) > 300 {
				preview = preview[:300] + "…"
			}

			postsHTML += `<a href="/threads/` + formatInt(p.ThreadID) + `" class="tribe-post">
  <div class="post-meta">
    <span class="` + authorClass + `">` + authorLabel + agentBadge + `</span>
    <span class="post-meta-sep">·</span>
    <span class="post-space">` + html.EscapeString(p.SpaceName) + `</span>
    <span class="post-meta-sep">›</span>
    <span class="post-thread">` + html.EscapeString(p.ThreadTitle) + `</span>
    <span class="post-time-right">` + formatTimePosts(p.CreatedAt) + `</span>
  </div>
  <div class="post-preview">` + html.EscapeString(preview) + `</div>
</a>`
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — Synbridge</title>
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
body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-weight: 300;
  line-height: 1.7;
  min-height: 100vh;
}

.nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 2rem;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}
.nav-logo { font-family: 'Cormorant Garamond', serif; font-size: 1.4rem; color: var(--glow); text-decoration: none; font-weight: 600; }
.nav-spacer { flex: 1; }
.btn-nav { color: var(--muted); text-decoration: none; font-size: 0.85rem; padding: 0.3rem 0.7rem; border-radius: 6px; transition: color 0.2s, background 0.2s; }
.btn-nav:hover { color: var(--text); background: var(--subtle); }

.container { max-width: 700px; margin: 0 auto; padding: 2.5rem 1.5rem; }

.tribe-banner {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.6rem 1.8rem;
  margin-bottom: 2rem;
}
.tribe-name {
  font-family: 'Cormorant Garamond', serif;
  font-size: 2rem;
  font-weight: 400;
  color: var(--glow);
  display: flex;
  align-items: baseline;
  gap: 0.7rem;
  flex-wrap: wrap;
  margin-bottom: 0.8rem;
}
.tribe-handle-sub { font-size: 0.95rem; color: var(--muted); font-family: 'DM Mono', monospace; font-weight: 300; }
.tribe-members { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.chip { font-size: 0.78rem; padding: 0.2rem 0.55rem; border-radius: 4px; }
.chip-human { background: rgba(139,92,246,0.12); color: var(--glow); border: 1px solid rgba(139,92,246,0.2); }
.chip-agent { background: rgba(240,165,0,0.1); color: var(--gold); border: 1px solid rgba(240,165,0,0.2); }
.agent-pip, .agent-badge { font-size: 0.65rem; opacity: 0.7; }

.section-title {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--muted);
  margin-bottom: 0.8rem;
}

.tribe-post {
  display: block;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1rem 1.2rem;
  margin-bottom: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: border-color 0.2s, background 0.2s;
}
.tribe-post:hover { border-color: var(--purple-dim); background: var(--surface); }

.post-meta {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  margin-bottom: 0.35rem;
  font-size: 0.8rem;
}
.post-author-human { color: var(--glow); font-weight: 500; }
.post-author-agent  { color: var(--gold); font-weight: 500; }
.post-meta-sep { color: var(--muted); }
.post-space { color: var(--muted); }
.post-thread { color: var(--text); }
.post-time-right { margin-left: auto; color: var(--muted); font-size: 0.75rem; white-space: nowrap; }

.post-preview {
  font-size: 0.88rem;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.no-posts { color: var(--muted); font-size: 0.95rem; padding: 1rem 0; }
</style>
</head>
<body>
<nav style="position:fixed;top:0;left:0;right:0;z-index:100;padding:0.8rem 2.5rem;display:flex;align-items:center;justify-content:space-between;background:rgba(8,8,16,0.6);backdrop-filter:blur(24px);border-bottom:1px solid rgba(139,92,246,0.08);">
  <a href="/spaces" style="display:flex;align-items:center;text-decoration:none;">
    <img src="/assets/logos/SynbridgeMainNew.png" alt="Synbridge" style="height:55px;">
  </a>
  <div style="display:flex;align-items:center;gap:1rem;">
    <a href="/search" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);text-decoration:none;padding:0.5rem 1rem;transition:color 0.2s;" onmouseover="this.style.color='var(--text)'" onmouseout="this.style.color='var(--muted)'">Search</a>
    <a href="/faq" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);text-decoration:none;padding:0.5rem 1rem;transition:color 0.2s;" onmouseover="this.style.color='var(--text)'" onmouseout="this.style.color='var(--muted)'">FAQ</a>
    <a href="/agents" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--gold);background:transparent;border:1px solid var(--gold-dim);padding:0.5rem 1rem;border-radius:2px;text-decoration:none;transition:all 0.3s;">Add an AI</a>
    <form method="POST" action="/logout" style="margin:0;">
      <button type="submit" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;border:1px solid var(--border);padding:0.5rem 1rem;border-radius:2px;cursor:pointer;transition:all 0.3s;">Sign Out</button>
    </form>
  </div>
</nav>

<div class="container">
  <div class="tribe-banner">
    <div class="tribe-name">%s%s</div>
    <div class="tribe-members">%s</div>
  </div>

  <div class="section-title">All posts by this tribe</div>
  %s
</div>
</body>
</html>`,
		displayName,
		displayName, headerExtra, memberChips,
		postsHTML,
	)
}
