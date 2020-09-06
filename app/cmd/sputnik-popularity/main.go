package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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

type Config struct {
	InfluxDB influx.Config
	Scrape   scrape.Config
}

func main() {
	logger := log.New(os.Stdout, "app ",
		log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)

	var config Config
	envPrefix := "SPUTNIK"
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		logger.Fatalf("failed to start app: processign env vars: %v", err)
	}

	signalCtx, cancel := signalContext(syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	scrapedData := make(chan *gym.Utilization)
	scrapeTicker := time.NewTicker(config.Scrape.Period)
	defer scrapeTicker.Stop()

	g, ctx := errgroup.WithContext(signalCtx)

	// launch a scraper of gym utilization data
	g.Go(func() error {
		return startScraping(
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

	logger.Println("starting app...")

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

func startScraping(
	ctx context.Context,
	logger *log.Logger,
	config scrape.Config,
	trigger <-chan time.Time,
	scraped chan<- *gym.Utilization,
) error {
	const prefix = "scraping"

	logger.Printf("%s: started...", prefix)
	defer logger.Printf("%s: stopped", prefix)

	client := &http.Client{Timeout: config.Timeout * time.Second}

	scraper := scrape.NewScraper(
		logger,
		client,
		time.Now,
		config,
	)

	for waitForTrigger := false; ; waitForTrigger = true {
		if waitForTrigger {
			// wait for a trigger or a cancelation of the context
			select {
			case _, ok := <-trigger:
				if !ok {
					return fmt.Errorf("%s: closed trigger channel", prefix)
				}
			case <-ctx.Done():
				return fmt.Errorf("%s: %w", prefix, ctx.Err())
			}
		}

		u, err := scraper.Scrape(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return fmt.Errorf("%s: %w", prefix, err)
			}

			logger.Printf("%s: %v\n", prefix, err)
			continue
		}

		scraped <- u
	}

	return nil
}

func processScrapedData(
	ctx context.Context,
	logger *log.Logger,
	config influx.Config,
	scraped <-chan *gym.Utilization,
) error {
	const prefix = "processing scraped data"

	logger.Printf("%s: started...\n", prefix)
	defer logger.Printf("%s: stopped\n", prefix)

	store, cancel := influx.NewStore(config)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", prefix, ctx.Err())
		case u, ok := <-scraped:
			if !ok {
				return fmt.Errorf("%s: closed data channel", prefix)
			}

			if err := store.Add(ctx, u); err != nil {
				logger.Printf("%s: adding to store: %v\n", prefix, err)
			}
		}
	}
}
