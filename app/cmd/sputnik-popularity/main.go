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
	// TODO: Influx and Scrape config should be defined here so the env
	// is straighforward and to avoid env tags in inner packages.
	InfluxDB influx.Config
	Scrape   scrape.Config
	Recent   RecentConfig
	Web      WebConfig
	Refresh  RefreshConfig
}

type RecentConfig struct {
	Retention time.Duration `default:"168h" split_words:"true"` // 168h is 1 week
}

type WebConfig struct {
	Port            int           `default:"8080"`
	ReadTimeout     time.Duration `default:"10s"`
	WriteTimeout    time.Duration `default:"10s"`
	ShutdownTimeout time.Duration `default:"10s"`
}

type RefreshConfig struct {
	Period time.Duration `default:"1h" split_words:"true"`
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

	recentStore, err := recent.NewStore(config.Recent.Retention)
	if err != nil {
		logger.Fatalf("failed to start app: creating a recent store: %v", err)
	}

	influxStore, cancel := influx.NewStore(config.InfluxDB)
	defer cancel()

	// channel where the scraper sends the scraped data
	scrapedCh := make(chan *gym.Utilization)

	g, ctx := errgroup.WithContext(signalCtx)

	// launch a scraper of gym utilization data
	g.Go(func() error {
		return startScraping(
			ctx,
			logger,
			config.Scrape,
			time.Tick(config.Scrape.Period),
			scrapedCh,
		)
	})

	// launch a processor for the scraped data
	g.Go(func() error {
		return processScrapedData(
			ctx,
			logger,
			scrapedCh,
			influxStore,
			recentStore,
		)
	})

	// launch the web server
	g.Go(func() error {
		return launchWebServer(
			ctx,
			logger,
			config.Web,
			recentStore,
		)
	})

	// refresh the recent from the DB regularly
	g.Go(func() error {
		return refreshRecentWithDB(
			ctx,
			logger,
			config.Recent.Retention,
			influxStore,
			recentStore,
			time.Tick(config.Refresh.Period),
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
	scraped <-chan *gym.Utilization,
	influxStore *influx.Store,
	recentStore *recent.Store,
) error {
	const prefix = "processing scraped data"

	logger.Printf("%s: starting...\n", prefix)
	defer logger.Printf("%s: stopped\n", prefix)

	do := func(u *gym.Utilization) {
		go func() {
			if err := influxStore.Add(ctx, u); err != nil {
				logger.Printf("%s: adding to influx store: %v\n",
					prefix, err)
			}
		}()

		go func() {
			if err := recentStore.Add(ctx, u); err != nil {
				logger.Printf("%s: adding to recent store: %v\n",
					prefix, err)
			}
		}()
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

func launchWebServer(
	ctx context.Context,
	logger *log.Logger,
	config WebConfig,
	latest web.Getter,
) error {
	const prefix = "web server"

	logger.Printf("%s: starting at port %d...\n", prefix, config.Port)
	defer logger.Printf("%s: stopped\n", prefix)

	w := web.Web{Recent: latest}

	http.Handle("/popularity.html", httpdeco.Decorate(
		w.PopularityHandler(),
		httpdeco.WithLogs(logger),
	))

	http.Handle("/chart.js", httpdeco.Decorate(
		w.ChartHandler(),
		httpdeco.WithLogs(logger),
	))

	http.Handle("/style.css", httpdeco.Decorate(
		w.StyleHandler(),
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

func refreshRecentWithDB(
	ctx context.Context,
	logger *log.Logger,
	retention time.Duration,
	store getSincer,
	recentStore *recent.Store,
	trigger <-chan time.Time,
) error {
	const prefix = "recent refresher"

	logger.Printf("%s: starting...\n", prefix)
	defer logger.Printf("%s: stopped\n", prefix)

	do := func() {
		since := time.Now().Add(-retention)

		data, err := store.Get(ctx, since)
		if err != nil {
			logger.Printf("%s: getting data from influx: %v\n", prefix, err)
			return
		}

		if err := recentStore.Add(ctx, data...); err != nil {
			logger.Printf("%s: adding data to recent: %v\n", prefix, err)
			return
		}
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

// getSincer knows how to retrieve gym utilization data since a certain
// date. See influx.Store for example.
type getSincer interface {
	Get(context.Context, time.Time) ([]*gym.Utilization, error)
}
