package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Server struct {
	HTTPAddr string
}

func LoadServer() Server {
	_ = godotenv.Load()

	return Server{
		HTTPAddr: envString("TUNNEL_HTTP_ADDR", ":8080"),
	}
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
