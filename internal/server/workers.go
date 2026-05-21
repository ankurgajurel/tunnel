package server

import (
	"net/http"

	"nhooyr.io/websocket"
)

func (s *Server) workHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}

	tunnelID := r.URL.Query().Get("tunnel_id")
	tunnel, ok := s.registry.GetByID(tunnelID)
	if !ok {
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("accept worker websocket", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	s.logger.Info("worker connected",
		"tunnel_id", tunnel.ID,
		"subdomain", tunnel.Subdomain,
	)

	for {
		if _, _, err := conn.Read(r.Context()); err != nil {
			s.logger.Info("worker disconnected",
				"tunnel_id", tunnel.ID,
				"subdomain", tunnel.Subdomain,
				"error", err,
			)
			return
		}
	}
}
