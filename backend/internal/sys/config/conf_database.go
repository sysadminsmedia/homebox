package config

const (
	DriverSqlite3 = "sqlite3"
)

type Storage struct {
	// Data is the path to the root directory
	PrefixPath string `yaml:"prefix_path" conf:"default:.data"`
	ConnString string `yaml:"conn_string" conf:"default:file:///./"`
}

type Database struct {
	Driver           string `yaml:"driver"             conf:"default:sqlite3"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Host             string `yaml:"host"`
	Port             string `yaml:"port"`
	Database         string `yaml:"database"`
	SslMode          string `yaml:"ssl_mode"`
	SqlitePath       string `yaml:"sqlite_path"        conf:"default:./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite"`
	PubSubConnString string `yaml:"pubsub_conn_string" conf:"default:mem://{{ .Topic }}"`
	SslRootCert      string `yaml:"sslrootcert"`
	SslCert          string `yaml:"sslcert"`
	SslKey           string `yaml:"sslkey"`
}
