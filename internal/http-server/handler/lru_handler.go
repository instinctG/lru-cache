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

// ILRUCache интерфейс LRU-кэша. Поддерживает только строковые ключи. Поддерживает только простые типы данных в значениях.
type ILRUCache interface {
	// Put запись данных в кэш
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get получение данных из кэша по ключу
	Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error)
	// GetAll получение всего наполнения кэша в виде двух слайсов: слайса ключей и слайса значений. Пары ключ-значения из кэша располагаются на соответствующих позициях в слайсах.
	GetAll(ctx context.Context) (keys []string, values []interface{}, err error)
	// Evict ручное удаление данных по ключу
	Evict(ctx context.Context, key string) (value interface{}, err error)
	// EvictAll ручная инвалидация всего кэша
	EvictAll(ctx context.Context) error
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	var req models.PutRequest

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		h.Log.Error("request body is empty")

		render.JSON(w, r, Response{Error: err.Error()})
		return
	}

	h.Log.Info("request body decoded ", slog.Any("request", req))

	// TODO : ДОДЕЛАТЬ ВАЛИДАЦИЮ ДАННЫХ
	if err = validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		h.Log.Error("invalid request", sl.Err(validateErr))

		jsonRespond(w, r, http.StatusBadRequest, ValidationError(validateErr))
		return
	}

	err = h.LRU.Put(r.Context(), req.Key, req.Value, time.Duration(req.TTLSeconds)*time.Second)
	if err != nil {
		h.Log.Error("failed to put lru cache", sl.Err(err))

		jsonRespond(w, r, http.StatusInternalServerError, Response{Error: err.Error()})

		return
	}

	h.Log.Info("cache added successfully")

	jsonRespond(w, r, http.StatusCreated, Response{Message: "cache added"})
}

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
		ExpiresAt: exp.Unix(), // Получаем время в секундах с помощью преобразования времени в Unix(отображается в секундах)
	}

	jsonRespond(w, r, http.StatusOK, resp)
}

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

func (h *Handler) EvictAll(w http.ResponseWriter, r *http.Request) {

	if err := h.LRU.EvictAll(r.Context()); err != nil {
		h.Log.Error("failed to evict lru cache", sl.Err(err))

		jsonRespond(w, r, http.StatusInternalServerError, Response{Error: err.Error()})

		return
	}

	h.Log.Info("all cache evicted successfully")
	jsonRespond(w, r, http.StatusNoContent, nil)
}
