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

	cache, err := pair.NewHistory(10)
	if err != nil {
		logger.Fatalf("creating popularity historian: %v", err)
	}

	popularity := web.NewPopularity(cache)

	http.Handle("/popularity.html", web.Decorate(
		popularity.Handler(),
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
