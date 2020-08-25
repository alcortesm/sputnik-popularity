package config

import "github.com/alcortesm/sputnik-popularity/influx"

type Config struct {
	InfluxDB influx.Config
}
