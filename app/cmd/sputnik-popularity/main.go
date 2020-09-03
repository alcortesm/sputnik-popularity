package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/alcortesm/sputnik-popularity/app/influx"
	"github.com/alcortesm/sputnik-popularity/app/pair"
	"github.com/alcortesm/sputnik-popularity/app/scrape"
	"github.com/alcortesm/sputnik-popularity/app/web"
	"github.com/alcortesm/sputnik-popularity/pkg/httpdeco"
)

type Config struct {
	InfluxDB influx.Config
	Scrape   scrape.Config
}

func main() {
	ctx, cancel := signalContext(syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := log.New(os.Stdout, "",
		log.Ldate|log.Ltime|log.LUTC)

	logger.Println("starting app...")

	var config Config
	envPrefix := "SPUTNIK_POPULARITY"
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		logger.Fatalf("processign environment variables: %v", err)
	}

	/*
		if err := testScraper(ctx, logger, config.Scrape); err != nil {
			logger.Fatalf("testing scraper: %v", err)
		}
	*/

	capacity := 5

	popularity, err := web.NewPopularity(capacity)
	if err != nil {
		logger.Fatalf("creating popularity: %v", err)
	}

	http.Handle("/popularity.html", httpdeco.Decorate(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-type", "text/html")
				w.Write(popularity.HTML())
			},
		),
		httpdeco.WithLogs(logger),
	))

	http.Handle("/style.css", httpdeco.Decorate(
		web.StyleHandler(),
		httpdeco.WithLogs(logger),
	))

	http.Handle("/", httpdeco.Decorate(
		http.NotFoundHandler(),
		httpdeco.WithLogs(logger),
	))

	port := "8080"

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go addDemoValues(popularity, logger)

	go func() {
		logger.Printf("starting server at port %s...\n", port)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	log.Println("signal received: stopping server...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutting down server: %+v", err)
	}
}

func signalContext(signals ...os.Signal) (
	context.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(c)
	}()

	return ctx, cancel
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

func testScraper(
	ctx context.Context,
	logger *log.Logger,
	config scrape.Config,
) error {
	client := &http.Client{Timeout: 10 * time.Second}

	scraper := scrape.NewScraper(
		logger,
		client,
		time.Now,
		config,
	)

	pair, err := scraper.Scrape(ctx)
	if err != nil {
		return fmt.Errorf("scraping: %v", err)
	}

	fmt.Println(pair)

	return nil
}
