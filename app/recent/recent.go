package recent

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

// Cache is a collection of gym.Utilization values. It remembers the
// newest values and forgets the ones older older than a retention span
// before the newest one.
//
// The cache ignores values added already added to the cache.
//
// Note: The cache doesn't care about when the values are added just
// about their timestamps.
type Cache struct {
	retention time.Duration
	mux       sync.Mutex
	data      []*gym.Utilization
}

// NewCache returns a new cache with the give retention period.
func NewCache(retention time.Duration) (*Cache, error) {
	if retention == 0 {
		return nil, fmt.Errorf("retention must be >0, was %v", retention)
	}

	return &Cache{retention: retention}, nil
}

// Add adds values to the cache, ignoring the ones already in it.  After
// adding these values, all values in the cache older than its newest
// value minus the retention period will be forgotten.
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
	r.unique()
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

// Unique removes duplicated values from the cache. It assumes the data
// is sorted chronologically and the mutex is locked.
func (r *Cache) unique() {
	if len(r.data) < 2 {
		return
	}

	unique := make([]*gym.Utilization, 0, len(r.data))
	unique = append(unique, r.data[0])

	for _, d := range r.data[1:] {
		if d.Equal(unique[len(unique)-1]) {
			continue
		}

		unique = append(unique, d)
	}

	r.data = unique
}

// Get returns the most recent utilization values or an empty slice if
// the cache is still empty.
func (r *Cache) Get() []*gym.Utilization {
	r.mux.Lock()
	defer r.mux.Unlock()

	result := make([]*gym.Utilization, len(r.data))
	copy(result, r.data)

	return result
}
