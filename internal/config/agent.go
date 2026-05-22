package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Agent struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
}

func LoadAgent() Agent {
	_ = godotenv.Load()
	cfg := Agent{
		ServerURL: "http://localhost:8080",
		Token:     "dev-token",
	}

	saved, err := readAgentFile()
	if err == nil {
		if strings.TrimSpace(saved.ServerURL) != "" {
			cfg.ServerURL = saved.ServerURL
		}
		if strings.TrimSpace(saved.Token) != "" {
			cfg.Token = saved.Token
		}
	}

	cfg.ServerURL = envString("TUNNEL_SERVER_URL", cfg.ServerURL)
	cfg.Token = envString("TUNNEL_TOKEN", cfg.Token)

	return cfg
}

func SaveAgent(cfg Agent) error {
	path, err := AgentConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return os.WriteFile(path, append(body, '\n'), 0o600)
}

func DeleteAgent() error {
	path, err := AgentConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func AgentConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home dir: %w", err)
	}

	return filepath.Join(home, ".tunneld", "config.json"), nil
}

func readAgentFile() (Agent, error) {
	path, err := AgentConfigPath()
	if err != nil {
		return Agent{}, err
	}

	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Agent{}, err
	}
	if err != nil {
		return Agent{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Agent
	if err := json.Unmarshal(body, &cfg); err != nil {
		return Agent{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
