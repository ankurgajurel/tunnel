package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ankurgajurel/tunnel/internal/protocol"
)

const maxBodyBytes = 10 << 20

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
	req := protocol.Request{
		ID:     id,
		Method: r.Method,
		Path:   path,
		Header: cleanHeader(r.Header),
		Body:   body,
	}

	s.proxyViaWorker(w, r, tunnel, req)
}

func (s *Server) proxyViaWorker(w http.ResponseWriter, r *http.Request, tunnel *Tunnel, req protocol.Request) {
	select {
	case worker := <-tunnel.Workers:
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		resp, err := worker.RoundTrip(ctx, req)
		if err != nil {
			worker.Close()
			http.Error(w, "worker request failed", http.StatusBadGateway)
			return
		}

		select {
		case tunnel.Workers <- worker:
		default:
			worker.Close()
		}

		writeProxyResponse(w, resp)
		return
	case <-time.After(5 * time.Second):
		http.Error(w, "tunnel is busy", http.StatusServiceUnavailable)
	case <-r.Context().Done():
		return
	}
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
