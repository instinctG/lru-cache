// Package logger содержит функции для работы с логированием ошибок.
package logger

import (
	"log/slog"
)

// Err создает атрибут лога для ошибки.
func Err(err error) slog.Attr {
	// Возвращает атрибут с ошибкой, который можно использовать в логах.
	return slog.Attr{
		Key:   "error",                       // Ключ атрибута.
		Value: slog.StringValue(err.Error()), // Значение — строковое представление ошибки.
	}
}
