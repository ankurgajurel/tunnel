package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: tunnel <command>")
		return
	}

	command := os.Args[1]

	switch command {
	case "login":
		runLogin()
	case "http":
		runHTTP()
	default:
		fmt.Println("unknown command", command)
	}
}

func runLogin() {
	reader := bufio.NewReader(os.Stdin)

	serverURL := prompt(reader, "enter your server url: ")
	token := prompt(reader, "enter your tunnel token: ")

	cfg := config.Agent{
		ServerURL: strings.TrimRight(serverURL, "/"),
		Token:     token,
	}
	if err := config.SaveAgent(cfg); err != nil {
		fmt.Println("save config failed:", err)
		return
	}

	path, _ := config.AgentConfigPath()
	fmt.Println("saved config", path)
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func runHTTP() {
	cfg := config.LoadAgent()
	serverURL := strings.TrimRight(cfg.ServerURL, "/")

	fmt.Println("server URL", serverURL)

	resp, err := http.Get(serverURL + "/healthz")
	if err != nil {
		fmt.Println("tunneld not reachable: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("tunnel health check failed: ", resp.Status)
		return
	}

	fmt.Println("tunneld is reachable")

	if len(os.Args) < 3 {
		fmt.Println("usage: tunnel http <port>")
		return
	}

	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("port must be a number")
		return
	}

	if port < 1 || port > 65535 {
		fmt.Println("port must be between 1 and 65535")
		return
	}

	workers, err := parseWorkers(os.Args[3:])
	if err != nil {
		fmt.Println(err)
		return
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Println("local target is not reachable")
		return
	}
	defer conn.Close()

	targetURL := "http://" + addr

	connectResp, err := postJSON(http.DefaultClient, serverURL+"/_agent/connect", cfg.Token, agentConnectRequest{
		TargetURL: targetURL,
	})
	if err != nil {
		fmt.Println("agent connect failed:", err)
		return
	}
	defer connectResp.Body.Close()

	if connectResp.StatusCode != http.StatusOK {
		fmt.Println("agent connect failed:", connectResp.Status)
		return
	}

	var payload agentConnectResponse
	if err := json.NewDecoder(connectResp.Body).Decode(&payload); err != nil {
		fmt.Println("decode agent connect response:", err)
		return
	}

	fmt.Println("tunnel ID", payload.ID)
	fmt.Println("public URL", payload.PublicURL)
	fmt.Println("exposing local target", targetURL)
	fmt.Println("websocket workers", workers)
	fmt.Println("waiting for requests")

	for i := 1; i <= workers; i++ {
		go runWorker(i, serverURL, cfg.Token, payload.ID, targetURL)
	}
	select {}
}

func parseWorkers(args []string) (int, error) {
	if len(args) == 0 {
		return 4, nil
	}
	if len(args) != 2 || args[0] != "--workers" {
		return 0, fmt.Errorf("usage: tunnel http <port> [--workers n]")
	}

	workers, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, fmt.Errorf("workers must be a number")
	}
	if workers < 1 || workers > 16 {
		return 0, fmt.Errorf("workers must be between 1 and 16")
	}

	return workers, nil
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
