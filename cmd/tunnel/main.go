package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

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

	fmt.Println("exposing local port", port)
}
