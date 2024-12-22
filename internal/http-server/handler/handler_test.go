package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/instinctG/lru-cache/internal/http-server/handler"
	"github.com/instinctG/lru-cache/internal/logger"
	"github.com/instinctG/lru-cache/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockCache реализует интерфейс ILRUCache для моков.
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Get(ctx context.Context, key string) (interface{}, time.Time, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockCache) GetAll(ctx context.Context) ([]string, []interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Get(1).([]interface{}), args.Error(2)
}

func (m *MockCache) Evict(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) EvictAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/lru", h.Put)
	r.Get("/api/lru/{key}", h.Get)
	r.Get("/api/lru", h.GetAll)
	r.Delete("/api/lru/{key}", h.Evict)
	r.Delete("/api/lru", h.EvictAll)
	return r
}

func TestPutHandler(t *testing.T) {
	mockCache := new(MockCache)

	// Используем DiscardLogger для игнорирования логов в тестах
	discardLogger := logger.NewDiscardLogger()

	h := &handler.Handler{
		LRU: mockCache,
		Log: discardLogger, // Передаем логгер в Handler
	}
	router := setupRouter(h)

	body := models.PutRequest{
		Key:        "test-key",
		Value:      "test-value",
		TTLSeconds: 60,
	}

	mockCache.On("Put", mock.Anything, "test-key", "test-value", time.Minute).Return(nil)

	requestBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/lru", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockCache.AssertCalled(t, "Put", mock.Anything, "test-key", "test-value", time.Minute)
}

func TestGetHandler(t *testing.T) {
	mockCache := new(MockCache)

	// Используем DiscardLogger для игнорирования логов в тестах
	discardLogger := logger.NewDiscardLogger()

	h := &handler.Handler{
		LRU: mockCache,
		Log: discardLogger, // Передаем логгер в Handler
	}
	router := setupRouter(h)

	mockCache.On("Get", mock.Anything, "test-key").Return("test-value", time.Now().Add(time.Minute), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/lru/test-key", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp models.LRUResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "test-key", resp.Key)
	assert.Equal(t, "test-value", resp.Value)
}

func TestGetAllHandler(t *testing.T) {
	mockCache := new(MockCache)

	// Используем DiscardLogger для игнорирования логов в тестах
	discardLogger := logger.NewDiscardLogger()

	h := &handler.Handler{
		LRU: mockCache,
		Log: discardLogger, // Передаем логгер в Handler
	}
	router := setupRouter(h)

	mockCache.On("GetAll", mock.Anything).Return([]string{"key1", "key2"}, []interface{}{"value1", "value2"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/lru", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp models.GetLRU
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2"}, resp.Keys)
	assert.ElementsMatch(t, []interface{}{"value1", "value2"}, resp.Values)
}

func TestEvictHandler(t *testing.T) {
	mockCache := new(MockCache)

	// Используем DiscardLogger для игнорирования логов в тестах
	discardLogger := logger.NewDiscardLogger()

	h := &handler.Handler{
		LRU: mockCache,
		Log: discardLogger, // Передаем логгер в Handler
	}
	router := setupRouter(h)

	mockCache.On("Evict", mock.Anything, "test-key").Return("test-value", nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/lru/test-key", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockCache.AssertCalled(t, "Evict", mock.Anything, "test-key")
}

func TestEvictAllHandler(t *testing.T) {
	mockCache := new(MockCache)

	// Используем DiscardLogger для игнорирования логов в тестах
	discardLogger := logger.NewDiscardLogger()

	h := &handler.Handler{
		LRU: mockCache,
		Log: discardLogger, // Передаем логгер в Handler
	}
	router := setupRouter(h)

	mockCache.On("EvictAll", mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/lru", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockCache.AssertCalled(t, "EvictAll", mock.Anything)
}
