package web_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/alcortesm/sputnik-popularity/pkg/pair"
	"github.com/alcortesm/sputnik-popularity/pkg/web"
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

	p1 := pair.Pair{
		Timestamp: time.Time{}.Add(time.Second),
		Value:     1,
	}

	p2 := pair.Pair{
		Timestamp: p1.Timestamp.Add(time.Second),
		Value:     2,
	}

	p3 := pair.Pair{
		Timestamp: p2.Timestamp.Add(time.Second),
		Value:     3,
	}

	subtests := []struct {
		name  string
		cap   int
		pairs []pair.Pair
		want  []pair.Pair
	}{
		{
			name:  "empty",
			cap:   10,
			pairs: []pair.Pair{},
			want:  []pair.Pair{},
		},
		{
			name:  "add 1 Pair",
			cap:   10,
			pairs: []pair.Pair{p1},
			want:  []pair.Pair{p1},
		},
		{
			name:  "add 2 Pairs in chronological order",
			cap:   10,
			pairs: []pair.Pair{p1, p2},
			want:  []pair.Pair{p1, p2},
		},
		{
			name:  "add 2 Pairs in non-chronological order",
			cap:   10,
			pairs: []pair.Pair{p2, p1},
			want:  []pair.Pair{p1, p2},
		},
		{
			name:  "add chronologically, cap reached",
			cap:   2,
			pairs: []pair.Pair{p1, p2, p3},
			want:  []pair.Pair{p2, p3},
		},
		{
			name:  "add non-chronologically, cap reached",
			cap:   2,
			pairs: []pair.Pair{p2, p1, p3},
			want:  []pair.Pair{p2, p3},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			h, err := web.NewCache(test.cap)
			if err != nil {
				t.Fatal(err)
			}

			for _, p := range test.pairs {
				h.Add(p)
			}

			got := h.Get()

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}
}
