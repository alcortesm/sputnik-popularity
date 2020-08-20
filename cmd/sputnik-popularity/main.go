package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alcortesm/sputnik-popularity/pair"
	"github.com/alcortesm/sputnik-popularity/web"
)

func main() {
	logger := log.New(os.Stdout, "",
		log.Ldate|log.Ltime|log.LUTC)

	capacity := 5

	popularity, err := web.NewPopularity(capacity)
	if err != nil {
		logger.Fatalf("creating popularity: %v", err)
	}

	http.Handle("/popularity.html", web.Decorate(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-type", "text/html")
				w.Write(popularity.HTML())
			},
		),
		web.WithLogs(logger),
	))

	http.Handle("/style.css", web.Decorate(
		web.StyleHandler(),
		web.WithLogs(logger),
	))

	http.Handle("/", web.Decorate(
		http.NotFoundHandler(),
		web.WithLogs(logger),
	))

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go addDemoValues(popularity, logger)

	log.Fatal(s.ListenAndServe())
}

func addDemoValues(p *web.Popularity, logger *log.Logger) {
	var value float64 = 0.0

	for {
		time.Sleep(time.Second)
		value++
		pair := pair.Pair{
			Timestamp: time.Now(),
			Value:     value,
		}
		if err := p.Add(pair); err != nil {
			logger.Fatalf("adding pair %v: %v", pair, err)
		}
	}
}
