package web_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/alcortesm/sputnik-popularity/app/gym"
	"github.com/alcortesm/sputnik-popularity/app/web"
)

func TestCacheError(t *testing.T) {
	t.Parallel()

	caps := []int{0, -1, -10}

	for _, cap := range caps {
		cap := cap
		name := fmt.Sprintf("%d", cap)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h, err := web.NewCache(cap)
			if err == nil {
				t.Errorf("unexpected success, got %#v", h)
			}
		})
	}
}

func TestCacheOK(t *testing.T) {
	t.Parallel()

	t1 := time.Time{}.Add(1 * time.Second)
	t2 := time.Time{}.Add(2 * time.Second)
	t3 := time.Time{}.Add(3 * time.Second)

	u1 := &gym.Utilization{
		Timestamp: t1,
		People:    1,
		Capacity:  42,
	}

	u2 := &gym.Utilization{
		Timestamp: t2,
		People:    2,
		Capacity:  42,
	}

	u3 := &gym.Utilization{
		Timestamp: t3,
		People:    3,
		Capacity:  42,
	}

	subtests := []struct {
		name  string
		cap   int
		toAdd []*gym.Utilization
		toGet []*gym.Utilization
	}{
		{
			name:  "empty",
			cap:   10,
			toAdd: []*gym.Utilization{},
			toGet: []*gym.Utilization{},
		},
		{
			name:  "add 1 utilization sample",
			cap:   10,
			toAdd: []*gym.Utilization{u1},
			toGet: []*gym.Utilization{u1},
		},
		{
			name:  "add 2 samples in chronological order",
			cap:   10,
			toAdd: []*gym.Utilization{u1, u2},
			toGet: []*gym.Utilization{u1, u2},
		},
		{
			name:  "add 2 samples in non-chronological order",
			cap:   10,
			toAdd: []*gym.Utilization{u2, u1},
			toGet: []*gym.Utilization{u1, u2},
		},
		{
			name:  "add chronologically, cap reached",
			cap:   2,
			toAdd: []*gym.Utilization{u1, u2, u3},
			toGet: []*gym.Utilization{u2, u3},
		},
		{
			name:  "add non-chronologically, cap reached",
			cap:   2,
			toAdd: []*gym.Utilization{u2, u1, u3},
			toGet: []*gym.Utilization{u2, u3},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cache, err := web.NewCache(test.cap)
			if err != nil {
				t.Fatal(err)
			}

			cache.Add(test.toAdd...)

			got := cache.Get()

			if diff := cmp.Diff(test.toGet, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}
}
