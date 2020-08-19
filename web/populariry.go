package web

import (
	"net/http"

	"github.com/alcortesm/sputnik-popularity/pair"
)

type Popularity struct {
	cache Cache
	page  []byte
}

// Cache is a collection of pairs.
type Cache interface {
	// Add adds some pairs to the collection.
	Add(...pair.Pair)
	// Get returns the pairs in chronological order.
	Get() []pair.Pair
}

func NewPopularity(c Cache) *Popularity {
	return &Popularity{
		cache: c,
		page: []byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Hello world!</h1>
	<p>No data yet.</p>

</body>
</html>`),
	}
}

func (p Popularity) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/html")
		w.Write(p.page)
	})
}

func (p Popularity) Add(pairs ...pair.Pair) {
	p.cache.Add(pairs...)
	p.page = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Hello world!</h1>
	<p>Foo.</p>

</body>
</html>`)
}
