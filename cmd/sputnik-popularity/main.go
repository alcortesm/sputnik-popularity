package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/alcortesm/sputnik-popularity/config"
	"github.com/alcortesm/sputnik-popularity/influx"
	"github.com/alcortesm/sputnik-popularity/pair"
	"github.com/alcortesm/sputnik-popularity/web"
)

func main() {
	logger := log.New(os.Stdout, "",
		log.Ldate|log.Ltime|log.LUTC)

	var config config.Config
	envPrefix := "SPUTNIK_POPULARITY"
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		logger.Fatalf("processign environment variables: %v", err)
	}

	if err := testInflux(config.InfluxDB, logger); err != nil {
		logger.Fatalf("testing influxdb: %v", err)
	}

	/*
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

		logger.Fatal(s.ListenAndServe())
	*/
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

func testInflux(config influx.Config, logger *log.Logger) error {
	const (
		measurement = "capacity_utilization"
		field       = "percent"
	)

	store, cancel := influx.NewStore(
		config,
		measurement,
		field,
	)
	defer cancel()

	p := pair.Pair{
		Value:     42.0,
		Timestamp: time.Now(),
	}

	if err := store.Add(context.Background(), p); err != nil {
		return fmt.Errorf("writing to influx: %v", err)
	}

	month := 30 * 24 * time.Hour
	got, err := store.Get(context.Background(), month)
	if err != nil {
		return fmt.Errorf("reading from influx: %v", err)
	}

	fmt.Println(got)

	return nil
}
