// Package logger реализует DiscardLogger для тестов, чтобы игнорировать логи
// и не вызывать панику.  В этом пакете методы это просто пустые оболочки, чтобы удовлетворить handler логгера
package logger

import (
	"context"
	"log/slog"
)

// NewDiscardLogger создает новый логгер, который игнорирует все логи.
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler()) // Возвращает логгер с обработчиком, который не записывает логи.
}

// DiscardHandler - пустая структура
type DiscardHandler struct{}

// NewDiscardHandler создает новый обработчик, который игнорирует все записи.
func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

// Handle игнорирует запись журнала, ничего не делает.
func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs возвращает тот же обработчик без изменения атрибутов.
func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup возвращает тот же обработчик без изменения группы.
func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	return h
}

// Enabled всегда возвращает false, так как записи игнорируются.
func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}
