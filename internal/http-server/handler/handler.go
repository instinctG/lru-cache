package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/instinctG/lru-cache/internal/http-server/middleware/logger"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Handler struct {
	log    *slog.Logger
	Router *chi.Mux
	Server *http.Server
}

func NewHandler(address string, log *slog.Logger) *Handler {
	h := &Handler{
		log:    log,
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

}

func (h *Handler) Serve() error {
	h.log.Info("starting server on port: " + h.Server.Addr)
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	start := time.Now()

	h.log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.log.Info("Server shutdown complete", slog.String("shutdown_time", time.Since(start).String()))
	if err := h.Server.Shutdown(ctx); err != nil {
		h.log.Error("Server Shutdown:", err)
	}

	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		h.log.Debug("timeout of 5 seconds.")
	}
	h.log.Debug("Server exiting")

	return nil
}
