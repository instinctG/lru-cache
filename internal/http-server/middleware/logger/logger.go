// Package logger предоставляет функциональность для логирования HTTP-запросов.
package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// New создает middleware для логирования HTTP-запросов.
// Логирует метод, путь, удаленный адрес, User-Agent, идентификатор запроса и время выполнения запроса.
// Параметры:
//   - log: объект логгера для записи информации о запросах.
//
// Возвращает функцию middleware, которая логирует информацию о запросах и передает управление следующему обработчику.
func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		log := log.With(
			slog.String("component", "middleware/logger"),
		)
		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			// Создание записи лога с данными запроса
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			// Обертка для ResponseWriter для получения статуса и количества записанных байт
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				// Логируем завершение обработки запроса
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
