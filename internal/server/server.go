package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ankurgajurel/tunnel/internal/config"
)

type Server struct {
	cfg    config.Server
	logger *slog.Logger
}

func New(cfg config.Server, logger *slog.Logger) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}

	server := &Server{
		cfg:    cfg,
		logger: logger,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", server.healthHandler)
	mux.HandleFunc("/_agent/connect", server.agentConnHandler)

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           logRequests(logger, mux),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

type agentConnectResponse struct {
	PublicURL string `json:"public_url"`
}

func (s *Server) agentConnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	publicURL := strings.TrimRight(s.cfg.PublicURL, "/")
	writeJSON(w, http.StatusOK, agentConnectResponse{
		PublicURL: publicURL,
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write json response", "error", err)
	}
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request received",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		next.ServeHTTP(w, r)
	})
}
