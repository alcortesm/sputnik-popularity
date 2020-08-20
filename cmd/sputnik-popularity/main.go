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

	capacity := 2

	popularity, err := web.NewPopularity(capacity)
	if err != nil {
		logger.Fatalf("creating popularity: %v", err)
	}

	initialPairs := []pair.Pair{
		{Timestamp: time.Time{}.Add(time.Second), Value: 1.0},
		{Timestamp: time.Time{}.Add(2 * time.Second), Value: 2.0},
		{Timestamp: time.Time{}.Add(3 * time.Second), Value: 3.0},
	}

	popularity.Add(initialPairs...)

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

	log.Fatal(s.ListenAndServe())
}
