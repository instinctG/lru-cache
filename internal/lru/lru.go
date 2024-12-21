package lru

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Node struct {
	key       string
	value     interface{}
	expiresAt time.Time
	prev      *Node
	next      *Node
}

type LRUCache struct {
	cap        int
	bucket     map[string]*Node
	head, tail *Node
	mu         *sync.RWMutex
	ttl        time.Duration
}

func (c *LRUCache) remove(node *Node) {
	prev, next := node.prev, node.next
	prev.next, next.prev = next, prev
}

func (c *LRUCache) insert(node *Node) {
	prev, next := c.tail.prev, c.tail
	prev.next, next.prev = node, node
	node.prev, node.next = prev, next
}

func NewLRUCache(capacity int, defaultTTL time.Duration) *LRUCache {
	head, tail := new(Node), new(Node)
	head.next, tail.prev = tail, head

	return &LRUCache{
		cap:    capacity,
		bucket: make(map[string]*Node),
		head:   head,
		tail:   tail,
	}
}

func (c *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
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

func (c *LRUCache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
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

func (c *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
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

	return keys, values, nil
}

func (c *LRUCache) Evict(ctx context.Context, key string) (value interface{}, err error) {
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

func (c *LRUCache) EvictAll(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bucket = make(map[string]*Node)
	c.head.next, c.tail.prev = c.tail, c.head

	return nil
}

func (n *Node) IsExpired() bool { return time.Now().After(n.expiresAt) }

func (c *LRUCache) evictElement(node *Node) {
	c.remove(node)
	delete(c.bucket, node.key)
}
