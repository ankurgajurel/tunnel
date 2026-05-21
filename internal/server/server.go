package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ankurgajurel/tunnel/internal/config"
)

type Server struct {
	cfg       config.Server
	logger    *slog.Logger
	registry  *Registry
	pendingMu sync.Mutex
	pending   map[string]*pendingRequest
}

func New(cfg config.Server, logger *slog.Logger) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}

	server := &Server{
		cfg:      cfg,
		logger:   logger,
		registry: NewRegistry(),
		pending:  make(map[string]*pendingRequest),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", server.healthHandler)
	mux.HandleFunc("/_agent/connect", server.agentConnHandler)
	mux.HandleFunc("/_agent/poll", server.pollHandler)
	mux.HandleFunc("/_agent/respond", server.respondHandler)
	mux.HandleFunc("/", server.publicHandler)

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           logRequests(logger, mux),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

type agentConnectResponse struct {
	ID        string `json:"id"`
	Subdomain string `json:"subdomain"`
	PublicURL string `json:"public_url"`
}

type agentConnectRequest struct {
	TargetURL string `json:"target_url"`
}

func (s *Server) agentConnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req agentConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.TargetURL = strings.TrimSpace(req.TargetURL)
	if req.TargetURL == "" {
		http.Error(w, "target_url is required", http.StatusBadRequest)
		return
	}

	subdomain := randomSubdomain()
	publicURL := s.publicURLFor(subdomain)

	tunnel, err := s.registry.Register(subdomain, req.TargetURL, publicURL)
	if errors.Is(err, errSubdomainTaken) {
		http.Error(w, "subdomain already taken", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "register tunnel failed", http.StatusInternalServerError)
		return
	}

	s.logger.Info("tunnel registered",
		"tunnel_id", tunnel.ID,
		"subdomain", tunnel.Subdomain,
		"target_url", tunnel.TargetURL,
	)

	writeJSON(w, http.StatusOK, agentConnectResponse{
		ID:        tunnel.ID,
		Subdomain: tunnel.Subdomain,
		PublicURL: tunnel.PublicURL,
	})
}

func (s *Server) publicURLFor(subdomain string) string {
	publicURL := strings.TrimRight(s.cfg.PublicURL, "/")
	if s.cfg.BaseDomain == "localhost" {
		return publicURL + "/t/" + subdomain
	}

	scheme := "https"
	if strings.HasPrefix(publicURL, "http://") {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s.%s", scheme, subdomain, s.cfg.BaseDomain)
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
