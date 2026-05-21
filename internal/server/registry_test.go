package server

import "testing"

func TestRegistryRegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	tunnel, err := registry.Register("demo", "http://127.0.0.1:5050", "http://localhost:8080/t/demo")
	if err != nil {
		t.Fatalf("register tunnel: %v", err)
	}

	if tunnel.ID == "" {
		t.Fatal("expected tunnel id")
	}

	got, ok := registry.Get("demo")
	if !ok {
		t.Fatal("expected tunnel by subdomain")
	}
	if got.TargetURL != "http://127.0.0.1:5050" {
		t.Fatalf("target url = %q", got.TargetURL)
	}

	gotByID, ok := registry.GetByID(tunnel.ID)
	if !ok {
		t.Fatal("expected tunnel by id")
	}
	if gotByID.Subdomain != "demo" {
		t.Fatalf("subdomain = %q", gotByID.Subdomain)
	}
}

func TestRegistryRejectsDuplicateSubdomain(t *testing.T) {
	registry := NewRegistry()

	if _, err := registry.Register("demo", "http://127.0.0.1:5050", "http://localhost:8080/t/demo"); err != nil {
		t.Fatalf("register first tunnel: %v", err)
	}
	if _, err := registry.Register("demo", "http://127.0.0.1:5051", "http://localhost:8080/t/demo"); err == nil {
		t.Fatal("expected duplicate subdomain error")
	}
}
