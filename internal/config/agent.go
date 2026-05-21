package config

import "github.com/joho/godotenv"

type Agent struct {
	ServerURL string
}

func LoadAgent() Agent {
	_ = godotenv.Load()

	return Agent{
		ServerURL: envString("TUNNEL_SERVER_URL", "http://localhost:8080"),
	}
}
