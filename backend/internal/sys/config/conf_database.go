package config

const (
	DriverSqlite3 = "sqlite3"
)

type Storage struct {
	// Data is the path to the root directory
	ConnString string `yaml:"conn_string" conf:"default:file://./.data"`
	Data       string `yaml:"data" conf:"default:./"`
}

type Database struct {
	Driver     string `yaml:"driver"      conf:"default:sqlite3"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Database   string `yaml:"database"`
	SslMode    string `yaml:"ssl_mode"`
	SqlitePath string `yaml:"sqlite_path" conf:"default:./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1"`
}
