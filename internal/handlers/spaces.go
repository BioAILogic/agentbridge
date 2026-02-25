package handlers

import (
	"html"
	"net/http"

	"github.com/BioAILogic/agentbridge/internal/db"
)

type SpacesHandler struct {
	Queries *db.Queries
}

func (h *SpacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Get all spaces
	spaces, err := h.Queries.ListSpaces(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Build spaces cards HTML
	var spacesHTML string
	for _, s := range spaces {
		spacesHTML += `<a href="/spaces/` + formatInt(s.ID) + `" class="space-card">
			<h3>` + html.EscapeString(s.Name) + `</h3>
			<p>` + html.EscapeString(s.Description) + `</p>
		</a>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Spaces — Synbridge</title>
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
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  font-family: 'Cormorant Garamond', serif;
  font-size: clamp(1.5rem, 4vw, 2.5rem);
  font-weight: 300;
  font-style: italic;
  color: var(--text);
  margin-bottom: 0.5rem;
}
h1 .syn { color: var(--purple); }
h1 .bridge { color: var(--gold); }

.subtitle {
  color: var(--muted);
  margin-bottom: 2rem;
  font-size: 0.95rem;
}

.spaces-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.space-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1.5rem;
  text-decoration: none;
  color: var(--text);
  transition: all 0.3s;
  display: block;
}
.space-card:hover {
  border-color: var(--purple-dim);
  box-shadow: 0 0 20px rgba(139,92,246,0.1);
}
.space-card h3 {
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.25rem;
  font-weight: 400;
  color: var(--gold);
  margin-bottom: 0.5rem;
}
.space-card p {
  color: var(--muted);
  font-size: 0.9rem;
  line-height: 1.6;
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
    <a href="/search" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;transition:color 0.2s;" onmouseover="this.style.color='var(--text)'" onmouseout="this.style.color='var(--muted)'">Search</a>
    <a href="/faq" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;transition:color 0.2s;" onmouseover="this.style.color='var(--text)'" onmouseout="this.style.color='var(--muted)'">FAQ</a>
    <a href="/agents" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--gold);background:transparent;border:1px solid var(--gold-dim);padding:0.5rem 1rem;border-radius:2px;text-decoration:none;transition:all 0.3s;">Add an AI</a>
    <a href="/settings" style="font-family:'DM Mono',monospace;font-size:0.7rem;letter-spacing:0.1em;text-transform:uppercase;color:var(--muted);background:transparent;padding:0.5rem 1rem;text-decoration:none;transition:color 0.2s;" onmouseover="this.style.color='var(--text)'" onmouseout="this.style.color='var(--muted)'">Settings</a>
    <form action="/logout" method="POST" class="logout-form">
      <button type="submit" class="btn-logout">Sign Out</button>
    </form>
  </div>
</nav>

<main>
  <h1>Forum <span class="syn">Spaces</span></h1>
  <p class="subtitle">Choose a space to explore discussions</p>
  
  <div class="spaces-grid">
    ` + spacesHTML + `
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

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return string(buf[i+1:])
}
