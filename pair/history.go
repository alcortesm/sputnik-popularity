package pair

import (
	"fmt"
	"sort"
)

// History is a collection of Pairs.
type History struct {
	cap  int
	data []Pair
}

// Returns a new History with the given capacity.
func NewHistory(cap int) (*History, error) {
	if cap < 1 {
		return nil, fmt.Errorf("invalid capacity %d", cap)
	}

	return &History{
		cap:  cap,
		data: []Pair{},
	}, nil
}

// Add adds some pairs pp to the history. If the number of pairs in the
// history exceeds its capacity, the oldest surplus pairs will be
// deleted.
func (h *History) Add(pairs ...Pair) {
	h.data = append(h.data, pairs...)
	sort.Slice(h.data, func(i, j int) bool {
		return h.data[i].Timestamp.After(h.data[j].Timestamp)
	})

	if len(h.data) > h.cap {
		h.data = h.data[:h.cap]
	}
}

// Get returns all the pairs in the history, in reverse chronological
// order.
func (h *History) Get() []Pair {
	return h.data
}
