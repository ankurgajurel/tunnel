package main

import (
	"testing"

	"github.com/ankurgajurel/tunnel/internal/config"
)

func TestParseHTTPArgsUsesSavedConfig(t *testing.T) {
	opts, err := parseHTTPArgs([]string{"5050"}, config.Agent{
		ServerURL: "http://localhost:8080",
		Token:     "secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	if opts.port != 5050 {
		t.Fatalf("port = %d, want 5050", opts.port)
	}
	if opts.workers != 4 {
		t.Fatalf("workers = %d, want 4", opts.workers)
	}
	if opts.serverURL != "http://localhost:8080" {
		t.Fatalf("serverURL = %q", opts.serverURL)
	}
	if opts.token != "secret" {
		t.Fatalf("token = %q", opts.token)
	}
}

func TestParseHTTPArgsAllowsFlagOverrides(t *testing.T) {
	opts, err := parseHTTPArgs([]string{
		"5050",
		"--workers", "2",
		"--server-url", "localhost:8080",
		"--token", "secret",
	}, config.Agent{})
	if err != nil {
		t.Fatal(err)
	}

	if opts.workers != 2 {
		t.Fatalf("workers = %d, want 2", opts.workers)
	}
	if opts.serverURL != "http://localhost:8080" {
		t.Fatalf("serverURL = %q", opts.serverURL)
	}
	if opts.token != "secret" {
		t.Fatalf("token = %q", opts.token)
	}
}

func TestParseHTTPArgsRejectsBadWorkers(t *testing.T) {
	_, err := parseHTTPArgs([]string{"5050", "--workers", "0"}, config.Agent{
		ServerURL: "http://localhost:8080",
		Token:     "secret",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
