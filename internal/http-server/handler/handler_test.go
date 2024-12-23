package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/instinctG/lru-cache/internal/http-server/handler"
	"github.com/instinctG/lru-cache/internal/logger"
	"github.com/instinctG/lru-cache/internal/lru"
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
	tests := []struct {
		name          string
		body          models.PutRequest
		mockReturnErr error
		expectedCode  int
	}{
		{
			name: "Successful put",
			body: models.PutRequest{
				Key:        "test-key",
				Value:      "test-value",
				TTLSeconds: 60,
			},
			mockReturnErr: nil,
			expectedCode:  http.StatusCreated,
		},
		{
			name: "Cache error during put",
			body: models.PutRequest{
				Key:        "test-key",
				Value:      "test-value",
				TTLSeconds: 60,
			},
			mockReturnErr: fmt.Errorf("cache error"),
			expectedCode:  http.StatusInternalServerError,
		},
		{
			name:          "Empty body",
			body:          models.PutRequest{},
			mockReturnErr: fmt.Errorf("invalid body"),
			expectedCode:  http.StatusBadRequest,
		},
		{
			name: "Empty Key",
			body: models.PutRequest{
				Value: "123",
			},
			mockReturnErr: fmt.Errorf("invalid body"),
			expectedCode:  http.StatusBadRequest,
		},
		{
			name: "Invalid TTL seconds",
			body: models.PutRequest{
				Key:        "test-key",
				Value:      "test-value",
				TTLSeconds: -10,
			},
			mockReturnErr: nil,
			expectedCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			discardLogger := logger.NewDiscardLogger()

			h := &handler.Handler{
				LRU: mockCache,
				Log: discardLogger,
			}
			router := setupRouter(h)

			mockCache.On("Put", mock.Anything, tt.body.Key, tt.body.Value, time.Duration(tt.body.TTLSeconds)*time.Second).
				Return(tt.mockReturnErr)

			requestBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/lru", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name              string
		key               string
		mockReturnExpires time.Time
		mockReturnVal     interface{}
		mockReturnErr     error
		expectedCode      int
		expectedResp      models.LRUResponse
	}{
		{
			name:          "Successful get",
			key:           "test-key",
			mockReturnVal: "test-value",
			mockReturnErr: nil,
			expectedCode:  http.StatusOK,
			expectedResp: models.LRUResponse{
				Key:   "test-key",
				Value: "test-value",
			},
		},
		{
			name:          "Key not found",
			key:           "",
			mockReturnVal: "10",
			mockReturnErr: fmt.Errorf("key not found"),
			expectedCode:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			discardLogger := logger.NewDiscardLogger()

			h := &handler.Handler{
				LRU: mockCache,
				Log: discardLogger,
			}
			router := setupRouter(h)

			mockCache.On("Get", mock.Anything, tt.key).
				Return(tt.mockReturnVal, time.Unix(0, 0), tt.mockReturnErr)

			req := httptest.NewRequest(http.MethodGet, "/api/lru/"+tt.key, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			fmt.Println("Response Body:", rec.Body.String())
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusOK {
				var resp models.LRUResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestGetAllHandler(t *testing.T) {
	tests := []struct {
		name          string
		mockKeys      []string
		mockValues    []interface{}
		mockReturnErr error
		expectedCode  int
		expectedResp  models.GetLRU
	}{
		{
			name:          "Successful get all",
			mockKeys:      []string{"key1", "key2"},
			mockValues:    []interface{}{"value1", "value2"},
			mockReturnErr: nil,
			expectedCode:  http.StatusOK,
			expectedResp: models.GetLRU{
				Keys:   []string{"key1", "key2"},
				Values: []interface{}{"value1", "value2"},
			},
		},
		{
			name:          "Empty cache",
			mockKeys:      []string{},
			mockValues:    []interface{}{},
			mockReturnErr: errors.New("no elements found"),
			expectedCode:  http.StatusNoContent,
			expectedResp: models.GetLRU{
				Keys:   []string{},
				Values: []interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			discardLogger := logger.NewDiscardLogger()

			h := &handler.Handler{
				LRU: mockCache,
				Log: discardLogger,
			}
			router := setupRouter(h)

			mockCache.On("GetAll", mock.Anything).Return(tt.mockKeys, tt.mockValues, tt.mockReturnErr)

			req := httptest.NewRequest(http.MethodGet, "/api/lru", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusOK {
				var resp models.GetLRU
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedResp.Keys, resp.Keys)
				assert.ElementsMatch(t, tt.expectedResp.Values, resp.Values)
			}
		})
	}
}

func TestEvictHandler(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		mockReturnErr error
		expectedCode  int
	}{
		{
			name:          "Successful evict",
			key:           "test-key",
			mockReturnErr: nil,
			expectedCode:  http.StatusNoContent,
		},
		{
			name:          "Key not found for evict",
			key:           "test-key",
			mockReturnErr: lru.ErrKeyNotFound,
			expectedCode:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			discardLogger := logger.NewDiscardLogger()

			h := &handler.Handler{
				LRU: mockCache,
				Log: discardLogger,
			}
			router := setupRouter(h)

			mockCache.On("Evict", mock.Anything, tt.key).Return(nil, tt.mockReturnErr)

			req := httptest.NewRequest(http.MethodDelete, "/api/lru/"+tt.key, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestEvictAllHandler(t *testing.T) {
	tests := []struct {
		name          string
		mockReturnErr error
		expectedCode  int
	}{
		{
			name:          "Successful evict all",
			mockReturnErr: nil,
			expectedCode:  http.StatusNoContent,
		},
		{
			name:          "Cache error during evict all",
			mockReturnErr: fmt.Errorf("cache error"),
			expectedCode:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCache)
			discardLogger := logger.NewDiscardLogger()

			h := &handler.Handler{
				LRU: mockCache,
				Log: discardLogger,
			}
			router := setupRouter(h)

			mockCache.On("EvictAll", mock.Anything).Return(tt.mockReturnErr)

			req := httptest.NewRequest(http.MethodDelete, "/api/lru", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
