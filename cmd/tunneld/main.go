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

	srv := server.New(cfg.HTTPAddr, logger)

	logger.Info("server is listening", "addr", cfg.HTTPAddr)

	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("server error", "error", err)
	}
}
