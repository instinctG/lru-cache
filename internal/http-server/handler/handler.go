// Package handler предоставляет функциональность для управления HTTP-сервером и обработкой маршрутов.
package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/instinctG/lru-cache/internal/http-server/middleware/logger"
	sl "github.com/instinctG/lru-cache/internal/logger"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Handler представляет структуру для обработки HTTP-запросов и управления сервером.
type Handler struct {
	LRU    ILRUCache    // Интерфейс LRU-кэша для обработки запросов.
	Log    *slog.Logger // Логгер для записи событий сервера.
	Router *chi.Mux     // Роутер для маршрутизации запросов.
	Server *http.Server // HTTP-сервер.
}

// NewHandler создает новый экземпляр Handler.
// Параметры:
//   - lru: реализация интерфейса LRU-кэша.
//   - address: адрес для запуска HTTP-сервера (например, "localhost:8080").
//   - log: логгер для обработки событий.
//
// Возвращает: указатель на созданный Handler.
func NewHandler(lru ILRUCache, address string, log *slog.Logger) *Handler {
	h := &Handler{
		LRU:    lru,
		Log:    log,
		Router: chi.NewRouter(),
	}

	// Настройка middleware
	h.Router.Use(middleware.RequestID) // Генерация идентификаторов запросов.
	h.Router.Use(logger.New(log))      // Логирование запросов.
	h.Router.Use(middleware.Recoverer) // Восстановление после паники.
	h.Router.Use(middleware.URLFormat) // Поддержка форматов URL.

	h.mapRoutes() // Настройка маршрутов.

	h.Server = &http.Server{
		Addr:    address,
		Handler: h.Router,
	}

	return h
}

// mapRoutes настраивает маршруты HTTP-обработчиков.
func (h *Handler) mapRoutes() {
	h.Router.Post("/api/lru", h.Put)           // Добавление или обновление элемента в кэше.
	h.Router.Get("/api/lru/{key}", h.Get)      // Получение элемента по ключу.
	h.Router.Get("/api/lru", h.GetAll)         // Получение всех элементов кэша.
	h.Router.Delete("/api/lru/{key}", h.Evict) // Удаление элемента по ключу.
	h.Router.Delete("/api/lru", h.EvictAll)    // Удаление всех элементов кэша.
}

// Serve запускает HTTP-сервер и обрабатывает сигналы завершения работы.
// Возвращает: ошибку в случае, если сервер не может быть запущен.
func (h *Handler) Serve() error {
	h.Log.Info("starting server on port: " + h.Server.Addr)

	// Запуск сервера в отдельной горутине
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Ожидание сигнала завершения (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	start := time.Now()

	h.Log.Info("Received shutdown signal")

	// Завершение работы сервера с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.Log.Info("Server shutdown complete", slog.String("shutdown_time", time.Since(start).String()), slog.String("signal", sig.String()))
	if err := h.Server.Shutdown(ctx); err != nil {
		h.Log.Error("Server Shutdown:", sl.Err(err))
	}

	// Обработка завершения контекста
	select {
	case <-ctx.Done():
		h.Log.Debug("timeout of 5 seconds.")
	}
	h.Log.Debug("Server exiting")

	return nil
}
