package config

type Config struct {
	InfluxDB InfluxDB
}

type InfluxDB struct {
	URL        string `required:"true"`
	TokenWrite string `required:"true" split_words:"true"`
	TokenRead  string `required:"true" split_words:"true"`
	Org        string `required:"true"`
	Bucket     string `required:"true"`
}
