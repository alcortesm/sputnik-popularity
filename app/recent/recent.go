package recent

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

// Cache is a collection of gym.Utilization values. It remembers the
// newest values and forgets values older than a certain retention span
// before the newest one. It doesn't care about when the values are
// added just about their timestamps.
type Cache struct {
	retention time.Duration
	mux       sync.Mutex
	data      []*gym.Utilization
}

func NewCache(retention time.Duration) (*Cache, error) {
	if retention == 0 {
		return nil, fmt.Errorf("retention must be >0, was %v", retention)
	}

	return &Cache{retention: retention}, nil
}

func (r *Cache) Add(data ...*gym.Utilization) {
	if len(data) == 0 {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.data = append(r.data, data...)

	sort.Slice(r.data, func(i, j int) bool {
		return r.data[i].Timestamp.Before(r.data[j].Timestamp)
	})

	r.trim()
}

// Trim removes elements from r.data with a timestamp older than the
// timestamp of the newest element minus retention.  It assumes the
// elements are sorted chronologically and the mutex is locked.
func (r *Cache) trim() {
	newest := r.data[len(r.data)-1]

	// elements before threshold will be forgotten
	threshold := newest.Timestamp.Add(-r.retention)

	// find the index of the first element to keep, this is, the older
	// element after the threshold
	first := 0
	for ; r.data[first].Timestamp.Before(threshold); first++ {
	}

	r.data = r.data[first:]
}

// Get returns the most recent utilization values or an empty string if
// the cache is still empty.
func (r *Cache) Get() []*gym.Utilization {
	r.mux.Lock()
	defer r.mux.Unlock()

	result := make([]*gym.Utilization, len(r.data))
	copy(result, r.data)

	return result
}
