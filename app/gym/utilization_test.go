package gym_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

func TestUtilization_Percent(t *testing.T) {
	t.Parallel()

	subtests := []struct {
		people   uint64
		capacity uint64
		percent  float64
		ok       bool
	}{
		{people: 0, capacity: 0, percent: 0.0, ok: false},
		{people: 42, capacity: 0, percent: 0.0, ok: false},
		{people: 1, capacity: 100, percent: 1.0, ok: true},
		{people: 42, capacity: 100, percent: 42.0, ok: true},
		{people: 100, capacity: 100, percent: 100.0, ok: true},
		{people: 150, capacity: 100, percent: 150.0, ok: true},
	}

	for _, test := range subtests {
		test := test
		name := fmt.Sprintf("%d %d", test.people, test.capacity)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u := &gym.Utilization{
				People:   test.people,
				Capacity: test.capacity,
			}

			got, ok := u.Percent()

			if !equalFloats(got, test.percent) {
				t.Errorf("wrong percent: want %f, got %f\n",
					test.percent, got)
			}

			if ok != test.ok {
				t.Errorf("wrong ok: want %t, got %t\n", test.ok, ok)
			}
		})
	}
}

func equalFloats(a, b float64) bool {
	const tolerance = 1e-2
	return math.Abs(a-b) < tolerance
}

func TestUtilization_Equal(t *testing.T) {
	t.Parallel()

	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	t1 := time.Time{}.UTC().Add(1 * time.Second)
	t1NY := t1.In(newYork)
	t2 := time.Time{}.Add(2 * time.Second)

	subtests := []struct {
		name string
		a, b *gym.Utilization
		want bool
	}{
		{
			name: "same data, empty",
			a:    &gym.Utilization{},
			b:    &gym.Utilization{},
			want: true,
		},
		{
			name: "same data",
			a:    &gym.Utilization{Timestamp: t1, People: 1, Capacity: 2},
			b:    &gym.Utilization{Timestamp: t1, People: 1, Capacity: 2},
			want: true,
		},
		{
			name: "same data but different timezones",
			a:    &gym.Utilization{Timestamp: t1, People: 1, Capacity: 2},
			b:    &gym.Utilization{Timestamp: t1NY, People: 1, Capacity: 2},
			want: true,
		},
		{
			name: "different times",
			a:    &gym.Utilization{Timestamp: t1},
			b:    &gym.Utilization{Timestamp: t2},
			want: false,
		},
		{
			name: "different people",
			a:    &gym.Utilization{People: 1},
			b:    &gym.Utilization{People: 2},
			want: false,
		},
		{
			name: "different capacity",
			a:    &gym.Utilization{Capacity: 1},
			b:    &gym.Utilization{Capacity: 2},
			want: false,
		},
	}

	for _, test := range subtests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.a.Equal(test.b)
			if got != test.want {
				t.Errorf("direct test, want %t, got %t", test.want, got)
			}

			got = test.b.Equal(test.a)
			if got != test.want {
				t.Errorf("reverse test, want %t, got %t", test.want, got)
			}
		})
	}
}
