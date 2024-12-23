package main

import (
	"github.com/instinctG/lru-cache/internal/config"
	transportHTTP "github.com/instinctG/lru-cache/internal/http-server/handler"
	sl "github.com/instinctG/lru-cache/internal/logger"
	"github.com/instinctG/lru-cache/internal/lru"
	"log/slog"
)

// Run конфигурирует и запускает сервер с LRU-кэшом.
func Run() error {
	// Загружает конфигурацию из файла или переменных окружения.
	cfg := config.MustLoad()

	// Настраивает логгер в зависимости от уровня логирования.
	log := sl.SetupLogger(cfg.LogLevel)

	log.Info("starting lru-cache", slog.String("LOG-LEVEL", cfg.LogLevel))
	log.Debug("debug messages are enabled")

	// Создает новый LRU-кэш с заданным размером и временем жизни по умолчанию.
	LRUCache := lru.NewLRUCache(cfg.CacheSize, cfg.DefaultCacheTTL)

	// Создает HTTP-обработчик с кэшем и логированием.
	handler := transportHTTP.NewHandler(LRUCache, cfg.Port, log)

	// Запускает HTTP-сервер.
	if err := handler.Serve(); err != nil {
		log.Error("failed to start server")
		return err
	}

	log.Info("server started")

	return nil
}

// main точка входа в приложение. Запускает Run и обрабатывает ошибки.
func main() {
	if err := Run(); err != nil {
		slog.Error("could not run the application", sl.Err(err))
	}
}
