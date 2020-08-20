package config

type Config struct {
	InfluxDB InfluxDB
}

type InfluxDB struct {
	URL    string `required:"true"`
	Token  string `required:"true"`
	Org    string `required:"true"`
	Bucket string `required:"true"`
}
