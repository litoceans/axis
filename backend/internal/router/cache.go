package router

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Key       string
	Response  []byte
	CostUSD   float64
	CreatedAt time.Time
	ExpiresAt time.Time
}

// LRUCache is a thread-safe LRU cache
type LRUCache struct {
	mu        sync.RWMutex
	maxSize   int
	entries   map[string]*list.Element
	lru       *list.List
	ttl       time.Duration
}

// listElement wraps a cache entry for the LRU list
type cacheListEntry struct {
	key   string
	entry *CacheEntry
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	if maxSize <= 0 {
		maxSize = 100000
	}
	if ttl == 0 {
		ttl = 1 * time.Hour
	}

	return &LRUCache{
		maxSize: maxSize,
		entries: make(map[string]*list.Element),
		lru:     list.New(),
		ttl:     ttl,
	}
}

// Get retrieves a cached entry
func (c *LRUCache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	entry := elem.Value.(*cacheListEntry).entry

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		c.lru.Remove(elem)
		delete(c.entries, key)
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)

	return entry, true
}

// Set stores a response in cache
func (c *LRUCache) Set(key string, response []byte, costUSD float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key exists
	if elem, exists := c.entries[key]; exists {
		entry := elem.Value.(*cacheListEntry).entry
		entry.Response = response
		entry.CostUSD = costUSD
		entry.CreatedAt = time.Now()
		entry.ExpiresAt = time.Now().Add(c.ttl)
		c.lru.MoveToFront(elem)
		return
	}

	// Evict oldest if at capacity
	for c.lru.Len() >= c.maxSize {
		oldest := c.lru.Back()
		if oldest != nil {
			entry := oldest.Value.(*cacheListEntry)
			delete(c.entries, entry.key)
			c.lru.Remove(oldest)
		}
	}

	// Add new entry
	now := time.Now()
	entry := &CacheEntry{
		Key:       key,
		Response:  response,
		CostUSD:   costUSD,
		CreatedAt: now,
		ExpiresAt: now.Add(c.ttl),
	}

	c.entries[key] = c.lru.PushFront(&cacheListEntry{key: key, entry: entry})
}

// Delete removes an entry from cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.entries[key]; exists {
		c.lru.Remove(elem)
		delete(c.entries, key)
	}
}

// Clear removes all entries
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lru.Init()
}

// Size returns current cache size
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() (size, maxSize int, ttl time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len(), c.maxSize, c.ttl
}

// Cache provides response caching functionality
type Cache struct {
	lru            *LRUCache
	enabled        bool
}

// NewCache creates a new cache instance
func NewCache(enabled bool, maxEntries int, ttl time.Duration) *Cache {
	return &Cache{
		lru:     NewLRUCache(maxEntries, ttl),
		enabled: enabled,
	}
}

// GenerateKey generates a cache key from request data
func (c *Cache) GenerateKey(orgID, model string, messages []map[string]string) string {
	// Normalize messages for consistent hashing
	normalized, _ := json.Marshal(messages)

	// Create hash
	hash := sha256.Sum256(append([]byte(orgID+model), normalized...))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached response
func (c *Cache) Get(ctx context.Context, key string) ([]byte, float64, bool) {
	if !c.enabled {
		return nil, 0, false
	}

	entry, found := c.lru.Get(key)
	if !found {
		return nil, 0, false
	}

	return entry.Response, entry.CostUSD, true
}

// Set stores a response in cache
func (c *Cache) Set(ctx context.Context, key string, response []byte, costUSD float64) {
	if !c.enabled {
		return
	}

	c.lru.Set(key, response, costUSD)
}

// Invalidate removes a specific key from cache
func (c *Cache) Invalidate(ctx context.Context, key string) {
	c.lru.Delete(key)
}

// Clear removes all cached entries
func (c *Cache) Clear(ctx context.Context) {
	c.lru.Clear()
}

// Stats returns cache statistics
func (c *Cache) Stats() (size, maxSize int, ttl time.Duration, enabled bool) {
	size, maxSize, ttl = c.lru.Stats()
	return size, maxSize, ttl, c.enabled
}

// IsEnabled returns whether caching is enabled
func (c *Cache) IsEnabled() bool {
	return c.enabled
}
