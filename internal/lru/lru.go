// Package lru реализует кэш с вытеснением по принципу LRU (Least Recently Used) и использованием TTL.
package lru

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrKeyNotFound возвращается, если ключ не найден в кэше.
var (
	ErrCacheIsEmpty = errors.New("cache is empty") // ErrCacheIsEmpty возвращается, если кэш пустой.
	ErrKeyNotFound  = errors.New("key not found")  // ErrKeyNotFound возвращается, если ключ не найден в кэше.
)

// Node представляет элемент в кэше.
type Node struct {
	key       string      // Ключ элемента.
	value     interface{} // Значение элемента.
	expiresAt time.Time   // Время истечения срока действия элемента.
	prev      *Node       // Указатель на предыдущий элемент.
	next      *Node       // Указатель на следующий элемент.
}

// Cache представляет кэш с вытеснением по принципу LRU.
type Cache struct {
	cap        int              // Максимальная емкость кэша.
	bucket     map[string]*Node // Хранилище для элементов кэша.
	head, tail *Node            // Начало и конец двусвязного списка.
	mu         sync.RWMutex     // Мьютекс для обеспечения потокобезопасности.
	ttl        time.Duration    // Время жизни элемента по умолчанию.
}

func (c *Cache) remove(node *Node) {
	prev, next := node.prev, node.next
	prev.next, next.prev = next, prev
}

func (c *Cache) insert(node *Node) {
	prev, next := c.tail.prev, c.tail
	prev.next, next.prev = node, node
	node.prev, node.next = prev, next
}

// NewLRUCache создает новый кэш LRU с заданной емкостью и временем жизни по умолчанию.
func NewLRUCache(capacity int, ttl time.Duration) *Cache {
	head, tail := new(Node), new(Node)
	head.next, tail.prev = tail, head

	return &Cache{
		cap:    capacity,
		bucket: make(map[string]*Node),
		head:   head,
		tail:   tail,
		ttl:    ttl,
	}
}

// Put добавляет элемент в кэш. Если ключ уже существует, элемент и TTL обновляется.
// Если емкость превышена, самый старый элемент удаляется.
func (c *Cache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.ttl
	}
	expiresAt := time.Now().Add(ttl)

	if node, exists := c.bucket[key]; exists {
		c.remove(node)
	}

	c.bucket[key] = &Node{key, value, expiresAt, nil, nil}
	c.insert(c.bucket[key])

	if len(c.bucket) > c.cap {
		lru := c.head.next
		c.evictElement(lru)
	}

	return nil
}

// Get возвращает значение и время истечения для указанного ключа.
// Если ключ отсутствует или истек, возвращается ошибка.
func (c *Cache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.bucket[key]
	if !exists {
		return nil, time.Time{}, ErrKeyNotFound
	}

	if node.IsExpired() {
		c.evictElement(node)
		return nil, time.Time{}, ErrKeyNotFound
	}

	c.remove(node)
	c.insert(node)
	return node.value, node.expiresAt, nil
}

// GetAll возвращает все ключи и значения, которые еще не истекли.
// Если кэш пуст, возвращается ошибка.
func (c *Cache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node := c.head.next

	for node != c.tail {
		if !node.IsExpired() {
			keys = append(keys, node.key)
			values = append(values, node.value)
		} else {
			c.evictElement(node)
		}
		node = node.next
	}

	if len(c.bucket) == 0 {
		return nil, nil, ErrCacheIsEmpty
	}

	return keys, values, nil
}

// Evict удаляет указанный ключ из кэша и возвращает его значение.
// Если ключ отсутствует или истек, возвращается ошибка ErrKeyNotFound.
func (c *Cache) Evict(ctx context.Context, key string) (value interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.bucket[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if node.IsExpired() {
		c.evictElement(node)
		return nil, ErrKeyNotFound
	}

	c.evictElement(node)
	return node.value, nil
}

// EvictAll удаляет все элементы из кэша.
func (c *Cache) EvictAll(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bucket = make(map[string]*Node)
	c.head.next, c.tail.prev = c.tail, c.head

	return nil
}

// IsExpired проверяет, истек ли срок действия элемента.
func (n *Node) IsExpired() bool { return time.Now().After(n.expiresAt) }

// evictElement удаляет элемент из кэша.
func (c *Cache) evictElement(node *Node) {
	c.remove(node)
	delete(c.bucket, node.key)
}
