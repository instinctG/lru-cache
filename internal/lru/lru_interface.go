package lru

import (
	"context"
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
