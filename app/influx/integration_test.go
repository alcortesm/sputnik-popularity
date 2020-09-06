// +build integration

package influx_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/alcortesm/sputnik-popularity/app/gym"
	"github.com/alcortesm/sputnik-popularity/app/influx"
)

const (
	dbURL                     = "http://influxdb:9999"
	readyTimeoutSeconds       = 120
	readyIntervalMilliseconds = 200
	org                       = "test_org"
	bucket                    = "test_bucket"
)

// 2020-01-01 00:00:00 +0000 UTC
var year2020 = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

// the authorization token to talk with the DB, written by TestMain and
// used inside the test functions.
var token string

// This tests assume there is an InfluxDB instance running at
// influxdb:9999, see the integration folder at the root of the
// project for more details.
func TestMain(m *testing.M) {
	timeout := readyTimeoutSeconds * time.Second
	interval := readyIntervalMilliseconds * time.Millisecond

	if err := waitForInflux(interval, timeout); err != nil {
		log.Fatalf("waiting for InfluxDB to be ready: %v", err)
	}

	var err error
	token, err = influxCreateOrgBucket()
	if err != nil {
		log.Fatalf("creating InfluxDB org and bucket: %v", err)
	}

	fmt.Println("debug info:")
	fmt.Printf("\torg: %s\n", org)
	fmt.Printf("\tbucket: %s\n", bucket)
	fmt.Printf("\ttoken: %s\n", token)

	tip1 := fmt.Sprintf(`$ docker start integration_influxdb_1`)
	tip2 := fmt.Sprintf(`$ docker run --rm --network 'integration_default' -it quay.io/influxdb/influxdb:2.0.0-beta influx query 'from(bucket:%q) |> range(start: %s)' --org '%s' --host '%s' -t '%s'`, bucket, year2020.Format(time.RFC3339), org, dbURL, token)
	fmt.Printf("debug tip: show database contents:\n\t%s\n\t%s\n", tip1, tip2)

	os.Exit(m.Run())
}

// createOrgBucket creates a new org and bucket in the database and
// returns the token to access it.
// waitForInflux will wait if InfluxDB is ready by pulling at the given
// intervals.  It will return an error if it is not ready after the
// timeout.
func waitForInflux(interval, timeout time.Duration) error {
	noToken := ""
	client := influxdb2.NewClient(dbURL, noToken)
	defer client.Close()

	start := time.Now()

	lastError := (error)(nil)

	for {
		if time.Since(start) > timeout {
			if lastError != nil {
				return fmt.Errorf("timeout, last error was: %v", lastError)
			}
			return errors.New("timeout")
		}

		var ok bool
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		ok, lastError = client.Ready(ctx)
		cancel()

		if ok {
			break
		}

		deadline, _ := ctx.Deadline()
		remaining := deadline.Sub(time.Now())
		time.Sleep(remaining)
	}

	return nil
}

// influxCreateOrgBucket creates an org and bucket in the DB and returns
// the authorization token to use them.
func influxCreateOrgBucket() (string, error) {
	const (
		user     = "test_user"
		password = "test_password"
	)

	emptyToken := ""
	client := influxdb2.NewClient(dbURL, emptyToken)
	defer client.Close()

	timeout := readyTimeoutSeconds * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	retentionHours := 0 // forever

	setup, err := client.Setup(
		ctx,
		user,
		password,
		org,
		bucket,
		retentionHours,
	)
	if err != nil {
		return "", fmt.Errorf("influxdb setup: %v", err)
	}

	return *setup.Auth.Token, nil
}

func TestInflux_AddGet(t *testing.T) {
	t.Parallel()

	fix := struct {
		measurement string
		field       string
		start       time.Time
		timeout     time.Duration
	}{
		measurement: "m_" + t.Name(),
		field:       "test_field",
		start:       year2020,
		timeout:     10 * time.Second,
	}

	t1 := fix.start.Add(1 * time.Second)
	t2 := fix.start.Add(2 * time.Second)
	t3 := fix.start.Add(3 * time.Second)

	data := []*gym.Utilization{
		{Timestamp: t1, People: 1, Capacity: 42},
		{Timestamp: t2, People: 2, Capacity: 42},
		{Timestamp: t3, People: 3, Capacity: 42},
	}

	// Add some utilization data to the DB
	{
		store, cancel := influx.NewStore(
			influx.Config{
				URL:         dbURL,
				Org:         org,
				TokenWrite:  token,
				Bucket:      bucket,
				Measurement: fix.measurement,
			},
		)
		t.Cleanup(cancel)

		ctx, cancel := context.WithTimeout(
			context.Background(), fix.timeout)
		t.Cleanup(cancel)

		if err := store.Add(ctx, data...); err != nil {
			t.Fatal(err)
		}
	}

	// get the utilization data from the DB
	var got []*gym.Utilization
	{
		store, cancel := influx.NewStore(
			influx.Config{
				URL:         dbURL,
				Org:         org,
				TokenRead:   token,
				Bucket:      bucket,
				Measurement: fix.measurement,
			},
		)
		t.Cleanup(cancel)

		ctx, cancel := context.WithTimeout(
			context.Background(), fix.timeout)
		t.Cleanup(cancel)
		var err error
		got, err = store.Get(ctx, fix.start)
		if err != nil {
			t.Fatal(err)
		}
	}

	if diff := cmp.Diff(data, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}
