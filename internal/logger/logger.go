// Package logger содержит логику настройки логирования для приложения.
package logger

import (
	"log/slog"
	"os"
)

// Уровень логирования: ERROR, DEBUG, INFO
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	ERROR = "ERROR"
)

// SetupLogger настраивает логгер в зависимости от заданного уровня логирования.
func SetupLogger(logLevel string) *slog.Logger {
	var log *slog.Logger

	// Настройка логирования в зависимости от уровня.
	switch logLevel {
	case DEBUG:
		log = setupPrettySlog() // Настройка для отладочного режима.
	case INFO:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case ERROR:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn})) // По умолчанию уровень WARNING.
	}

	return log
}

// setupPrettySlog настраивает логгер с красивым (pretty) выводом для уровня DEBUG.
func setupPrettySlog() *slog.Logger {
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	// Возвращаем логгер с красивым выводом в консоль.
	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
