package main

import (
	"github.com/instinctG/lru-cache/internal/config"
	transportHTTP "github.com/instinctG/lru-cache/internal/http-server/handler"
	"github.com/instinctG/lru-cache/internal/logger"
	"log/slog"
)

func Run() error {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.LogLevel)

	log.Info("starting lru-cache", slog.String("LOG-LEVEL", cfg.LogLevel))
	log.Debug("debug messages are enabled")

	//TODO: init lru-cache : in-memory lru-cache

	handler := transportHTTP.NewHandler(cfg.Port, log)

	if err := handler.Serve(); err != nil {
		log.Error("failed to start server")
		return err
	}

	log.Info("server started")

	return nil
}

func main() {
	if err := Run(); err != nil {
		slog.Error("could not run the application", logger.Err(err))
	}
}
