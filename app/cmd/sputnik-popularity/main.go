package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	"github.com/alcortesm/sputnik-popularity/app/gym"
	"github.com/alcortesm/sputnik-popularity/app/influx"
	"github.com/alcortesm/sputnik-popularity/app/scrape"
)

const (
	influxMeasurement = "capacity_utilization"
	influxField       = "percent"
)

type Config struct {
	InfluxDB influx.Config
	Scrape   scrape.Config
}

func main() {
	logger := log.New(os.Stdout, "app ",
		log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)

	logger.Println("starting app...")

	var config Config
	envPrefix := "SPUTNIK_POPULARITY"
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		logger.Fatalf("processign environment variables: %v", err)
	}

	signalCtx, cancel := signalContext(syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	scrapedData := make(chan *gym.Utilization)
	scrapeTicker := time.NewTicker(config.Scrape.Period)
	defer scrapeTicker.Stop()

	g, ctx := errgroup.WithContext(signalCtx)

	// launch a scraper of gym utilization data
	g.Go(func() error {
		return scrape.Run(
			ctx,
			logger,
			config.Scrape,
			scrapeTicker.C,
			scrapedData,
		)
	})

	// launch a processor for the scraped data
	g.Go(func() error {
		return processScrapedData(
			ctx,
			logger,
			config.InfluxDB,
			scrapedData,
		)
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Printf("stopping the app: signal received\n")
		} else {
			logger.Printf("stopping the app: %v\n", err)
		}
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

func processScrapedData(
	ctx context.Context,
	logger *log.Logger,
	config influx.Config,
	scraped <-chan *gym.Utilization,
) error {
	logger.Println("start processing scraped data...")
	defer logger.Println("stopped processing scraped data")

	store, cancel := influx.NewStore(
		config,
		influxMeasurement,
	)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case u, ok := <-scraped:
			if !ok {
				return fmt.Errorf("closed scraped data channel")
			}

			if err := store.Add(ctx, u); err != nil {
				logger.Printf("adding to store: %v\n", err)
			}
		}
	}
}
