package server

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) shouldWarn(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		return false
	}
	if _, err := r.Cookie(s.cfg.WarningCookieName); err == nil {
		return false
	}
	return true
}

func (s *Server) renderWarning(w http.ResponseWriter, r *http.Request) {
	next := r.URL.RequestURI()
	continueURL := "/_tunnel/continue?next=" + url.QueryEscape(next)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
  <head><title>Tunnel warning</title></head>
  <body>
    <h1>You are opening a tunnel</h1>
    <p>This URL forwards traffic to a developer-controlled local system.</p>
    <p><a href="%s">Continue to %s</a></p>
  </body>
</html>`, html.EscapeString(continueURL), html.EscapeString(next))
}

func (s *Server) continueHandler(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.WarningCookieName,
		Value:    "1",
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, next, http.StatusFound)
}
