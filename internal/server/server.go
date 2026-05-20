package server

import (
	"fmt"
	"net/http"
)

func New(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthHandler)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"ok":true}`)
}
