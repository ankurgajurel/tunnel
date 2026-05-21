package server

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ankurgajurel/tunnel/internal/protocol"
)

const maxBodyBytes = 10 << 20

type pendingRequest struct {
	request protocol.Request
	reply   chan protocol.Response
}

type pollRequest struct {
	TunnelID string `json:"tunnel_id"`
}

func (s *Server) publicHandler(w http.ResponseWriter, r *http.Request) {
	tunnel, path, ok := s.findTunnel(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if s.shouldWarn(r) {
		s.renderWarning(w, r)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	id := "req_" + time.Now().Format("20060102150405.000000000")
	pending := &pendingRequest{
		request: protocol.Request{
			ID:     id,
			Method: r.Method,
			Path:   path,
			Header: cleanHeader(r.Header),
			Body:   body,
		},
		reply: make(chan protocol.Response, 1),
	}

	s.remember(pending)
	defer s.forget(id)

	select {
	case tunnel.Requests <- pending:
	case <-time.After(5 * time.Second):
		http.Error(w, "tunnel is busy", http.StatusServiceUnavailable)
		return
	case <-r.Context().Done():
		return
	}

	select {
	case resp := <-pending.reply:
		writeProxyResponse(w, resp)
	case <-time.After(60 * time.Second):
		http.Error(w, "tunnel response timed out", http.StatusGatewayTimeout)
	case <-r.Context().Done():
		return
	}
}

func (s *Server) pollHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tunnel, ok := s.registry.GetByID(req.TunnelID)
	if !ok {
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	select {
	case pending := <-tunnel.Requests:
		writeJSON(w, http.StatusOK, pending.request)
	case <-time.After(30 * time.Second):
		w.WriteHeader(http.StatusNoContent)
	case <-r.Context().Done():
		return
	}
}

func (s *Server) respondHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp protocol.Response
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	pending, ok := s.take(resp.ID)
	if !ok {
		http.Error(w, "request not found", http.StatusNotFound)
		return
	}

	pending.reply <- resp
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) findTunnel(r *http.Request) (*Tunnel, string, bool) {
	if strings.HasPrefix(r.URL.Path, "/t/") {
		rest := strings.TrimPrefix(r.URL.Path, "/t/")
		subdomain, path, _ := strings.Cut(rest, "/")
		if path == "" {
			path = "/"
		} else {
			path = "/" + path
		}
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}

		tunnel, ok := s.registry.Get(subdomain)
		return tunnel, path, ok
	}

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	suffix := "." + s.cfg.BaseDomain
	if !strings.HasSuffix(host, suffix) {
		return nil, "", false
	}

	subdomain := strings.TrimSuffix(host, suffix)
	tunnel, ok := s.registry.Get(subdomain)
	return tunnel, r.URL.RequestURI(), ok
}

func writeProxyResponse(w http.ResponseWriter, resp protocol.Response) {
	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusBadGateway)
		return
	}

	for key, values := range cleanHeader(resp.Header) {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	if resp.Status == 0 {
		resp.Status = http.StatusOK
	}

	w.WriteHeader(resp.Status)
	_, _ = w.Write(resp.Body)
}

func cleanHeader(header http.Header) http.Header {
	clean := header.Clone()
	for _, key := range []string{"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		clean.Del(key)
	}
	return clean
}

func (s *Server) remember(req *pendingRequest) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	s.pending[req.request.ID] = req
}

func (s *Server) forget(id string) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	delete(s.pending, id)
}

func (s *Server) take(id string) (*pendingRequest, bool) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	req, ok := s.pending[id]
	if ok {
		delete(s.pending, id)
	}
	return req, ok
}
