package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ankurgajurel/tunnel/internal/config"
)

func TestShouldWarnBrowserHTMLRequest(t *testing.T) {
	s := warningTestServer()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/t/demo", nil)
	req.Header.Set("Accept", "text/html")

	if !s.shouldWarn(req) {
		t.Fatal("expected warning")
	}
}

func TestShouldWarnSkipsAcknowledgedRequest(t *testing.T) {
	s := warningTestServer()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/t/demo", nil)
	req.Header.Set("Accept", "text/html")
	req.AddCookie(&http.Cookie{Name: s.cfg.WarningCookieName, Value: "1"})

	if s.shouldWarn(req) {
		t.Fatal("expected acknowledged request to skip warning")
	}
}

func warningTestServer() *Server {
	return &Server{
		cfg: config.Server{
			WarningCookieName: "tunnel_warning_ack",
		},
	}
}
