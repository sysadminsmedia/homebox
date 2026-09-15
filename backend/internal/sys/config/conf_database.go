package config

import "encoding/json"

const (
	DriverSqlite3  = "sqlite3"
	DriverPostgres = "postgres"
)

type Storage struct {
	// Data is the path to the root directory
	PrefixPath string `yaml:"prefix_path" conf:"default:.data"`
	ConnString string `yaml:"conn_string" conf:"default:file:///./"`
}

func (s Storage) MarshalJSON() ([]byte, error) {
	type alias Storage
	a := alias(s)
	a.ConnString = redactURLUserinfo(a.ConnString)
	return json.Marshal(a)
}

type Database struct {
	Driver           string `yaml:"driver"             conf:"default:sqlite3"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Host             string `yaml:"host"`
	Port             string `yaml:"port"`
	Database         string `yaml:"database"`
	SslMode          string `yaml:"ssl_mode"           conf:"default:require"`
	SslRootCert      string `yaml:"ssl_rootcert"`
	SslCert          string `yaml:"ssl_cert"`
	SslKey           string `yaml:"ssl_key"`
	SqlitePath       string `yaml:"sqlite_path"        conf:"default:./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite"`
	PubSubConnString string `yaml:"pubsub_conn_string" conf:"default:mem://{{ .Topic }}"`
}

func (d Database) MarshalJSON() ([]byte, error) {
	type alias Database
	a := alias(d)
	if a.Password != "" {
		a.Password = redactedValue
	}
	a.PubSubConnString = redactURLUserinfo(a.PubSubConnString)
	return json.Marshal(a)
}
