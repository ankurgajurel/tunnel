package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ankurgajurel/tunnel/internal/protocol"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Worker struct {
	conn *websocket.Conn
}

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

	worker := &Worker{conn: conn}
	s.logger.Info("worker connected",
		"tunnel_id", tunnel.ID,
		"subdomain", tunnel.Subdomain,
	)

	select {
	case tunnel.Workers <- worker:
	case <-r.Context().Done():
		return
	}

	<-r.Context().Done()
	s.logger.Info("worker disconnected",
		"tunnel_id", tunnel.ID,
		"subdomain", tunnel.Subdomain,
		"error", r.Context().Err(),
	)
}

func (w *Worker) RoundTrip(ctx context.Context, req protocol.Request) (protocol.Response, error) {
	if err := wsjson.Write(ctx, w.conn, req); err != nil {
		return protocol.Response{}, fmt.Errorf("write request: %w", err)
	}

	var resp protocol.Response
	if err := wsjson.Read(ctx, w.conn, &resp); err != nil {
		return protocol.Response{}, fmt.Errorf("read response: %w", err)
	}

	return resp, nil
}

func (w *Worker) Close() {
	w.conn.Close(websocket.StatusNormalClosure, "")
}
