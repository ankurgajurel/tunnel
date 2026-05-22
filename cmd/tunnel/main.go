package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
	case "version":
		fmt.Println(version)
	case "login":
		return runLogin(args[1:])
	case "config":
		return runConfig(args[1:])
	case "logout":
		return runLogout(args[1:])
	case "http":
		return runHTTP(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\nrun: tunnel help", args[0])
	}

	return nil
}

func printUsage() {
	fmt.Println(`tunnel exposes a local http app through a tunneld server

usage:
  tunnel <command> [args]

commands:
  login      save server url and token
  config     show the current cli config
  logout     remove saved cli config
  http       expose a local http port
  version    show cli version
  help       show this help

examples:
  tunnel login
  tunnel config
  tunnel http 5050
  tunnel http 5050 --workers 2 --server-url http://localhost:8080 --token secret`)
}
