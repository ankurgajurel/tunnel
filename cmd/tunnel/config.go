package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/ankurgajurel/tunnel/internal/config"
)

func runLogin(args []string) error {
	fs := newFlagSet("login")
	serverURL := fs.String("server-url", "", "tunneld server url")
	token := fs.String("token", "", "tunnel token")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("usage: tunnel login [--server-url url] [--token token]")
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: tunnel login [--server-url url] [--token token]")
	}

	reader := bufio.NewReader(os.Stdin)
	if strings.TrimSpace(*serverURL) == "" {
		*serverURL = prompt(reader, "enter your server url: ")
	}
	if strings.TrimSpace(*token) == "" {
		*token = prompt(reader, "enter your tunnel token: ")
	}

	cleanURL, err := cleanServerURL(*serverURL)
	if err != nil {
		return err
	}

	cfg := config.Agent{
		ServerURL: cleanURL,
		Token:     strings.TrimSpace(*token),
	}
	if cfg.Token == "" {
		return errors.New("token is required")
	}
	if err := config.SaveAgent(cfg); err != nil {
		return fmt.Errorf("save config failed: %w", err)
	}

	path, _ := config.AgentConfigPath()
	fmt.Println("saved config", path)
	return nil
}

func runConfig(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: tunnel config")
	}

	cfg := config.LoadAgent()
	path, err := config.AgentConfigPath()
	if err != nil {
		return err
	}

	serverURL, err := cleanServerURL(cfg.ServerURL)
	if err != nil {
		serverURL = cfg.ServerURL
	}

	fmt.Println("config file", path)
	fmt.Println("server URL", serverURL)
	if strings.TrimSpace(cfg.Token) == "" {
		fmt.Println("token not set")
	} else {
		fmt.Println("token set")
	}
	return nil
}

func runLogout(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: tunnel logout")
	}

	path, err := config.AgentConfigPath()
	if err != nil {
		return err
	}
	if err := config.DeleteAgent(); err != nil {
		return fmt.Errorf("remove config failed: %w", err)
	}

	fmt.Println("removed config", path)
	return nil
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func cleanServerURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("server url is required")
	}
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("invalid server url %q", value)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("server url must start with http:// or https://")
	}
	if strings.Trim(parsed.Path, "/") != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("server url must not include a path, query, or fragment")
	}

	parsed.Path = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
