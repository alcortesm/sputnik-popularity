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
	"github.com/alcortesm/sputnik-popularity/app/recent"
	"github.com/alcortesm/sputnik-popularity/app/scrape"
	"github.com/alcortesm/sputnik-popularity/app/web"
	"github.com/alcortesm/sputnik-popularity/pkg/httpdeco"
)

type Config struct {
	InfluxDB        influx.Config
	Scrape          scrape.Config
	RecentRetention time.Duration `default:"168h" split_words:"true"`
	Web             Web
}

type Web struct {
	Port            int           `default:"8080"`
	ReadTimeout     time.Duration `default:"10s"`
	WriteTimeout    time.Duration `default:"10s"`
	ShutdownTimeout time.Duration `default:"10s"`
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

	latest, err := recent.NewCache(config.RecentRetention)
	if err != nil {
		logger.Fatalf("failed to start app: creating a recent.Cache: %v", err)
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
			latest,
		)
	})

	// launch the web server
	g.Go(func() error {
		return launchWebServer(
			ctx,
			logger,
			config.Web,
			latest,
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

	logger.Printf("%s: starting...", prefix)
	defer logger.Printf("%s: stopped", prefix)

	client := &http.Client{Timeout: config.Timeout * time.Second}

	scraper := scrape.NewScraper(
		logger,
		client,
		time.Now,
		config,
	)

	do := func() {
		u, err := scraper.Scrape(ctx)
		if err != nil {
			logger.Printf("%s: %v\n", prefix, err)
			return
		}

		fmt.Println(u)
		scraped <- u
	}

	do()

	for {
		// wait for a trigger or a cancelation of the context
		select {
		case _, ok := <-trigger:
			if !ok {
				return fmt.Errorf("%s: closed trigger channel", prefix)
			}
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", prefix, ctx.Err())
		}

		do()
	}

	return nil
}

func processScrapedData(
	ctx context.Context,
	logger *log.Logger,
	config influx.Config,
	scraped <-chan *gym.Utilization,
	latest Adder,
) error {
	const prefix = "processing scraped data"

	logger.Printf("%s: starting...\n", prefix)
	defer logger.Printf("%s: stopped\n", prefix)

	store, cancel := influx.NewStore(config)
	defer cancel()

	do := func(u *gym.Utilization) {
		go func() {
			if err := store.Add(ctx, u); err != nil {
				logger.Printf("%s: adding to store: %v\n", prefix, err)
			}
		}()

		go latest.Add(u)
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", prefix, ctx.Err())
		case u, ok := <-scraped:
			if !ok {
				return fmt.Errorf("%s: closed data channel", prefix)
			}

			do(u)
		}
	}
}

// Adder know how to store gym Utilization data, for example, a
// recent.Cache.
//
// TODO: this should really be like the influx.Store.Add method instead,
// with context and returning and error.
type Adder interface {
	Add(...*gym.Utilization)
}

func launchWebServer(
	ctx context.Context,
	logger *log.Logger,
	config Web,
	latest web.RecentGetter,
) error {
	const prefix = "web server"

	logger.Printf("%s: starting at port %d...\n", prefix, config.Port)
	defer logger.Printf("%s: stopped\n", prefix)

	w := web.Web{Recent: latest}

	http.Handle("/popularity.html", httpdeco.Decorate(
		w.Handler(),
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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			config.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("shutting down server: %+v", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %s\n", err)
	}

	return nil
}
