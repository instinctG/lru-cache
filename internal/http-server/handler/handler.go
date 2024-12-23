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

type Handler struct {
	LRU    ILRUCache
	Log    *slog.Logger
	Router *chi.Mux
	Server *http.Server
}

func NewHandler(lru ILRUCache, address string, log *slog.Logger) *Handler {
	h := &Handler{
		LRU:    lru,
		Log:    log,
		Router: chi.NewRouter(),
	}

	h.Router.Use(middleware.RequestID)
	//Можно использовать logger от chi, но для более подробного ответа и удобства написал свой
	//h.Router.Use(middleware.Logger)
	h.Router.Use(logger.New(log))
	h.Router.Use(middleware.Recoverer)
	h.Router.Use(middleware.URLFormat)

	h.mapRoutes()

	h.Server = &http.Server{
		Addr:    address,
		Handler: h.Router,
	}

	return h
}

func (h *Handler) mapRoutes() {
	h.Router.Post("/api/lru", h.Put)

	h.Router.Get("/api/lru/{key}", h.Get)
	h.Router.Get("/api/lru", h.GetAll)

	h.Router.Delete("/api/lru/{key}", h.Evict)
	h.Router.Delete("/api/lru", h.EvictAll)
}

func (h *Handler) Serve() error {
	h.Log.Info("starting server on port: " + h.Server.Addr)
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	start := time.Now()

	h.Log.Info("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.Log.Info("Server shutdown complete", slog.String("shutdown_time", time.Since(start).String()), slog.String("signal", sig.String()))
	if err := h.Server.Shutdown(ctx); err != nil {
		h.Log.Error("Server Shutdown:", sl.Err(err))
	}

	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		h.Log.Debug("timeout of 5 seconds.")
	}
	h.Log.Debug("Server exiting")

	return nil
}
