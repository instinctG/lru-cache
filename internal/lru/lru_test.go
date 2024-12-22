package lru_test

import (
	"context"
	"github.com/instinctG/lru-cache/internal/lru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLRUCache_PutAndGet(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(2)

	// Добавляем элемент в кэш
	err := cache.Put(ctx, "key1", "value1", time.Minute)
	require.NoError(t, err)

	// Проверяем, что элемент можно получить
	value, expiresAt, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)
	assert.WithinDuration(t, time.Now().Add(time.Minute), expiresAt, time.Second)

	// Проверяем, что элемент удаляется после истечения времени
	time.Sleep(time.Second)
	_, _, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, nil)
}

func TestLRUCache_GetKeyNotFound(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(2)

	// Проверяем, что элемент отсутствует в пустом кэше
	_, _, err := cache.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, lru.ErrKeyNotFound)
}

func TestLRUCache_CapacityEviction(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(2)

	// Добавляем элементы в кэш
	require.NoError(t, cache.Put(ctx, "key1", "value1", time.Minute))
	require.NoError(t, cache.Put(ctx, "key2", "value2", time.Minute))

	// Добавляем третий элемент, проверяем, что первый удалился
	require.NoError(t, cache.Put(ctx, "key3", "value3", time.Minute))

	_, _, err := cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, lru.ErrKeyNotFound)

	value, _, err := cache.Get(ctx, "key2")
	require.NoError(t, err)
	assert.Equal(t, "value2", value)
}

func TestLRUCache_GetAll(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(3)

	// Добавляем элементы
	require.NoError(t, cache.Put(ctx, "key1", "value1", time.Minute))
	require.NoError(t, cache.Put(ctx, "key2", "value2", time.Minute))
	require.NoError(t, cache.Put(ctx, "key3", "value3", time.Minute))

	// Получаем все элементы
	keys, values, err := cache.GetAll(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys)
	assert.ElementsMatch(t, []interface{}{"value1", "value2", "value3"}, values)
}

func TestLRUCache_Evict(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(2)

	// Добавляем элемент
	require.NoError(t, cache.Put(ctx, "key1", "value1", time.Minute))

	// Удаляем элемент
	value, err := cache.Evict(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Проверяем, что элемент больше не существует
	_, _, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, lru.ErrKeyNotFound)
}

func TestLRUCache_EvictAll(t *testing.T) {
	ctx := context.Background()
	cache := lru.NewLRUCache(3)

	// Добавляем элементы
	require.NoError(t, cache.Put(ctx, "key1", "value1", time.Minute))
	require.NoError(t, cache.Put(ctx, "key2", "value2", time.Minute))
	require.NoError(t, cache.Put(ctx, "key3", "value3", time.Minute))

	// Удаляем все элементы
	err := cache.EvictAll(ctx)
	require.NoError(t, err)

	// Проверяем, что кэш пуст
	keys, values, err := cache.GetAll(ctx)
	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Nil(t, values)
}
