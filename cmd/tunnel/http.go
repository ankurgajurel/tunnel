package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ankurgajurel/tunnel/internal/config"
	"github.com/ankurgajurel/tunnel/internal/protocol"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type agentConnectResponse struct {
	ID        string `json:"id"`
	Subdomain string `json:"subdomain"`
	PublicURL string `json:"public_url"`
}

type agentConnectRequest struct {
	TargetURL string `json:"target_url"`
}

type httpOptions struct {
	port      int
	workers   int
	serverURL string
	token     string
}

func runHTTP(args []string) error {
	opts, err := parseHTTPArgs(args, config.LoadAgent())
	if err != nil {
		return err
	}

	fmt.Println("server URL", opts.serverURL)

	resp, err := http.Get(opts.serverURL + "/healthz")
	if err != nil {
		return fmt.Errorf("tunneld not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tunnel health check failed: %s", resp.Status)
	}

	fmt.Println("tunneld is reachable")

	addr := fmt.Sprintf("127.0.0.1:%d", opts.port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("local target is not reachable")
	}
	defer conn.Close()

	targetURL := "http://" + addr

	connectResp, err := postJSON(http.DefaultClient, opts.serverURL+"/_agent/connect", opts.token, agentConnectRequest{
		TargetURL: targetURL,
	})
	if err != nil {
		return fmt.Errorf("agent connect failed: %w", err)
	}
	defer connectResp.Body.Close()

	if connectResp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent connect failed: %s", connectResp.Status)
	}

	var payload agentConnectResponse
	if err := json.NewDecoder(connectResp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode agent connect response: %w", err)
	}

	fmt.Println("tunnel ID", payload.ID)
	fmt.Println("public URL", payload.PublicURL)
	fmt.Println("exposing local target", targetURL)
	fmt.Println("websocket workers", opts.workers)
	fmt.Println("waiting for requests")

	for i := 1; i <= opts.workers; i++ {
		go runWorker(i, opts.serverURL, opts.token, payload.ID, targetURL)
	}
	select {}
}

func parseHTTPArgs(args []string, cfg config.Agent) (httpOptions, error) {
	if len(args) == 0 {
		return httpOptions{}, fmt.Errorf("usage: tunnel http <port> [--workers n] [--server-url url] [--token token]")
	}

	port, err := strconv.Atoi(args[0])
	if err != nil {
		return httpOptions{}, fmt.Errorf("port must be a number")
	}
	if port < 1 || port > 65535 {
		return httpOptions{}, fmt.Errorf("port must be between 1 and 65535")
	}

	fs := newFlagSet("http")
	workers := fs.Int("workers", 4, "number of websocket workers")
	serverURL := fs.String("server-url", "", "tunneld server url")
	token := fs.String("token", "", "tunnel token")
	if err := fs.Parse(args[1:]); err != nil {
		return httpOptions{}, fmt.Errorf("usage: tunnel http <port> [--workers n] [--server-url url] [--token token]")
	}
	if fs.NArg() != 0 {
		return httpOptions{}, fmt.Errorf("usage: tunnel http <port> [--workers n] [--server-url url] [--token token]")
	}
	if *workers < 1 || *workers > 16 {
		return httpOptions{}, fmt.Errorf("workers must be between 1 and 16")
	}

	if strings.TrimSpace(*serverURL) != "" {
		cfg.ServerURL = *serverURL
	}
	if strings.TrimSpace(*token) != "" {
		cfg.Token = *token
	}

	cleanURL, err := cleanServerURL(cfg.ServerURL)
	if err != nil {
		return httpOptions{}, err
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return httpOptions{}, fmt.Errorf("token is required")
	}

	return httpOptions{
		port:      port,
		workers:   *workers,
		serverURL: cleanURL,
		token:     strings.TrimSpace(cfg.Token),
	}, nil
}

func runWorker(id int, serverURL string, token string, tunnelID string, targetURL string) {
	for {
		if err := connectWorker(id, serverURL, token, tunnelID, targetURL); err != nil {
			fmt.Println("websocket worker disconnected", "worker", id, "error", err)
			time.Sleep(time.Second)
		}
	}
}

func connectWorker(id int, serverURL string, token string, tunnelID string, targetURL string) error {
	workURL, err := workerURL(serverURL, tunnelID)
	if err != nil {
		return err
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, workURL, &websocket.DialOptions{
		HTTPHeader: header,
	})
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	fmt.Println("websocket worker connected", "worker", id)
	for {
		var req protocol.Request
		if err := wsjson.Read(ctx, conn, &req); err != nil {
			return err
		}

		resp := forwardLocal(http.DefaultClient, targetURL, req)
		if err := wsjson.Write(ctx, conn, resp); err != nil {
			return err
		}
	}
}

func workerURL(serverURL string, tunnelID string) (string, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = "/_agent/work"
	u.RawQuery = "tunnel_id=" + url.QueryEscape(tunnelID)

	return u.String(), nil
}

func forwardLocal(client *http.Client, targetURL string, req protocol.Request) protocol.Response {
	localReq, err := http.NewRequest(req.Method, targetURL+req.Path, bytes.NewReader(req.Body))
	if err != nil {
		return protocol.Response{ID: req.ID, Error: "build local request failed"}
	}

	localReq.Header = req.Header.Clone()

	resp, err := client.Do(localReq)
	if err != nil {
		return protocol.Response{ID: req.ID, Error: "local target is unreachable"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return protocol.Response{ID: req.ID, Error: "read local response failed"}
	}

	return protocol.Response{
		ID:     req.ID,
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   body,
	}
}

func postJSON(client *http.Client, url string, token string, value any) (*http.Response, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return client.Do(req)
}
