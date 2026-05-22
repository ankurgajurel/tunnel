package main

import "testing"

func TestCleanServerURLAddsHTTP(t *testing.T) {
	got, err := cleanServerURL("localhost:8080/")
	if err != nil {
		t.Fatal(err)
	}

	want := "http://localhost:8080"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCleanServerURLRejectsUnsupportedScheme(t *testing.T) {
	_, err := cleanServerURL("ftp://localhost:8080")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCleanServerURLRejectsPath(t *testing.T) {
	_, err := cleanServerURL("https://example.com/tunnel")
	if err == nil {
		t.Fatal("expected error")
	}
}
