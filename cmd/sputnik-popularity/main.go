package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	influx "github.com/influxdata/influxdb-client-go/v2"
	influxlog "github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/kelseyhightower/envconfig"

	"github.com/alcortesm/sputnik-popularity/config"
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

func testInflux(cfg config.InfluxDB, logger *log.Logger) error {
	// disable influxdb client logs
	influxlog.Log = nil

	const (
		measurement = "capacity_utilization"
		field       = "percent"
	)

	if err := testWrite(cfg, measurement, field); err != nil {
		return fmt.Errorf("testing write: %v", err)
	}

	if err := testRead(cfg, measurement, field); err != nil {
		return fmt.Errorf("testing read: %v", err)
	}

	return nil
}

func testWrite(cfg config.InfluxDB, measurement, field string) error {
	client := influx.NewClient(cfg.URL, cfg.TokenWrite)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)

	p := influx.NewPoint(measurement,
		nil,
		map[string]interface{}{field: 45},
		time.Now(),
	)

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		return fmt.Errorf("writing point: %v", err)
	}

	return nil
}

func testRead(cfg config.InfluxDB, measurement, field string) error {
	client := influx.NewClient(cfg.URL, cfg.TokenRead)
	defer client.Close()

	queryAPI := client.QueryAPI(cfg.Org)

	const start = "-30d"

	query := fmt.Sprintf(`from(bucket:%q)
			|> range(start: %s)
			|> filter( fn: (r) =>
				(r._measurement == %q) and
				(r._field == %q)
			)`,
		cfg.Bucket,
		start,
		measurement,
		field,
	)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("creating query: %v", err)
	}

	for result.Next() {
		r := result.Record()
		v := r.Value()
		asFloat, ok := v.(int64)
		if !ok {
			return fmt.Errorf("value (%#v, %[1]T) is not a float64", v)
		}

		fmt.Printf("%s %d\n", r.Time().UTC().Format(time.RFC3339), asFloat)
	}
	if result.Err() != nil {
		return fmt.Errorf("query error: %s\n", result.Err().Error())
	}

	return nil
}
