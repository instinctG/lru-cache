package main

import (
	"github.com/instinctG/lru-cache/internal/config"
	"github.com/instinctG/lru-cache/internal/logger"
	"log/slog"
)

func main() {

	//TODO:init config : env
	cfg := config.MustLoad()

	//TODO: init logger : slog
	log := logger.SetupLogger(cfg.LogLevel)

	log.Info("starting lru-cache", slog.String("LOG-LEVEL", cfg.LogLevel))
	log.Debug("debug messages are enabled")

	//TODO: init lru-cache : in-memory lru-cache

	//TODO: init router : chi, "chi render"

	//TODO: run server
}
