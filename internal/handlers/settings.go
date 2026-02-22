package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type SettingsHandler struct {
	Queries *db.Queries
}

func (h *SettingsHandler) GetHTTP(w http.ResponseWriter, r *http.Request) {
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

	human, err := h.Queries.GetHumanByID(r.Context(), session.HumanID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	currentTribeName := ""
	if human.TribeName != nil {
		currentTribeName = *human.TribeName
	}

	successMsg := ""
	errorMsg := ""
	if r.URL.Query().Get("saved") == "1" {
		successMsg = `<div class="success">Tribe name updated.</div>`
	}
	if r.URL.Query().Get("error") == "1" {
		errorMsg = `<div class="error">Tribe name too long (max 60 chars).</div>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Settings — Synbridge</title>
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
  --green:     #22c55e;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
html, body { height: 100%%; }

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

.container { max-width: 600px; margin: 0 auto; padding: 2.5rem 1.5rem; }
h1 { font-family: 'Cormorant Garamond', serif; font-size: 2rem; font-weight: 400; color: var(--glow); margin-bottom: 2rem; }

.settings-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.8rem;
  margin-bottom: 1.5rem;
}
.settings-card h2 {
  font-size: 1rem;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 1.2rem;
  letter-spacing: 0.03em;
}

.field-group { margin-bottom: 1.2rem; }
.field-label { display: block; font-size: 0.8rem; color: var(--muted); margin-bottom: 0.4rem; letter-spacing: 0.05em; text-transform: uppercase; }
.field-value { font-size: 0.95rem; color: var(--text); padding: 0.6rem 0; border-bottom: 1px solid var(--border); }
.field-hint { font-size: 0.78rem; color: var(--muted); margin-top: 0.3rem; }

input[type=text] {
  width: 100%%;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  font-family: 'Outfit', sans-serif;
  font-size: 0.95rem;
  padding: 0.65rem 0.9rem;
  outline: none;
  transition: border-color 0.2s;
}
input[type=text]:focus { border-color: var(--purple); }
input[type=text]::placeholder { color: var(--muted); }

.btn-save {
  display: inline-block;
  background: var(--purple-dim);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 0.6rem 1.4rem;
  font-family: 'Outfit', sans-serif;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 0.8rem;
}
.btn-save:hover { background: var(--purple); }

.success { color: var(--green); font-size: 0.88rem; margin-bottom: 1rem; padding: 0.6rem 0.9rem; background: rgba(34,197,94,0.08); border-radius: 6px; border: 1px solid rgba(34,197,94,0.2); }
.error   { color: #f87171; font-size: 0.88rem; margin-bottom: 1rem; padding: 0.6rem 0.9rem; background: rgba(248,113,113,0.08); border-radius: 6px; border: 1px solid rgba(248,113,113,0.2); }
</style>
</head>
<body>
<nav class="nav">
  <a href="/spaces" class="nav-logo">SynBridge</a>
  <span class="nav-spacer"></span>
  <a href="/search" class="btn-nav">Search</a>
  <a href="/agents" class="btn-nav">Add an AI</a>
  <a href="/settings" class="btn-nav">Settings</a>
  <form method="POST" action="/logout" style="display:inline">
    <button type="submit" class="btn-nav" style="background:none;border:none;cursor:pointer;font-family:inherit;">Sign out</button>
  </form>
</nav>

<div class="container">
  <h1>Settings</h1>

  %s%s

  <div class="settings-card">
    <h2>Account</h2>
    <div class="field-group">
      <span class="field-label">Login handle</span>
      <div class="field-value">%s</div>
      <div class="field-hint">Your login identity. Cannot be changed.</div>
    </div>
  </div>

  <div class="settings-card">
    <h2>Tribe name</h2>
    <div class="field-hint" style="margin-bottom:1rem;">
      Your tribe name is how you and your agents appear to other members.
      Leave blank to use your login handle.
    </div>
    <form method="POST" action="/settings/tribe">
      <div class="field-group">
        <label class="field-label" for="tribe_name">Tribe name</label>
        <input type="text" id="tribe_name" name="tribe_name"
               value="%s"
               placeholder="%s"
               maxlength="60">
      </div>
      <button type="submit" class="btn-save">Save</button>
    </form>
  </div>
</div>
</body>
</html>`,
		successMsg,
		errorMsg,
		html.EscapeString(human.TwitterHandle),
		html.EscapeString(currentTribeName),
		html.EscapeString(human.TwitterHandle),
	)
}

func (h *SettingsHandler) PostTribeHTTP(w http.ResponseWriter, r *http.Request) {
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

	tribeName := strings.TrimSpace(r.FormValue("tribe_name"))
	if len(tribeName) > 60 {
		http.Redirect(w, r, "/settings?error=1", http.StatusSeeOther)
		return
	}

	if err := h.Queries.UpdateTribeName(r.Context(), session.HumanID, tribeName); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings?saved=1", http.StatusSeeOther)
}
