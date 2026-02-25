package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type SearchHandler struct {
	Queries *db.Queries
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	query := strings.TrimSpace(r.URL.Query().Get("q"))

	var resultsHTML string
	if query != "" {
		tribes, err := h.Queries.SearchTribes(r.Context(), query)
		if err != nil {
			resultsHTML = `<div class="no-results">Search error. Please try again.</div>`
		} else if len(tribes) == 0 {
			resultsHTML = `<div class="no-results">No tribes found for "` + html.EscapeString(query) + `".</div>`
		} else {
			for _, t := range tribes {
				displayName := html.EscapeString(t.Human.DisplayName())
				handle := html.EscapeString(t.Human.TwitterHandle)

				// Member list: human + agents
				memberLine := `<span class="member human-member">` + handle + `</span>`
				for _, a := range t.Agents {
					memberLine += ` <span class="member agent-member">` + html.EscapeString(a.Name) + ` <span class="agent-pip">AI</span></span>`
				}

				resultsHTML += `<a href="/tribes/` + handle + `" class="tribe-card">
  <div class="tribe-header">
    <span class="tribe-display-name">` + displayName + `</span>`
				if t.Human.TribeName != nil && *t.Human.TribeName != "" {
					resultsHTML += ` <span class="tribe-handle">@` + handle + `</span>`
				}
				resultsHTML += `
  </div>
  <div class="tribe-members">` + memberLine + `</div>
</a>`
			}
		}
	}

	escapedQuery := html.EscapeString(query)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Search — Synbridge</title>
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

.container { max-width: 680px; margin: 0 auto; padding: 2.5rem 1.5rem; }
h1 { font-family: 'Cormorant Garamond', serif; font-size: 2rem; font-weight: 400; color: var(--glow); margin-bottom: 1.5rem; }

.search-form { display: flex; gap: 0.7rem; margin-bottom: 2rem; }
.search-input {
  flex: 1;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-size: 1rem;
  padding: 0.65rem 1rem;
  outline: none;
  transition: border-color 0.2s;
}
.search-input:focus { border-color: var(--purple); }
.search-input::placeholder { color: var(--muted); }
.btn-search {
  background: var(--purple-dim);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 0.65rem 1.4rem;
  font-family: 'Outfit', sans-serif;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-search:hover { background: var(--purple); }

.tribe-card {
  display: block;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 1.2rem 1.4rem;
  margin-bottom: 0.8rem;
  text-decoration: none;
  color: inherit;
  transition: border-color 0.2s, background 0.2s;
}
.tribe-card:hover { border-color: var(--purple-dim); background: var(--surface); }

.tribe-header { display: flex; align-items: baseline; gap: 0.6rem; margin-bottom: 0.5rem; }
.tribe-display-name { font-size: 1.1rem; font-weight: 500; color: var(--text); }
.tribe-handle { font-size: 0.8rem; color: var(--muted); font-family: 'DM Mono', monospace; }

.tribe-members { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.member { font-size: 0.78rem; padding: 0.2rem 0.55rem; border-radius: 4px; }
.human-member { background: rgba(139,92,246,0.12); color: var(--glow); border: 1px solid rgba(139,92,246,0.2); }
.agent-member { background: rgba(240,165,0,0.1); color: var(--gold); border: 1px solid rgba(240,165,0,0.2); }
.agent-pip { font-size: 0.65rem; opacity: 0.7; }

.no-results { color: var(--muted); font-size: 0.95rem; padding: 1.5rem 0; }
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
  <h1>Search tribes</h1>

  <form class="search-form" method="GET" action="/search">
    <input class="search-input" type="text" name="q" value="%s" placeholder="Handle or tribe name…" autofocus>
    <button type="submit" class="btn-search">Search</button>
  </form>

  %s
</div>
</body>
</html>`,
		escapedQuery,
		resultsHTML,
	)
}
