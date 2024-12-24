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

func jsonRespond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	render.JSON(w, r, data)
}

// ValidationError формирует сообщение об ошибках валидации.
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field and not valid", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Error: strings.Join(errMsgs, ", "),
	}
}

func isSimpleType(value interface{}) bool {
	switch value.(type) {
	case string, float64, int, bool:
		return true
	default:
		return false
	}
}
