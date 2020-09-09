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
		"retention must be more than zero": invalidRetention,
		"can get from empty cache":         canGetFromEmpty,
		"remembers recent values":          remembersRecentValues,
		"forgets old values":               forgetsOldValues,
		"overwrites values":                overwritesValues,
		"all mixed together":               allMixed,
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

func remembersRecentValues(t *testing.T) {
	u1 := fixValue(t, 1)
	u2 := fixValue(t, 2)
	u3 := fixValue(t, 3)

	retention := 2 * time.Second // enough to remember all data points

	subtests := []struct {
		name    string
		batches [][]*gym.Utilization
		want    []*gym.Utilization
	}{
		{
			name:    "add nothing",
			batches: nil,
			want:    []*gym.Utilization{},
		}, {
			name:    "add empty batch",
			batches: [][]*gym.Utilization{{}},
			want:    []*gym.Utilization{},
		}, {
			name:    "add empty batches",
			batches: [][]*gym.Utilization{{}, {}, {}},
			want:    []*gym.Utilization{},
		}, {
			name:    "one batch with one value",
			batches: [][]*gym.Utilization{{u1}},
			want:    []*gym.Utilization{u1},
		}, {
			name:    "one batch with three values",
			batches: [][]*gym.Utilization{{u1, u2, u3}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "one batch with three values, unsorted",
			batches: [][]*gym.Utilization{{u3, u1, u2}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "three batches with one value each",
			batches: [][]*gym.Utilization{{u1}, {u2}, {u3}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "three batches with one value each, unsorted",
			batches: [][]*gym.Utilization{{u3}, {u1}, {u2}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "batches of different sizes",
			batches: [][]*gym.Utilization{{u1, u2}, {u3}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "batches of different sizes, unsorted",
			batches: [][]*gym.Utilization{{u3, u1}, {u2}},
			want:    []*gym.Utilization{u1, u2, u3},
		}, {
			name:    "repeated batches with one value",
			batches: [][]*gym.Utilization{{u1}, {u1}},
			want:    []*gym.Utilization{u1},
		}, {
			name:    "repeated batches with two values",
			batches: [][]*gym.Utilization{{u1, u2}, {u1, u2}},
			want:    []*gym.Utilization{u1, u2},
		}, {
			name:    "repeated batches with two values unsorted",
			batches: [][]*gym.Utilization{{u1, u2}, {u2, u1}},
			want:    []*gym.Utilization{u1, u2},
		}, {
			name:    "batches mixing unique and repeated values",
			batches: [][]*gym.Utilization{{u1, u3}, {u2, u1}, {u3, u2}},
			want:    []*gym.Utilization{u1, u2, u3},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cache, err := recent.NewCache(retention)
			if err != nil {
				t.Fatal(err)
			}

			for _, b := range test.batches {
				cache.Add(b...)
			}

			got := cache.Get()

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}
}

// fixValue returns a gym.Utilization data point
// with some fixed values useful in tests:
// - timestamp will be year 2020 plus n seconds
// - capacity 100 + n
// - people will be n.
func fixValue(t *testing.T, n int) *gym.Utilization {
	t.Helper()

	if n < 1 {
		t.Fatalf("fixDataPoint arg must be greater than 0, was %d", n)
	}

	year2020 := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	return &gym.Utilization{
		Timestamp: year2020.Add(time.Duration(n) * time.Second),
		Capacity:  uint64(100 + n),
		People:    uint64(n),
	}
}

func forgetsOldValues(t *testing.T) {
	u1 := fixValue(t, 1)
	u2 := fixValue(t, 2)
	u3 := fixValue(t, 3)
	u4 := fixValue(t, 4)
	u5 := fixValue(t, 5)

	// remember up to 3 one-second apart values
	retention := 2 * time.Second

	subtests := []struct {
		name    string
		batches [][]*gym.Utilization
		want    []*gym.Utilization
	}{
		{
			name:    "one batch",
			batches: [][]*gym.Utilization{{u1, u2, u3, u4, u5}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "two batches",
			batches: [][]*gym.Utilization{{u1}, {u2, u3, u4, u5}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "old values in second batch",
			batches: [][]*gym.Utilization{{u5}, {u1, u2}},
			want:    []*gym.Utilization{u5},
		}, {
			name:    "two batches, two forgets",
			batches: [][]*gym.Utilization{{u1, u4}, {u2, u3, u4, u5}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "two batches, interleaved",
			batches: [][]*gym.Utilization{{u1, u3}, {u2, u4, u5}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "two batches, interleaved, unsorted",
			batches: [][]*gym.Utilization{{u3, u1}, {u4, u5, u2}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "two batches, interleaved, unsorted, repeated",
			batches: [][]*gym.Utilization{{u3, u2, u1}, {u3, u4, u5, u2}},
			want:    []*gym.Utilization{u3, u4, u5},
		}, {
			name:    "a very new value forces to forget all old values",
			batches: [][]*gym.Utilization{{u1, u2}, {u5}},
			want:    []*gym.Utilization{u5},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cache, err := recent.NewCache(retention)
			if err != nil {
				t.Fatal(err)
			}

			for _, b := range test.batches {
				cache.Add(b...)
			}

			got := cache.Get()

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}
}

func overwritesValues(t *testing.T) {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	u1 := fixValue(t, 1)
	u2 := fixValue(t, 2)

	// a, b and c all have the same timestamp, but different values:
	// - b and c has different capacity (42) than a (101)
	// - c has different timezone (New York) than a and b (UTC)
	var u3a, u3b, u3c *gym.Utilization
	{
		u3a = fixValue(t, 3)

		u3b = fixValue(t, 3)
		u3b.Capacity = 42

		u3c = fixValue(t, 3)
		u3c.Capacity = 42
		u3c.Timestamp = u3a.Timestamp.In(newYork)
	}

	// remembers up to 2 one-second apart values
	retention := 1 * time.Second

	subtests := []struct {
		name    string
		batches [][]*gym.Utilization
		want    []*gym.Utilization
	}{
		{
			name:    "one batch, one value",
			batches: [][]*gym.Utilization{{u3a, u3b}},
			want:    []*gym.Utilization{u3b},
		}, {
			name:    "one batch, one value, different timezones",
			batches: [][]*gym.Utilization{{u3a, u3c}},
			want:    []*gym.Utilization{u3c},
		}, {
			name:    "same batch",
			batches: [][]*gym.Utilization{{u2, u3a, u3b}},
			want:    []*gym.Utilization{u2, u3b},
		}, {
			name:    "same batch, different timezone",
			batches: [][]*gym.Utilization{{u2, u3a, u3c}},
			want:    []*gym.Utilization{u2, u3c},
		}, {
			name:    "different batches",
			batches: [][]*gym.Utilization{{u2, u3a}, {u3b}},
			want:    []*gym.Utilization{u2, u3b},
		}, {
			name:    "different batches, different timezone",
			batches: [][]*gym.Utilization{{u2, u3a}, {u3c}},
			want:    []*gym.Utilization{u2, u3c},
		}, {
			name:    "same batch, forgets",
			batches: [][]*gym.Utilization{{u1, u2, u3a, u3b}},
			want:    []*gym.Utilization{u2, u3b},
		}, {
			name:    "same batch, different timezone, forgets",
			batches: [][]*gym.Utilization{{u1, u2, u3a, u3c}},
			want:    []*gym.Utilization{u2, u3c},
		}, {
			name:    "different batches, forgets",
			batches: [][]*gym.Utilization{{u1, u2, u3a}, {u3b}},
			want:    []*gym.Utilization{u2, u3b},
		}, {
			name:    "different batches, different timezone, forgets",
			batches: [][]*gym.Utilization{{u1, u2, u3a}, {u3c}},
			want:    []*gym.Utilization{u2, u3c},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cache, err := recent.NewCache(retention)
			if err != nil {
				t.Fatal(err)
			}

			for _, b := range test.batches {
				cache.Add(b...)
			}

			got := cache.Get()

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}
}

func allMixed(t *testing.T) {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	u1 := fixValue(t, 1)
	u2 := fixValue(t, 2)
	u3 := fixValue(t, 3)
	u4 := fixValue(t, 4)
	u4b := fixValue(t, 4)
	u4b.Capacity = 42
	u5 := fixValue(t, 5)
	u6 := fixValue(t, 6)
	u6b := fixValue(t, 6)
	u6b.Timestamp = u6b.Timestamp.In(newYork)
	u6b.Capacity = 42

	batches := [][]*gym.Utilization{
		{u1, u5, u1, u2},
		{u3, u5, u1, u4, u4, u2},
		{u4b, u1, u6, u3},
		{u6b},
		{u1, u1},
	}

	// remember up to 3 one-second values
	retention := 2 * time.Second

	want := []*gym.Utilization{u4b, u5, u6b}

	cache, err := recent.NewCache(retention)
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range batches {
		cache.Add(b...)
	}

	got := cache.Get()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}
