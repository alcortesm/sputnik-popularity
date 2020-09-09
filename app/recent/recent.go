package recent

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

// Cache is a collection of gym.Utilization values sorted
// chronologically. You can add values in any order and they will be
// sorted internally.
//
// Within the cache values should have unique timestamps; if you try
// to add values with repeated timestamps, the old values will be
// overwritten with the new ones.
//
// The cache will forget values older than the newest value minus a
// certain retention period, so its contents are always "recent" with
// respect of the newest value.
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

// Add adds values to the cache, overwriting previous values with the
// same timestamps as the ones being added. After adding these values,
// all values in the cache older than its newest value minus the
// retention period will be forgotten.
func (r *Cache) Add(data ...*gym.Utilization) {
	if len(data) == 0 {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.data = append(r.data, data...)

	// stability here and the unique method below helps to overwrite
	// elements with repeated timestamps with the values that has been
	// added later.
	sort.SliceStable(r.data, func(i, j int) bool {
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

// Unique removes duplicated values from the cache, keeping the ones
// that show up later in the data. It assumes the data is sorted
// chronologically and the mutex is locked.
func (r *Cache) unique() {
	if len(r.data) < 2 {
		return
	}

	unique := make([]*gym.Utilization, 0, len(r.data))
	unique = append(unique, r.data[0])

	for _, d := range r.data[1:] {
		l := len(unique) - 1 // index of the last unique element
		if d.Timestamp.Equal(unique[l].Timestamp) {
			// overwrite values with the same timestamp as the last one
			unique[l] = d
		} else {
			// add values with different timestamps than the last one
			unique = append(unique, d)
		}
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
