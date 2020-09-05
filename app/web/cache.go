package web

import (
	"fmt"
	"sort"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

// Cache is a collection of gym.Utilization data with a fixed capacity.
type Cache struct {
	cap  int
	data []*gym.Utilization
}

// Returns a new Cache with the given capacity.
func NewCache(cap int) (*Cache, error) {
	if cap < 1 {
		return nil, fmt.Errorf("invalid capacity %d", cap)
	}

	return &Cache{
		cap:  cap,
		data: []*gym.Utilization{},
	}, nil
}

// Add adds some data to the cache. If the number of items in the
// cache exceeds its capacity, the oldest surplus items will be
// deleted.
func (c *Cache) Add(data ...*gym.Utilization) {
	c.data = append(c.data, data...)
	sort.Slice(c.data, func(i, j int) bool {
		return c.data[i].Timestamp.Before(c.data[j].Timestamp)
	})

	if l := len(c.data); l > c.cap {
		c.data = c.data[l-c.cap : l]
	}
}

// Get returns all the data in the cache, in reverse chronological
// order.
func (c *Cache) Get() []*gym.Utilization {
	return c.data
}
