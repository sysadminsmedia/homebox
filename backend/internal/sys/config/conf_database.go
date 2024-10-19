package config

const (
	DriverSqlite3 = "sqlite3"
)

type Storage struct {
	// Data is the path to the root directory
	Data      string `yaml:"data"       conf:"default:./.data"`
	SqliteURL string `yaml:"sqlite-url" conf:"default:./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1"`
}

type Database struct {
	Driver     string `yaml:"driver" conf:"default:sqlite"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Database   string `yaml:"database"`
	SslMode    string `yaml:"ssl-mode"`
	SqlitePath string `yaml:"sqlite-path" conf:"default:./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1"`
}
