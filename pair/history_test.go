package pair_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/alcortesm/sputnik-popularity/pair"
	"github.com/google/go-cmp/cmp"
)

func TestHistoryError(t *testing.T) {
	t.Parallel()

	caps := []int{0, -1, -10}

	for _, cap := range caps {
		cap := cap
		name := fmt.Sprintf("%d", cap)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h, err := pair.NewHistory(cap)
			if err == nil {
				t.Errorf("unexpected success, got %#v", h)
			}
		})
	}
}

func TestHistoryOK(t *testing.T) {
	t.Parallel()

	p1 := pair.Pair{
		Timestamp: time.Time{}.Add(time.Second),
		Value:     1,
	}

	p2 := pair.Pair{
		Timestamp: p1.Timestamp.Add(time.Second),
		Value:     2,
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
			want:  []pair.Pair{p2, p1},
		},
		{
			name:  "add 2 Pairs in non-chronological order",
			cap:   10,
			pairs: []pair.Pair{p2, p1},
			want:  []pair.Pair{p2, p1},
		},
		{
			name:  "add chronologically, cap reached",
			cap:   1,
			pairs: []pair.Pair{p1, p2},
			want:  []pair.Pair{p2},
		},
		{
			name:  "add non-chronologically, cap reached",
			cap:   1,
			pairs: []pair.Pair{p2, p1},
			want:  []pair.Pair{p2},
		},
	}

	for _, test := range subtests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			h, err := pair.NewHistory(test.cap)
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
