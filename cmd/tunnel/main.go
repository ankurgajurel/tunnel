package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ankurgajurel/tunnel/internal/config"
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
	case "http":
		runHTTP()
	default:
		fmt.Println("unknown command", command)
	}
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

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Println("local target is not reachable")
		return
	}
	defer conn.Close()

	targetURL := "http://" + addr

	body, err := json.Marshal(agentConnectRequest{
		TargetURL: targetURL,
	})
	if err != nil {
		fmt.Println("encode agent connect request:", err)
		return
	}

	connectResp, err := http.Post(serverURL+"/_agent/connect", "application/json", bytes.NewReader(body))
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
}
