package logger

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"

	"github.com/fatih/color"
)

const layout = "[2006/01/02 15:04:05]"

// PrettyHandlerOptions содержит параметры для PrettyHandler.
type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

// PrettyHandler форматирует и выводит логи с цветом и структурированным выводом.
type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

// NewPrettyHandler создает новый PrettyHandler с настройками.
func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts), // Использует JSONHandler для внутренней обработки.
		l:       stdLog.New(out, "", 0),                  // Логгер для печати в стандартный вывод.
	}

	return h
}

// Handle обрабатывает запись журнала и выводит ее с цветами и форматированием.
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":" // Устанавливает уровень логирования.

	// Применяет цвета к уровню логирования.
	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	// Собирает атрибуты записи.
	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	// Добавляет дополнительные атрибуты.
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	// Форматирует атрибуты в JSON.
	var b []byte
	var err error
	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	// Печатает лог с цветом, временем и аттрибутами.
	timeStr := r.Time.Format(layout)
	msg := color.CyanString(r.Message)
	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

// WithAttrs добавляет дополнительные атрибуты в лог.
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs, // Добавляет новые атрибуты.
	}
}

// WithGroup задает группу для обработки логов.
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name), // Устанавливает группу для логирования.
		l:       h.l,
	}
}
