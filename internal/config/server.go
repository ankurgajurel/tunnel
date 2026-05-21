package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Server struct {
	HTTPAddr   string
	BaseDomain string
	PublicURL  string
}

func LoadServer() Server {
	_ = godotenv.Load()

	return Server{
		HTTPAddr:   envString("TUNNEL_HTTP_ADDR", ":8080"),
		BaseDomain: envString("TUNNEL_BASE_DOMAIN", "localhost"),
		PublicURL:  envString("TUNNEL_PUBLIC_URL", "http://localhost:8080"),
	}
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	return value
}
