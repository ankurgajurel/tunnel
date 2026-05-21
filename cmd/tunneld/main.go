package main

import (
	"log/slog"
	"os"

	"github.com/ankurgajurel/tunnel/internal/config"
	"github.com/ankurgajurel/tunnel/internal/server"
)

func main() {
	cfg := config.LoadServer()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	srv := server.New(cfg, logger)

	logger.Info("server is listening",
		"addr", cfg.HTTPAddr,
		"base_domain", cfg.BaseDomain,
		"public_url", cfg.PublicURL,
	)

	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("server error", "error", err)
	}
}
