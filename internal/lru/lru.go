// Package lru реализует кэш с вытеснением по принципу LRU (Least Recently Used) и использованием TTL.
package lru

import (
	"context"
	"errors"
	"sync"
	"time"
)

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
	Cap        int              // Максимальная емкость кэша.
	Bucket     map[string]*Node // Хранилище для элементов кэша.
	Head, Tail *Node            // Начало и конец двусвязного списка.
	Mu         sync.RWMutex     // Мьютекс для обеспечения потокобезопасности.
	TTL        time.Duration    // Время жизни элемента по умолчанию.
}

func (c *Cache) remove(node *Node) {
	prev, next := node.prev, node.next
	prev.next, next.prev = next, prev
}

func (c *Cache) insert(node *Node) {
	prev, next := c.Tail.prev, c.Tail
	prev.next, next.prev = node, node
	node.prev, node.next = prev, next
}

// NewLRUCache создает новый кэш LRU с заданной емкостью и временем жизни по умолчанию.
func NewLRUCache(capacity int, ttl time.Duration) *Cache {
	head, tail := new(Node), new(Node)
	head.next, tail.prev = tail, head

	return &Cache{
		Cap:    capacity,
		Bucket: make(map[string]*Node),
		Head:   head,
		Tail:   tail,
		TTL:    ttl,
	}
}

// Put добавляет элемент в кэш. Если ключ уже существует, элемент и TTL обновляется.
// Если емкость превышена, самый старый элемент удаляется.
func (c *Cache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if ttl == 0 {
		ttl = c.TTL
	}
	expiresAt := time.Now().Add(ttl)

	if node, exists := c.Bucket[key]; exists {
		c.remove(node)
	}

	c.Bucket[key] = &Node{key, value, expiresAt, nil, nil}
	c.insert(c.Bucket[key])

	if len(c.Bucket) > c.Cap {
		lru := c.Head.next
		c.evictElement(lru)
	}

	return nil
}

// Get возвращает значение и время истечения для указанного ключа.
// Если ключ отсутствует или истек, возвращается ошибка.
func (c *Cache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	node, exists := c.Bucket[key]
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
	c.Mu.Lock()
	defer c.Mu.Unlock()

	node := c.Head.next

	for node != c.Tail {
		if !node.IsExpired() {
			keys = append(keys, node.key)
			values = append(values, node.value)
		} else {
			c.evictElement(node)
		}
		node = node.next
	}

	if len(c.Bucket) == 0 {
		return nil, nil, ErrCacheIsEmpty
	}

	return keys, values, nil
}

// Evict удаляет указанный ключ из кэша и возвращает его значение.
// Если ключ отсутствует или истек, возвращается ошибка ErrKeyNotFound.
func (c *Cache) Evict(ctx context.Context, key string) (value interface{}, err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	node, exists := c.Bucket[key]
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
	c.Mu.Lock()
	defer c.Mu.Unlock()

	c.Bucket = make(map[string]*Node)
	c.Head.next, c.Tail.prev = c.Tail, c.Head

	return nil
}

// IsExpired проверяет, истек ли срок действия элемента.
func (n *Node) IsExpired() bool { return time.Now().After(n.expiresAt) }

func (c *Cache) evictElement(node *Node) {
	c.remove(node)
	delete(c.Bucket, node.key)
}
