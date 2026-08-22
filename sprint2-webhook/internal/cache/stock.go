// Package cache provides a thread-safe in-memory stock cache backed by
// sync.RWMutex. In production this would be replaced by Redis or a database.
package cache

import (
	"sync"
	"time"
)

// StockItem represents a single SKU's stock state.
type StockItem struct {
	Item      string    `json:"item"`
	Size      string    `json:"size"`
	SKU       string    `json:"sku,omitempty"`
	Qty       int       `json:"qty"`
	InStock   bool      `json:"inStock"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StockCache is a thread-safe in-memory cache keyed by "<item>|<size>".
type StockCache struct {
	mu    sync.RWMutex
	items map[string]StockItem
}

// New returns a StockCache seeded with Northstar's baseline inventory.
func New() *StockCache {
	c := &StockCache{items: make(map[string]StockItem)}
	baseline := []StockItem{
		{Item: "Trail Runner Jacket", Size: "S", Qty: 14, InStock: true},
		{Item: "Trail Runner Jacket", Size: "M", Qty: 6, InStock: true},
		{Item: "Trail Runner Jacket", Size: "L", Qty: 0, InStock: false},
		{Item: "Running Shoes", Size: "9", Qty: 22, InStock: true},
		{Item: "Running Shoes", Size: "10", Qty: 0, InStock: false},
		{Item: "Running Shoes", Size: "11", Qty: 9, InStock: true},
		{Item: "Ergonomic Office Chair", Size: "Standard", Qty: 3, InStock: true},
		{Item: "Yoga Mat Pro", Size: "Standard", Qty: 40, InStock: true},
	}
	for _, item := range baseline {
		c.Set(item)
	}
	return c
}

func cacheKey(item, size string) string { return item + "|" + size }

// Set writes or updates a stock item, stamping UpdatedAt to now.
func (c *StockCache) Set(item StockItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item.UpdatedAt = time.Now()
	c.items[cacheKey(item.Item, item.Size)] = item
}

// Get retrieves a stock item by item name and size.
func (c *StockCache) Get(itemName, size string) (StockItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[cacheKey(itemName, size)]
	return v, ok
}

// Search returns items whose Item or Size field contains query (case-insensitive).
func (c *StockCache) Search(query string) []StockItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var results []StockItem
	for _, item := range c.items {
		if containsCI(item.Item, query) || containsCI(item.Size, query) {
			results = append(results, item)
		}
	}
	return results
}

// All returns a snapshot of every cached item.
func (c *StockCache) All() []StockItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]StockItem, 0, len(c.items))
	for _, item := range c.items {
		result = append(result, item)
	}
	return result
}

func containsCI(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	sl, subl := toLower(s), toLower(substr)
	for i := 0; i <= len(sl)-len(subl); i++ {
		if sl[i:i+len(subl)] == subl {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
