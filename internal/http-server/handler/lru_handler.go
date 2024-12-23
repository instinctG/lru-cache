// Package handler предоставляет обработчики для работы с LRU-кэшем через HTTP API.
package handler

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	sl "github.com/instinctG/lru-cache/internal/logger"
	"github.com/instinctG/lru-cache/internal/lru"
	"github.com/instinctG/lru-cache/internal/models"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ILRUCache интерфейс для взаимодействия с LRU-кэшем.
type ILRUCache interface {
	// Put добавляет или обновляет элемент в кэше.
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get возвращает значение и время истечения элемента по ключу.
	Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error)
	// GetAll возвращает все ключи и значения кэша.
	GetAll(ctx context.Context) (keys []string, values []interface{}, err error)
	// Evict удаляет элемент из кэша по ключу.
	Evict(ctx context.Context, key string) (value interface{}, err error)
	// EvictAll удаляет все элементы из кэша.
	EvictAll(ctx context.Context) error
}

// Put обрабатывает запрос на добавление элемента в кэш.
func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	var req models.PutRequest

	// Декодируем тело запроса.
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {

		h.Log.Error("request body is empty")

		render.JSON(w, r, Response{Error: err.Error()})

		return
	}

	h.Log.Info("request body decoded", slog.Any("request", req))

	// Проверяем валидность данных.
	if err = validator.New().Struct(req); err != nil {

		validateErr := err.(validator.ValidationErrors)

		h.Log.Error("invalid request", sl.Err(validateErr))

		jsonRespond(w, r, http.StatusBadRequest, ValidationError(validateErr))

		return
	}

	// Добавляем элемент в кэш.
	err = h.LRU.Put(r.Context(), req.Key, req.Value, time.Duration(req.TTLSeconds)*time.Second)
	if err != nil {

		h.Log.Error("failed to put lru cache", sl.Err(err))

		jsonRespond(w, r, http.StatusInternalServerError, Response{Error: err.Error()})

		return
	}

	h.Log.Info("cache added successfully")

	jsonRespond(w, r, http.StatusCreated, Response{Message: "cache added"})
}

// Get обрабатывает запрос на получение элемента из кэша.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {

		h.Log.Error("key is empty")

		jsonRespond(w, r, http.StatusBadRequest, Response{Error: "key is empty"})

		return
	}

	val, exp, err := h.LRU.Get(r.Context(), key)
	if errors.Is(err, lru.ErrKeyNotFound) {

		h.Log.Error("key not found", sl.Err(err))

		jsonRespond(w, r, http.StatusNotFound, Response{Error: "key not found"})

		return
	}

	resp := models.LRUResponse{
		Key:       key,
		Value:     val,
		ExpiresAt: exp.Unix(),
	}

	jsonRespond(w, r, http.StatusOK, resp)
}

// GetAll обрабатывает запрос на получение всех элементов из кэша.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	keys, vals, err := h.LRU.GetAll(r.Context())
	if err != nil {

		h.Log.Error("failed to get lru cache", sl.Err(err))

		jsonRespond(w, r, http.StatusNoContent, nil)

		return
	}

	resp := models.GetLRU{
		Keys:   keys,
		Values: vals,
	}

	jsonRespond(w, r, http.StatusOK, resp)
}

// Evict обрабатывает запрос на удаление элемента из кэша по ключу.
func (h *Handler) Evict(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {

		h.Log.Error("key is empty")

		jsonRespond(w, r, http.StatusBadRequest, Response{Error: "key is empty"})

		return
	}

	_, err := h.LRU.Evict(r.Context(), key)
	if errors.Is(err, lru.ErrKeyNotFound) {

		h.Log.Error("key not found", sl.Err(err))

		jsonRespond(w, r, http.StatusNotFound, Response{Error: "key not found"})

		return
	}

	h.Log.Info("key is evicted")

	jsonRespond(w, r, http.StatusNoContent, nil)
}

// EvictAll обрабатывает запрос на удаление всех элементов из кэша.
func (h *Handler) EvictAll(w http.ResponseWriter, r *http.Request) {

	if err := h.LRU.EvictAll(r.Context()); err != nil {

		h.Log.Error("failed to evict lru cache", sl.Err(err))

		jsonRespond(w, r, http.StatusInternalServerError, Response{Error: err.Error()})

		return
	}

	h.Log.Info("all cache evicted successfully")

	jsonRespond(w, r, http.StatusNoContent, nil)
}
