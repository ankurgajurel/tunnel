package server

import (
	"net/http/httptest"
	"testing"

	"github.com/ankurgajurel/tunnel/internal/config"
)

func TestFindTunnelFromLocalPath(t *testing.T) {
	s := testServer("localhost")
	tunnel, err := s.registry.Register("demo", "http://127.0.0.1:5050", "http://localhost:8080/t/demo")
	if err != nil {
		t.Fatalf("register tunnel: %v", err)
	}

	req := httptest.NewRequest("GET", "http://localhost:8080/t/demo/api/users?page=1", nil)
	got, path, ok := s.findTunnel(req)
	if !ok {
		t.Fatal("expected tunnel")
	}
	if got != tunnel {
		t.Fatal("got wrong tunnel")
	}
	if path != "/api/users?page=1" {
		t.Fatalf("path = %q", path)
	}
}

func TestFindTunnelFromSubdomainHost(t *testing.T) {
	s := testServer("example.com")
	tunnel, err := s.registry.Register("demo", "http://127.0.0.1:5050", "http://demo.example.com")
	if err != nil {
		t.Fatalf("register tunnel: %v", err)
	}

	req := httptest.NewRequest("GET", "http://demo.example.com/api/users?page=1", nil)
	got, path, ok := s.findTunnel(req)
	if !ok {
		t.Fatal("expected tunnel")
	}
	if got != tunnel {
		t.Fatal("got wrong tunnel")
	}
	if path != "/api/users?page=1" {
		t.Fatalf("path = %q", path)
	}
}

func TestPublicURLForLocalhostUsesSubdomainAndPort(t *testing.T) {
	s := &Server{
		cfg: config.Server{
			BaseDomain: "localhost",
			PublicURL:  "http://localhost:8080",
		},
	}

	got := s.publicURLFor("demo")
	if got != "http://demo.localhost:8080" {
		t.Fatalf("public url = %q", got)
	}
}

func TestPublicURLForDomainUsesBaseDomain(t *testing.T) {
	s := &Server{
		cfg: config.Server{
			BaseDomain: "tunnel.example.com",
			PublicURL:  "https://tunnel.example.com",
		},
	}

	got := s.publicURLFor("demo")
	if got != "https://demo.tunnel.example.com" {
		t.Fatalf("public url = %q", got)
	}
}

func testServer(baseDomain string) *Server {
	return &Server{
		cfg: config.Server{
			BaseDomain: baseDomain,
		},
		registry: NewRegistry(),
		pending:  make(map[string]*pendingRequest),
	}
}
