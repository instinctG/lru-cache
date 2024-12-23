// Package handler содержит функции для обработки ответов и ошибок валидации.
package handler

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

// Response структура для ответа с сообщением или ошибкой.
type Response struct {
	Message string `json:"message,omitempty"` // Сообщение.
	Error   string `json:"error,omitempty"`   // Описание ошибки.
}

// jsonRespond отправляет JSON-ответ с заданным статусом.
func jsonRespond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	render.JSON(w, r, data)
}

// ValidationError формирует сообщение об ошибках валидации.
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	// Формируем сообщения для каждой ошибки валидации.
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field and not valid", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	// Возвращаем ошибку в виде строки.
	return Response{
		Error: strings.Join(errMsgs, ", "),
	}
}
