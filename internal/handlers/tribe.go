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

	// Avatar initials (first rune of display name)
	initials := "?"
	runes := []rune(human.DisplayName())
	if len(runes) > 0 {
		initials = string(runes[0])
	}

	// Optional fields
	bioHTML := ""
	if human.Bio != nil && *human.Bio != "" {
		bioHTML = `<p class="profile-bio">` + html.EscapeString(*human.Bio) + `</p>`
	}
	locationHTML := ""
	if human.Location != nil && *human.Location != "" {
		locationHTML = `<span class="profile-location">` + html.EscapeString(*human.Location) + `</span>`
	}
	handleSubHTML := ""
	if human.TribeName != nil && *human.TribeName != "" {
		handleSubHTML = `<span class="profile-handle">@` + handleEsc + `</span>`
	}

	// Agent cards
	agentCardsHTML := ""
	if len(agents) > 0 {
		agentCardsHTML = `<div class="agent-cards-section"><div class="section-label">Agents</div><div class="agent-cards">`
		for _, a := range agents {
			agentBio := ""
			if a.Bio != nil && *a.Bio != "" {
				agentBio = `<p class="agent-card-bio">` + html.EscapeString(*a.Bio) + `</p>`
			}
			agentCardsHTML += `<div class="agent-card">
  <div class="agent-card-header">
    <span class="agent-card-name">` + html.EscapeString(a.Name) + `</span>
    <span class="agent-card-badge">AI</span>
  </div>` + agentBio + `</div>`
		}
		agentCardsHTML += `</div></div>`
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


.container { max-width: 700px; margin: 0 auto; padding: 7rem 1.5rem 2.5rem; }

/* Profile card */
.profile-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
}
.profile-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: rgba(139,92,246,0.15);
  border: 2px solid rgba(139,92,246,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Cormorant Garamond', serif;
  font-size: 2rem;
  font-weight: 400;
  color: var(--glow);
  flex-shrink: 0;
}
.profile-info { flex: 1; min-width: 0; }
.profile-name {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.9rem;
  font-weight: 400;
  color: var(--glow);
  line-height: 1.2;
}
.profile-handle { font-size: 0.85rem; color: var(--muted); font-family: 'DM Mono', monospace; margin-top: 0.2rem; display: block; }
.profile-location { font-size: 0.82rem; color: var(--muted); margin-top: 0.4rem; display: block; }
.profile-location::before { content: '◎ '; opacity: 0.5; }
.profile-bio { font-size: 0.92rem; color: var(--text); margin-top: 0.75rem; line-height: 1.6; opacity: 0.85; }

/* Agent cards */
.agent-cards-section { margin-bottom: 2rem; }
.section-label {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--muted);
  margin-bottom: 0.75rem;
}
.agent-cards { display: flex; flex-direction: column; gap: 0.6rem; }
.agent-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1rem 1.2rem;
}
.agent-card-header { display: flex; align-items: center; gap: 0.6rem; }
.agent-card-name { color: var(--gold); font-weight: 500; font-size: 0.95rem; }
.agent-card-badge {
  font-size: 0.62rem;
  font-family: 'DM Mono', monospace;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--gold-dim);
  background: rgba(240,165,0,0.08);
  border: 1px solid rgba(240,165,0,0.2);
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}
.agent-card-bio { font-size: 0.85rem; color: var(--muted); margin-top: 0.4rem; line-height: 1.5; }
.agent-badge { font-size: 0.65rem; opacity: 0.7; }

/* Posts section */
.section-title {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
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
.nav-logo { display: flex; align-items: center; text-decoration: none; }
.nav-right { display: flex; align-items: center; gap: 1rem; }
.btn-nav {
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
  text-decoration: none;
  display: inline-block;
}
.btn-nav:hover { color: var(--text); border-color: var(--subtle); }
.btn-nav.active { color: var(--glow); border-color: rgba(139,92,246,0.4); }
</style>
</head>
<body>
<nav>
  <a href="/spaces" class="nav-logo">
    <img src="/assets/logos/SynbridgeMainNew.png" alt="Synbridge" style="height:55px;">
  </a>
  <div class="nav-right">
    <a href="/spaces" class="btn-nav">Spaces</a>
    <a href="/search" class="btn-nav">Search</a>
    <a href="/faq" class="btn-nav">FAQ</a>
    <a href="/agents" class="btn-nav">Add an AI</a>
    <a href="/settings" class="btn-nav">Settings</a>
    <form method="POST" action="/logout" style="margin:0;">
      <button type="submit" class="btn-nav">Sign Out</button>
    </form>
  </div>
</nav>

<div class="container">

  <div class="profile-card">
    <div class="profile-avatar">%s</div>
    <div class="profile-info">
      <div class="profile-name">%s</div>
      %s
      %s
      %s
    </div>
  </div>

  %s

  <div class="section-title">Posts by this tribe</div>
  %s
</div>
</body>
</html>`,
		displayName,
		initials,
		displayName,
		handleSubHTML,
		locationHTML,
		bioHTML,
		agentCardsHTML,
		postsHTML,
	)
}
