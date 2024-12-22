package main

import (
	"github.com/instinctG/lru-cache/internal/config"
	transportHTTP "github.com/instinctG/lru-cache/internal/http-server/handler"
	sl "github.com/instinctG/lru-cache/internal/logger"
	"github.com/instinctG/lru-cache/internal/lru"
	"log/slog"
)

func Run() error {
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.LogLevel)

	log.Info("starting lru-cache", slog.String("LOG-LEVEL", cfg.LogLevel))
	log.Debug("debug messages are enabled")

	LRUCache := lru.NewLRUCache(cfg.CacheSize)

	handler := transportHTTP.NewHandler(LRUCache, cfg.Port, log)

	if err := handler.Serve(); err != nil {
		log.Error("failed to start server")
		return err
	}

	log.Info("server started")

	return nil
}

func main() {
	if err := Run(); err != nil {
		slog.Error("could not run the application", sl.Err(err))
	}
}
