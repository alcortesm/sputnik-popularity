package web

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"

	"github.com/alcortesm/sputnik-popularity/app/pair"
)

var tmpl = template.Must(template.New("popularity table").
	Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Sputnik Popularity</h1>
	{{range .}}
	<p>{{.}}</p>
	{{end}}

</body>
</html>`))

const noData = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Hello World</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>

	<h1>Sputnik Popularity</h1>
	<p>There are no data available.</p>

</body>
</html>`

// Popularity keeps track of pairs and allow to generate an HTML
// representation of the newest ones. It has a miximum capcacity of
// pairs: when adding new pairs, the surplus oldest ones will be
// forgotten.
//
// The Add and HTML methods are thread-safe.
type Popularity struct {
	lock  *sync.Mutex
	cache *Cache
	page  []byte
}

// NewPopularity returns a new Popularity with the given capacity.
func NewPopularity(cap int) (*Popularity, error) {
	cache, err := NewCache(cap)
	if err != nil {
		return nil, err
	}

	p := &Popularity{
		lock:  &sync.Mutex{},
		cache: cache,
	}

	if err := p.createPage(); err != nil {
		return nil, fmt.Errorf("cannot create page: %v", err)
	}

	return p, nil
}

// HTML returns a web page with the newest pairs added.
func (p *Popularity) HTML() []byte {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.page
}

// Add adds the given pairs to the web page.
func (p *Popularity) Add(pairs ...pair.Pair) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.cache.Add(pairs...)

	if err := p.createPage(); err != nil {
		return fmt.Errorf("creating new page: %v", err)
	}

	return nil
}

func (p *Popularity) createPage() error {
	pairs := p.cache.Get()

	if len(pairs) == 0 {
		p.page = []byte(noData)
		return nil
	}

	b := &bytes.Buffer{}

	if err := tmpl.Execute(b, pairs); err != nil {
		return err
	}

	p.page = b.Bytes()

	return nil
}
