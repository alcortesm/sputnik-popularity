package influx

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/log"

	"github.com/alcortesm/sputnik-popularity/pkg/pair"
)

func init() {
	log.Log = nil
}

type Config struct {
	URL        string `required:"true"`
	TokenWrite string `required:"true" split_words:"true"`
	TokenRead  string `required:"true" split_words:"true"`
	Org        string `required:"true"`
	Bucket     string `required:"true"`
}

type Store struct {
	measurement string
	field       string
	config      Config
	bucket      string
	writeAPI    api.WriteAPIBlocking
	queryAPI    api.QueryAPI
}

func NewStore(
	config Config,
	measurement string,
	field string,
) (store *Store, cancel func()) {
	wc := influxdb2.NewClient(config.URL, config.TokenWrite)
	rc := influxdb2.NewClient(config.URL, config.TokenRead)

	store = &Store{
		measurement: measurement,
		field:       field,
		config:      config,
		writeAPI:    wc.WriteAPIBlocking(config.Org, config.Bucket),
		queryAPI:    rc.QueryAPI(config.Org),
	}

	cancel = func() {
		wc.Close()
		rc.Close()
	}

	return store, cancel
}

func (s *Store) Add(ctx context.Context, pairs ...pair.Pair) error {
	points := make([]*write.Point, len(pairs))
	{
		tags := map[string]string(nil)

		for i, p := range pairs {
			fields := map[string]interface{}{s.field: p.Value}
			points[i] = influxdb2.NewPoint(
				s.measurement,
				tags,
				fields,
				p.Timestamp.UTC(),
			)
		}
	}

	if err := s.writeAPI.WritePoint(ctx, points...); err != nil {
		return fmt.Errorf("writing points: %v", err)
	}

	return nil
}

func (s *Store) Get(
	ctx context.Context,
	since time.Time,
) ([]pair.Pair, error) {
	query := fmt.Sprintf(`from(bucket:%q)
			|> range(start: %s)
			|> filter( fn: (r) =>
				(r._measurement == %q) and
				(r._field == %q)
			)`,
		s.config.Bucket,
		since.UTC().Format(time.RFC3339),
		s.measurement,
		s.field,
	)

	table, err := s.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}

	result := []pair.Pair{}

	for table.Next() {
		r := table.Record()
		t := r.Time().UTC()
		v := r.Value()

		asFloat, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf(
				"value (%#v, %[1]T) at time %s is not a float64",
				v, t.Format(time.RFC3339))
		}

		result = append(result, pair.Pair{
			Value:     asFloat,
			Timestamp: t,
		})
	}

	if err := table.Err(); err != nil {
		return nil, fmt.Errorf("table error: %s", err)
	}

	return result, nil
}
