package httpdeco

import (
	"net/http"
	"time"
)

// Logger knows how to log messages.
type Logger interface {
	Printf(string, ...interface{})
}

func WithLogs(l Logger) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			verbose := &verboseResponseWriter{ResponseWriter: w}

			start := time.Now()
			h.ServeHTTP(verbose, r)
			elapsed := time.Since(start)

			if verbose.writeError != nil {
				l.Printf("%s %s - %d %dns %v", r.Method, r.URL,
					verbose.status, elapsed.Nanoseconds(),
					verbose.writeError)
				return
			}

			l.Printf("%s %s - %d %dns", r.Method, r.URL,
				verbose.status, elapsed.Nanoseconds())
		})
	}
}

// VerboseResponseWriter wraps an http.ResponseWriter so you can
// inspect the status code and the write error after writing
// the response.
//
// Note this will hide optional methods in the http.ResponseWriter like
// http.Flusher or http.Hijacker.
type verboseResponseWriter struct {
	http.ResponseWriter
	status     int   // the status code set by the handler
	writeError error // the error returned by the last call to Write
}

func (w *verboseResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *verboseResponseWriter) Write(b []byte) (int, error) {
	// If WriteHeader has not yet been called, Write sets
	// status to http.StatusOK before writing the data.
	if w.status == 0 {
		w.status = http.StatusOK
	}

	var n int
	n, w.writeError = w.ResponseWriter.Write(b)

	return n, w.writeError
}
