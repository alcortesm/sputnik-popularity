package recent_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/alcortesm/sputnik-popularity/app/gym"
	"github.com/alcortesm/sputnik-popularity/app/recent"
)

func TestCache(t *testing.T) {
	t.Parallel()

	subtests := map[string]func(t *testing.T){
		"retention must be more than zero":     invalidRetention,
		"can get from empty cache":             canGetFromEmpty,
		"keeps recent values in a batch":       keepsRecentInBatch,
		"adding unsorted values is ok":         canAddUnsorted,
		"adding zero elements is ok":           canAddZeroElements,
		"forgets old values in the same batch": forgetsOldInSameBatch,
		"forgets values in old batches":        forgetsOldInOtherBatches,
	}

	for name, fn := range subtests {
		fn := fn
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fn(t)
		})
	}
}

func invalidRetention(t *testing.T) {
	for _, retention := range []int{0, -1} {
		name := fmt.Sprintf("%d", retention)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := recent.NewCache(0)
			if err == nil {
				t.Fatal("unexpected success")
			}
		})
	}
}

func canGetFromEmpty(t *testing.T) {
	retention := time.Second // irrelevant

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	got := cache.Get()

	if len(got) != 0 {
		t.Errorf("want empty slice, got %#v", got)
	}
}

func keepsRecentInBatch(t *testing.T) {
	u1 := fixDataPoint(t, 1)
	u2 := fixDataPoint(t, 2)

	batch := []*gym.Utilization{u1, u2}
	retention := 1 * time.Second // the whole batch

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	cache.Add(batch...)

	got := cache.Get()
	want := batch

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}

// fixDataPoint returns a gym.Utilization data point
// with some fixed values useful in tests: timestamp will be n seconds,
// capacity 100 + n and people will be n.
func fixDataPoint(t *testing.T, n int) *gym.Utilization {
	t.Helper()

	if n < 1 {
		t.Fatalf("fixDataPoint arg must be greater than 0, was %d", n)
	}

	return &gym.Utilization{
		Timestamp: time.Time{}.Add(time.Duration(n) * time.Second),
		Capacity:  uint64(100 + n),
		People:    uint64(n),
	}
}

func canAddUnsorted(t *testing.T) {
	u1 := fixDataPoint(t, 1)
	u2 := fixDataPoint(t, 2)
	u3 := fixDataPoint(t, 3)

	batch := []*gym.Utilization{u2, u3, u1} // unsorted
	retention := 2 * time.Second            // the whole batch

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	cache.Add(batch...)

	got := cache.Get()
	want := []*gym.Utilization{u1, u2, u3} // sorted

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}

func canAddZeroElements(t *testing.T) {
	u1 := fixDataPoint(t, 1)
	u2 := fixDataPoint(t, 2)

	batch := []*gym.Utilization{u1, u2}
	retention := 1 * time.Second // the whole batch

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	cache.Add(batch...)
	cache.Add() // add zero elements

	got := cache.Get()
	want := batch

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}

func forgetsOldInSameBatch(t *testing.T) {
	// u1 and u2 will be forgotten due to an small retention period
	u1 := fixDataPoint(t, 1)
	u2 := fixDataPoint(t, 2)
	u3 := fixDataPoint(t, 3)
	u4 := fixDataPoint(t, 4)

	batch := []*gym.Utilization{u1, u2, u3, u4}
	retention := 1 * time.Second // only half of the batch

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	cache.Add(batch...)

	got := cache.Get()
	want := []*gym.Utilization{u3, u4}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}

func forgetsOldInOtherBatches(t *testing.T) {
	// u1 and u2 will be forgotten due to a small retention period
	u1 := fixDataPoint(t, 1)
	u2 := fixDataPoint(t, 2)
	u3 := fixDataPoint(t, 3)
	u4 := fixDataPoint(t, 4)
	u5 := fixDataPoint(t, 5)
	u6 := fixDataPoint(t, 6)

	oldBatch := []*gym.Utilization{u1, u2, u3, u4}
	newBatch := []*gym.Utilization{u5, u6}

	// enough to remember the whole first batch and then forget the
	// first two data points when the second batch arraives.
	retention := 3 * time.Second

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	cache.Add(oldBatch...)
	cache.Add(newBatch...)

	got := cache.Get()
	want := []*gym.Utilization{u3, u4, u5, u6}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}
