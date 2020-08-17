package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alcortesm/sputnik-popularity/web"
)

func main() {
	logger := log.New(os.Stdout, "",
		log.Ldate|log.Ltime|log.LUTC)

	http.Handle("/popularity.html", web.Decorate(
		web.Popularity(),
		web.WithLogs(logger),
	))

	http.Handle("/style.css", web.Decorate(
		web.Style(),
		web.WithLogs(logger),
	))

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(s.ListenAndServe())
}
