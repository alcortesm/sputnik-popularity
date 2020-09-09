package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/alcortesm/sputnik-popularity/pkg/httpdeco"
)

const (
	shutdownTimeoutSeconds   = 10
	readTimeoutSeconds       = 10
	writeTimeoutSeconds      = 10
	idleTimeoutSeconds       = 30
	readHeaderTimeoutSeconds = 2
)

type config struct {
	Port          int    `required:"true"`
	Capacity      uint64 `required:"true"`
	InitialPeople uint64 `default:"0" split_words:"true"`
}

func main() {
	logger := log.New(os.Stdout, "",
		log.Ldate|log.Ltime|log.LUTC)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	var config config
	envPrefix := "SCRAPEME"
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		logger.Fatalf("processign environment variables: %v", err)
	}

	web := &web{
		logger:   logger,
		people:   config.InitialPeople,
		capacity: config.Capacity,
	}

	http.Handle("/popularity", httpdeco.Decorate(
		web,
		httpdeco.WithLogs(logger),
	))

	http.Handle("/", httpdeco.Decorate(
		http.NotFoundHandler(),
		httpdeco.WithLogs(logger),
	))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		ReadTimeout:       readTimeoutSeconds * time.Second,
		WriteTimeout:      writeTimeoutSeconds * time.Second,
		IdleTimeout:       idleTimeoutSeconds * time.Second,
		ReadHeaderTimeout: readTimeoutSeconds * time.Second,
	}

	logger.Printf("starting server at port %d...\n", config.Port)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-done

	logger.Println("signal received: stopping server...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeoutSeconds*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutting down server: %+v", err)
	}
}

type web struct {
	logger   *log.Logger
	capacity uint64

	mutex  sync.Mutex
	people uint64
}

func (w *web) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.people = (w.people + 1) % w.capacity

	rw.Header().Set("Content-type", "application/json")

	payload := struct {
		People   uint64
		Capacity uint64
	}{
		People:   w.people,
		Capacity: w.capacity,
	}

	err := json.NewEncoder(rw).Encode(payload)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
