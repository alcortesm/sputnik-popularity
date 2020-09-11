package scrape

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

type Config struct {
	URL     string
	GymName string
	GymID   int
}

type Scraper struct {
	logger *log.Logger
	client HTTPer
	clock  Clock
	url    string
	body   string
}

type HTTPer interface {
	Do(*http.Request) (*http.Response, error)
}

type Clock func() time.Time

func NewScraper(
	logger *log.Logger,
	client HTTPer,
	clock Clock,
	config Config,
) *Scraper {
	body := fmt.Sprintf(`{"Namespace": %q, "GymID": %d}`,
		config.GymName, config.GymID,
	)

	return &Scraper{
		logger: logger,
		client: client,
		clock:  clock,
		url:    config.URL,
		body:   body,
	}
}

func (s *Scraper) Scrape(ctx context.Context) (*gym.Utilization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.url,
		strings.NewReader(s.body),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed POST %s: %v", s.url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(
				"status %d (%s); error reading response body: %v",
				resp.StatusCode, http.StatusText(resp.StatusCode), err)
		}

		return nil, fmt.Errorf(
			"unsuccessful response: status %d (%s); body: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode), body,
		)
	}

	var response struct {
		People   uint64
		Capacity uint64
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %v", err)
	}

	if response.Capacity == 0 {
		return nil, fmt.Errorf("ignoring server response with zero capacity")
	}

	result := &gym.Utilization{
		Timestamp: s.clock(),
		People:    response.People,
		Capacity:  response.Capacity,
	}

	return result, nil
}
