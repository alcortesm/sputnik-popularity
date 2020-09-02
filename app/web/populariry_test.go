package web_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/pair"
	"github.com/alcortesm/sputnik-popularity/app/web"
)

func TestPopularity(t *testing.T) {
	t.Parallel()

	subtests := map[string]func(t *testing.T){
		"shows special message if empty":    pEmpty,
		"shows added pairs":                 pShowsAddedPairs,
		"forgets oldes pairs":               pForgetsOldestPairs,
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

// Tests that after adding some pairs, they appear in the HTML.
func pShowsAddedPairs(t *testing.T) {
	pairs := []pair.Pair{
		{Timestamp: time.Time{}.Add(time.Second), Value: 1.0},
		{Timestamp: time.Time{}.Add(2 * time.Second), Value: 2.0},
	}

	cap := len(pairs)

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(pairs...)

	html := string(popularity.HTML())

	for _, p := range pairs {
		needle := p.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}
}

// Tests that after adding more pairs than the capacity, the oldest ones
// are forgotten.
func pForgetsOldestPairs(t *testing.T) {
	p1 := pair.Pair{Timestamp: time.Time{}.Add(time.Second), Value: 1.0}
	p2 := pair.Pair{Timestamp: p1.Timestamp.Add(time.Second), Value: 2.0}
	p3 := pair.Pair{Timestamp: p2.Timestamp.Add(time.Second), Value: 3.0}

	cap := 2 // p1 will be forgotten

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(p1, p2, p3)

	html := string(popularity.HTML())

	// check that p2 and p3 are part of the HTML
	for _, p := range []pair.Pair{p2, p3} {
		needle := p.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}

	// check that p1 is NOT in the HTML
	needle := p1.String()
	if strings.Contains(html, needle) {
		t.Errorf("found %q in HTML:\n%s", needle, html)
	}
}

// Tests that after adding more pairs than the capacity, the oldest ones
// are forgotten even if they are not added in chronological order.
func pOrderInsensitive(t *testing.T) {
	p1 := pair.Pair{Timestamp: time.Time{}.Add(time.Second), Value: 1.0}
	p2 := pair.Pair{Timestamp: p1.Timestamp.Add(time.Second), Value: 2.0}
	p3 := pair.Pair{Timestamp: p2.Timestamp.Add(time.Second), Value: 3.0}

	cap := 2 // p1 will be forgotten

	popularity, err := web.NewPopularity(cap)
	if err != nil {
		t.Fatal(err)
	}

	popularity.Add(p2, p3, p1) // out of order

	html := string(popularity.HTML())

	// check that p2 and p3 are part of the HTML
	for _, p := range []pair.Pair{p2, p3} {
		needle := p.String()
		if !strings.Contains(html, needle) {
			t.Errorf("cannot find %q in HTML:\n%s", needle, html)
		}
	}

	// check that p1 is NOT in the HTML
	needle := p1.String()
	if strings.Contains(html, needle) {
		t.Errorf("found %q in HTML:\n%s", needle, html)
	}
}
