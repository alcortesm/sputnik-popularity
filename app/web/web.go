package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

var tmpl = template.Must(
	template.New("popularity table").
		Funcs(template.FuncMap{
			"rfc3339": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
		}).
		Parse(popularityTemplate))

type Web struct {
	Recent RecentGetter
}

// RecentGetter knows how to get the most recent gym utilization data.
//
// TODO: Get probably needs a context and an error.
type RecentGetter interface {
	Get() []*gym.Utilization
}

func (w Web) Handler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-type", "text/html")

		data := w.Recent.Get()

		if len(data) == 0 {
			rw.Write([]byte(noData))
			return
		}

		if err := tmpl.Execute(rw, format(data)); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

type formatted struct {
	People template.JS
}

func format(data []*gym.Utilization) formatted {
	var people strings.Builder

	sep := ""
	for _, d := range data {
		fmt.Fprintf(&people, `%s{ t: %q, y: %d }`,
			sep, d.Timestamp.Format(time.RFC3339), d.People)
		sep = ", "
	}

	return formatted{
		People: template.JS(people.String()),
	}
}
