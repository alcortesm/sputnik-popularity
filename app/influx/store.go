package influx

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/log"

	"github.com/alcortesm/sputnik-popularity/app/gym"
)

func init() {
	log.Log = nil
}

type Config struct {
	URL         string `required:"true"`
	TokenWrite  string `required:"true" split_words:"true"`
	TokenRead   string `required:"true" split_words:"true"`
	Org         string `default:"tsDemo"`
	Bucket      string `default:"sputnik_popularity"`
	Measurement string `default:"capacity_utilization"`
}

const (
	peopleFieldKey   = "people"
	capacityFieldKey = "capacity"
)

type Store struct {
	config   Config
	bucket   string
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
}

func NewStore(config Config) (store *Store, cancel func()) {
	opts := influxdb2.DefaultOptions().
		SetPrecision(time.Second)

	wc := influxdb2.NewClientWithOptions(
		config.URL,
		config.TokenWrite,
		opts,
	)

	rc := influxdb2.NewClientWithOptions(
		config.URL,
		config.TokenRead,
		opts,
	)

	store = &Store{
		config:   config,
		writeAPI: wc.WriteAPIBlocking(config.Org, config.Bucket),
		queryAPI: rc.QueryAPI(config.Org),
	}

	cancel = func() {
		wc.Close()
		rc.Close()
	}

	return store, cancel
}

func (s *Store) Add(ctx context.Context, data ...*gym.Utilization) error {
	points := make([]*write.Point, len(data))
	{
		for i, d := range data {
			noTags := map[string]string(nil)

			fields := map[string]interface{}{
				peopleFieldKey:   d.People,
				capacityFieldKey: d.Capacity,
			}

			points[i] = influxdb2.NewPoint(
				s.config.Measurement,
				noTags,
				fields,
				d.Timestamp,
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
) ([]*gym.Utilization, error) {
	query := fmt.Sprintf(`from(bucket:%q)
			|> range(start: %s)
			|> filter( fn: (r) =>
				(r._measurement == %q) and
				(
					(r._field == %q) or
					(r._field == %q)
				)
			)
			|> pivot(
				rowKey:["_time"],
				columnKey:["_field"],
				valueColumn: "_value"
			)`,
		s.config.Bucket,
		since.Format(time.RFC3339),
		s.config.Measurement,
		peopleFieldKey,
		capacityFieldKey,
	)

	table, err := s.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}

	result := []*gym.Utilization{}

	for table.Next() {
		u, err := recordToUtilization(table.Record())
		if err != nil {
			return nil, fmt.Errorf("invalid influx record: %v", err)
		}

		result = append(result, u)
	}

	if err := table.Err(); err != nil {
		return nil, fmt.Errorf("table error: %s", err)
	}

	return result, nil
}

func recordToUtilization(r *query.FluxRecord) (*gym.Utilization, error) {
	result := &gym.Utilization{
		Timestamp: r.Time(),
	}

	raw := r.ValueByKey(peopleFieldKey)
	var err error

	result.People, err = toUint64(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing %s field value at %s: %v",
			peopleFieldKey,
			result.Timestamp.Format(time.RFC3339), err)
	}

	raw = r.ValueByKey(capacityFieldKey)

	result.Capacity, err = toUint64(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing %s field value at %s: %v",
			capacityFieldKey,
			result.Timestamp.Format(time.RFC3339), err)
	}

	if result.Capacity == 0 {
		return nil, fmt.Errorf("capacity at %s is 0",
			result.Timestamp.Format(time.RFC3339),
		)
	}

	return result, nil
}

func toUint64(v interface{}) (uint64, error) {
	result, ok := v.(uint64)
	if !ok {
		return 0, fmt.Errorf("want uint64, got %T instead", v)
	}

	return result, nil
}
