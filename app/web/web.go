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
	template.New("chart").
		Funcs(template.FuncMap{
			"rfc3339": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
		}).
		Parse(chartTemplate))

type Web struct {
	Recent RecentGetter
}

// RecentGetter knows how to get the most recent gym utilization data.
//
// TODO: Get probably needs a context and an error.
type RecentGetter interface {
	Get() []*gym.Utilization
}

func (w Web) PopularityHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-type", "text/html")
		rw.Write([]byte(popularity))
	})
}

func (w Web) StyleHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/css")
		w.Write([]byte(css))
	})
}

func (w Web) ChartHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-type", "application/javascript")

		data := w.Recent.Get()

		if err := tmpl.Execute(rw, format(data)); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

type formatted struct {
	People template.HTML
}

func format(data []*gym.Utilization) formatted {
	var people strings.Builder

	people.WriteString("[")

	sep := ""
	for _, d := range data {
		fmt.Fprintf(&people, `%s{t: %q, y: %d}`,
			sep, d.Timestamp.Format(time.RFC3339), d.People)
		sep = ", "
	}

	people.WriteString("]")

	return formatted{
		People: template.HTML(people.String()),
	}
}
