package web

import (
	"fmt"
	"sort"

	"github.com/alcortesm/sputnik-popularity/pkg/pair"
)

// Cache is a collection of Pairs with a fixed capacity.
type Cache struct {
	cap  int
	data []pair.Pair
}

// Returns a new Cache with the given capacity.
func NewCache(cap int) (*Cache, error) {
	if cap < 1 {
		return nil, fmt.Errorf("invalid capacity %d", cap)
	}

	return &Cache{
		cap:  cap,
		data: []pair.Pair{},
	}, nil
}

// Add adds some pairs to the cache. If the number of pairs in the
// cache exceeds its capacity, the oldest surplus pairs will be
// deleted.
func (c *Cache) Add(pairs ...pair.Pair) {
	c.data = append(c.data, pairs...)
	sort.Slice(c.data, func(i, j int) bool {
		return c.data[i].Timestamp.Before(c.data[j].Timestamp)
	})

	if l := len(c.data); l > c.cap {
		c.data = c.data[l-c.cap : l]
	}
}

// Get returns all the pairs in the cache, in reverse chronological
// order.
func (c *Cache) Get() []pair.Pair {
	return c.data
}
