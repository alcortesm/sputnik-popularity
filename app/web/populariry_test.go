package web_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
	"github.com/alcortesm/sputnik-popularity/app/web"
)

func TestPopularity(t *testing.T) {
	t.Parallel()

	subtests := map[string]func(t *testing.T){
		"shows special message if empty":    pEmpty,
		"shows added utilization data":      pShowsAddedData,
		"forgets oldes utilization data":    pForgetsOldestData,
		"add order do not affect deletions": pOrderInsensitive,
	}

	for name, fn := range subtests {
		fn := fn
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fn(t)
		})
	}
}

// Tests that if the cahce is empty, certain message is returned.
func pEmpty(t *testing.T) {
	cap := 10 // any valid value will do

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	html := string(popularity.HTML())

	cause := "no data available"

	if !strings.Contains(html, cause) {
		t.Errorf("cannot find %q in HTML:\n%s", cause, html)
	}
}

// Tests that after adding some data, they appear in the HTML.
func pShowsAddedData(t *testing.T) {
	t1 := time.Time{}.Add(1 * time.Second)
	t2 := time.Time{}.Add(2 * time.Second)

	data := []*gym.Utilization{
		{Timestamp: t1, People: 1.0, Capacity: 42},
		{Timestamp: t2, People: 2.0, Capacity: 42},
	}

	cap := len(data)

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(data...)

	html := string(popularity.HTML())

	for _, d := range data {
		needle := d.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}
}

// Tests that after adding more utilization data than the capacity, the
// oldest ones are forgotten.
func pForgetsOldestData(t *testing.T) {
	t1 := time.Time{}.Add(1 * time.Second)
	t2 := time.Time{}.Add(2 * time.Second)
	t3 := time.Time{}.Add(3 * time.Second)

	u1 := &gym.Utilization{Timestamp: t1, People: 1, Capacity: 42}
	u2 := &gym.Utilization{Timestamp: t2, People: 2, Capacity: 42}
	u3 := &gym.Utilization{Timestamp: t3, People: 3, Capacity: 42}

	cap := 2 // p1 will be forgotten

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(u1, u2, u3)

	html := string(popularity.HTML())

	// check that p2 and p3 are part of the HTML
	for _, u := range []*gym.Utilization{u2, u3} {
		needle := u.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}

	// check that p1 is NOT in the HTML
	needle := u1.String()
	if strings.Contains(html, needle) {
		t.Errorf("found %q in HTML:\n%s", needle, html)
	}
}

// Tests that after adding more utilization data than the capacity, the
// oldest ones are forgotten even if they are not added in chronological
// order.
func pOrderInsensitive(t *testing.T) {
	t1 := time.Time{}.Add(1 * time.Second)
	t2 := time.Time{}.Add(2 * time.Second)
	t3 := time.Time{}.Add(3 * time.Second)

	u1 := &gym.Utilization{Timestamp: t1, People: 1, Capacity: 42}
	u2 := &gym.Utilization{Timestamp: t2, People: 2, Capacity: 42}
	u3 := &gym.Utilization{Timestamp: t3, People: 3, Capacity: 42}

	cap := 2 // p1 will be forgotten

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(u2, u3, u1) // out of order

	html := string(popularity.HTML())

	// check that p2 and p3 are part of the HTML
	for _, u := range []*gym.Utilization{u2, u3} {
		needle := u.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}

	// check that p1 is NOT in the HTML
	needle := u1.String()
	if strings.Contains(html, needle) {
		t.Errorf("found %q in HTML:\n%s", needle, html)
	}
}
