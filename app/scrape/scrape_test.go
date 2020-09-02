package scrape_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/alcortesm/sputnik-popularity/app/pair"
	"github.com/alcortesm/sputnik-popularity/app/scrape"
)

func TestScrape(t *testing.T) {
	t.Parallel()

	subtests := map[string]func(t *testing.T){
		"returns correct pair if response is valid": correct,
		"handles errors from the HTTPer":            httperError,
		"handles non-OK responses from the httper":  httperNonOKResponse,
		"sends correct request":                     requestOK,
	}

	for name, testFunc := range subtests {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testFunc(t)
		})
	}
}

// logger returns a scrape.Logger that writes to Go's testing output.
func logger(t *testing.T) scrape.Logger {
	return log.New(testWriter{t}, "", 0)
}

// testWriter is a writer that writes to Go's testing output.
type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Logf("%s", string(p))
	return len(p), nil
}

type mockHTTPer struct {
	do func(*http.Request) (*http.Response, error)
}

func (m mockHTTPer) Do(r *http.Request) (*http.Response, error) {
	return m.do(r)
}

func correct(t *testing.T) {
	fix := struct {
		people    float64
		capacity  float64
		timestamp time.Time
	}{
		people:    42.0,
		capacity:  200.0,
		timestamp: time.Time{}.Add(time.Second),
	}

	// an httper that returns a valid response with the data we want.
	var httper scrape.HTTPer
	{
		data := fmt.Sprintf(`{"People": %.0f, "Capacity": %.0f}`,
			fix.people, fix.capacity)

		response := &http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(
				bytes.NewBufferString(data),
			),
		}

		httper = mockHTTPer{
			do: func(_ *http.Request) (*http.Response, error) {
				return response, nil
			},
		}
	}

	// a clock that returns a fixed moment in time
	clock := func() time.Time { return fix.timestamp }

	scraper := scrape.NewScraper(
		logger(t),
		httper,
		clock,
		scrape.Config{}, // irrelevant config
	)

	got, err := scraper.Scrape(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := &pair.Pair{
		Timestamp: fix.timestamp,
		Value:     fix.people / fix.capacity,
	}

	tolerance := 1e-02
	if !want.Equals(*got, tolerance) {
		t.Errorf("\nwant %v\n got %v", want, got)
	}
}

func httperError(t *testing.T) {
	cause := errors.New("some httper error")

	// an httper that returns an error
	httper := mockHTTPer{
		do: func(_ *http.Request) (*http.Response, error) {
			return nil, cause
		},
	}

	scraper := scrape.NewScraper(
		logger(t),
		httper,
		nil,             // irrelevant clock
		scrape.Config{}, // irrelevant config
	)

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("unexpected success")
	}

	if !strings.Contains(err.Error(), cause.Error()) {
		t.Errorf("cannot find cause (%v) in error: %v", cause, err)
	}
}

func httperNonOKResponse(t *testing.T) {
	fix := struct {
		statusCode int
		message    string
	}{
		statusCode: http.StatusTeapot,
		message:    "some server message",
	}

	// an httper that returns a non-200 response.
	var httper mockHTTPer
	{
		response := &http.Response{
			StatusCode: fix.statusCode,
			Body: ioutil.NopCloser(
				bytes.NewBufferString(fix.message),
			),
		}

		httper = mockHTTPer{
			do: func(_ *http.Request) (*http.Response, error) {
				return response, nil
			},
		}
	}

	scraper := scrape.NewScraper(
		logger(t),
		httper,
		nil,             // irrelevant clock
		scrape.Config{}, // irrelevant config
	)

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("unexpected success")
	}

	if !strings.Contains(err.Error(), fix.message) {
		t.Errorf("cannot find cause (%v) in error: %v", fix.message, err)
	}
}

func requestOK(t *testing.T) {
	type ctxKey string

	fix := struct {
		url     string
		gymName string
		gymID   int
		stop    error
		ctx     context.Context
	}{
		url:     "some_url",
		gymName: "some_gymName",
		gymID:   42,
		stop:    errors.New("stop here"),
		ctx: context.WithValue(
			context.Background(),
			ctxKey("key"),
			"value",
		),
	}

	// an httper that checks if the request is correct.
	httper := mockHTTPer{
		do: func(r *http.Request) (*http.Response, error) {
			{ // check method
				want := http.MethodPost
				got := r.Method
				if want != got {
					t.Errorf("method: want %s, got %s", want, got)
				}
			}

			{ // check url
				want := fix.url
				got := r.URL.String()
				if want != got {
					t.Errorf("URL:\nwant %s\n got %s", want, got)
				}
			}

			{ // check content type
				want := "application/json"
				got := r.Header.Get("Content-type")
				if want != got {
					t.Errorf("content type:\nwant %s\n got %s", want, got)
				}
			}

			{ // check body
				var want []byte
				{
					js := fmt.Sprintf(`{"Namespace": %q, "GymID": %d}`,
						fix.gymName, fix.gymID)

					want = []byte(js)
				}

				got, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				if diff := jsonDiff(t, want, got); diff != "" {
					t.Errorf("body: (-want +got)\n%s", diff)
				}
			}

			{ // check context
				got := r.Context()
				if fix.ctx != got {
					t.Errorf("context:\nwant %#v\n got %#v",
						fix.ctx, got)
				}
			}

			return nil, fix.stop
		},
	}

	config := scrape.Config{
		URL:     fix.url,
		GymName: fix.gymName,
		GymID:   fix.gymID,
	}
	scraper := scrape.NewScraper(
		logger(t),
		httper,
		nil, // irrelevant clock
		config,
	)

	_, err := scraper.Scrape(fix.ctx)
	if err == nil {
		t.Fatal("unexpected success")
	}

	if !strings.Contains(err.Error(), fix.stop.Error()) {
		t.Errorf("cannot find cause (%v) in error: %v", fix.stop, err)
	}
}

func jsonDiff(t *testing.T, want, got []byte) string {
	t.Helper()

	var dWant interface{}
	if err := json.Unmarshal(want, &dWant); err != nil {
		t.Fatalf("decoding want: %v", err)
	}

	var dGot interface{}
	fmt.Println(string(got))
	if err := json.Unmarshal(got, &dGot); err != nil {
		t.Fatalf("decoding got: %v", err)
	}

	return cmp.Diff(dWant, dGot)
}
