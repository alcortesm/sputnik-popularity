package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
	Recent Getter
}

// Getter knows how to get gym utilization data.
type Getter interface {
	Get(context.Context) ([]*gym.Utilization, error)
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

		dataRaw, err := w.Recent.Get(r.Context())
		if err != nil {
			msg := fmt.Sprintf("getting recent data: %v", err)
			http.Error(rw, msg, http.StatusInternalServerError)
		}

		dataJSON, err := dataToJSON(dataRaw)
		if err != nil {
			msg := fmt.Sprintf("marshaling data to JSON: %v", err)
			http.Error(rw, msg, http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(rw, dataJSON); err != nil {
			msg := fmt.Sprintf("executing template: %v", err)
			http.Error(rw, msg, http.StatusInternalServerError)
			return
		}
	})
}

func dataToJSON(data []*gym.Utilization) (template.HTML, error) {
	type pairInt struct {
		Timestamp time.Time `json:"t"`
		Value     uint64    `json:"y"`
	}

	type pairFloat struct {
		Timestamp time.Time `json:"t"`
		Value     float64   `json:"y"`
	}

	payload := struct {
		People   []pairInt
		Capacity []pairInt
		Percent  []pairFloat
	}{
		People:   make([]pairInt, len(data)),
		Capacity: make([]pairInt, len(data)),
		Percent:  make([]pairFloat, len(data)),
	}

	for i, d := range data {
		payload.People[i] = pairInt{
			Timestamp: d.Timestamp,
			Value:     d.People,
		}

		payload.Capacity[i] = pairInt{
			Timestamp: d.Timestamp,
			Value:     d.Capacity,
		}

		p, ok := d.Percent()
		if !ok {
			p = 0.0
		}

		payload.Percent[i] = pairFloat{
			Timestamp: d.Timestamp,
			Value:     p,
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("generating JSON from data: %v", err)
	}

	return template.HTML(b), nil
}
