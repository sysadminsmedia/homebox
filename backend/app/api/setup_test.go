package main

import (
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestSetupDatabaseURL_PostgresConnString(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Database
		expectedURL string
		expectError bool
	}{
		{
			name: "connection string takes priority",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "postgres://user:pass@localhost:5432/testdb?sslmode=require",
				Host:       "should-not-be-used.com",
				Port:       "9999",
				Database:   "should-not-be-used",
				SslMode:    "disable",
				Username:   "ignored",
				Password:   "ignored",
			},
			expectedURL: "postgres://user:pass@localhost:5432/testdb?sslmode=require",
			expectError: false,
		},
		{
			name: "key/value connection string format",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "host=localhost port=5432 dbname=testdb user=testuser sslmode=disable",
				Host:       "ignored",
				Port:       "ignored",
				Database:   "ignored",
			},
			expectedURL: "host=localhost port=5432 dbname=testdb user=testuser sslmode=disable",
			expectError: false,
		},
		{
			name: "individual fields when conn string is empty",
			config: config.Database{
				Driver:   config.DriverPostgres,
				Host:     "localhost",
				Port:     "5432",
				Database: "testdb",
				SslMode:  "disable",
			},
			expectedURL: "host=localhost port=5432 dbname=testdb sslmode=disable",
			expectError: false,
		},
		{
			name: "individual fields with username and password",
			config: config.Database{
				Driver:   config.DriverPostgres,
				Host:     "db.example.com",
				Port:     "5432",
				Database: "production",
				SslMode:  "require",
				Username: "dbuser",
				Password: "secret",
			},
			expectedURL: "host=db.example.com port=5432 dbname=production sslmode=require user=dbuser password=secret",
			expectError: false,
		},
		{
			name: "connection string with SSL parameters",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "postgres://user:pass@localhost:5432/testdb?sslmode=require&sslrootcert=/path/to/root.crt",
			},
			expectedURL: "postgres://user:pass@localhost:5432/testdb?sslmode=require&sslrootcert=/path/to/root.crt",
			expectError: false,
		},
		{
			name: "cloud provider connection string (AWS RDS)",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "postgres://user:pass@db-instance.abc123.us-east-1.rds.amazonaws.com:5432/mydb?sslmode=require",
			},
			expectedURL: "postgres://user:pass@db-instance.abc123.us-east-1.rds.amazonaws.com:5432/mydb?sslmode=require",
			expectError: false,
		},
		{
			name: "connection string has priority over individual fields",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "postgres://connstring:priority@localhost:5432/connstringdb?sslmode=require",
				Host:       "individual-fields.com",
				Port:       "1111",
				Database:   "individualdb",
				SslMode:    "disable",
				Username:   "individualuser",
				Password:   "individualpass",
			},
			expectedURL: "postgres://connstring:priority@localhost:5432/connstringdb?sslmode=require",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Database: tt.config,
			}

			url, err := setupDatabaseURL(cfg)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if url != tt.expectedURL {
				t.Errorf("expected URL:\n  %s\ngot:\n  %s", tt.expectedURL, url)
			}
		})
	}
}

func TestSetupDatabaseURL_Sqlite(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Database
		expectedURL string
		expectError bool
	}{
		{
			name: "SQLite with query parameters",
			config: config.Database{
				Driver:     config.DriverSqlite3,
				SqlitePath: "./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite",
			},
			expectedURL: "./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite",
			expectError: false,
		},
		{
			name: "SQLite simple path",
			config: config.Database{
				Driver:     config.DriverSqlite3,
				SqlitePath: "file:ent?mode=memory&cache=shared&_fk=1",
			},
			expectedURL: "file:ent?mode=memory&cache=shared&_fk=1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Database: tt.config,
			}

			url, err := setupDatabaseURL(cfg)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if url != tt.expectedURL {
				t.Errorf("expected URL:\n  %s\ngot:\n  %s", tt.expectedURL, url)
			}
		})
	}
}

func TestSetupDatabaseURL_UnsupportedDriver(t *testing.T) {
	cfg := &config.Config{
		Database: config.Database{
			Driver: "mysql",
		},
	}

	_, err := setupDatabaseURL(cfg)
	if err == nil {
		t.Errorf("expected error for unsupported driver but got none")
	}
}

func TestSetupDatabaseURL_PostgresSSLFileErrors(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Database
		expectError bool
	}{
		{
			name: "non-existent SSL root cert",
			config: config.Database{
				Driver:      config.DriverPostgres,
				Host:        "localhost",
				Port:        "5432",
				Database:    "testdb",
				SslMode:     "require",
				SslRootCert: "/nonexistent/path/to/root.crt",
			},
			expectError: true,
		},
		{
			name: "non-existent SSL cert",
			config: config.Database{
				Driver:   config.DriverPostgres,
				Host:     "localhost",
				Port:     "5432",
				Database: "testdb",
				SslMode:  "require",
				SslCert:  "/nonexistent/path/to/cert.crt",
			},
			expectError: true,
		},
		{
			name: "non-existent SSL key",
			config: config.Database{
				Driver:   config.DriverPostgres,
				Host:     "localhost",
				Port:     "5432",
				Database: "testdb",
				SslMode:  "require",
				SslKey:   "/nonexistent/path/to/key.key",
			},
			expectError: true,
		},
		{
			name: "connection string bypasses SSL file validation",
			config: config.Database{
				Driver:     config.DriverPostgres,
				ConnString: "postgres://user:pass@localhost:5432/testdb?sslmode=require&sslrootcert=/nonexistent/root.crt&sslcert=/nonexistent/cert.crt&sslkey=/nonexistent/key.key",
			},
			expectError: false, // Connection string is used as-is, no validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Database: tt.config,
			}

			_, err := setupDatabaseURL(cfg)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSetupDatabaseURL_BackwardCompatibility(t *testing.T) {
	// Test that existing configurations without ConnString still work
	cfg := &config.Config{
		Database: config.Database{
			Driver:   config.DriverPostgres,
			Host:     "localhost",
			Port:     "5432",
			Database: "homebox",
			SslMode:  "disable",
			Username: "homebox",
			Password: "homebox",
		},
	}

	expectedURL := "host=localhost port=5432 dbname=homebox sslmode=disable user=homebox password=homebox"

	url, err := setupDatabaseURL(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url != expectedURL {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expectedURL, url)
	}
}

func TestSetupDatabaseURL_PriorityOrder(t *testing.T) {
	// Test that ConnString takes priority even when all individual fields are set
	cfg := &config.Config{
		Database: config.Database{
			Driver:     config.DriverPostgres,
			ConnString: "postgres://priority:user@priority-host:9999/priority-db?sslmode=require",
			Host:       "fallback-host",
			Port:       "5432",
			Database:   "fallback-db",
			SslMode:    "disable",
			Username:   "fallback-user",
			Password:   "fallback-pass",
		},
	}

	expectedURL := "postgres://priority:user@priority-host:9999/priority-db?sslmode=require"

	url, err := setupDatabaseURL(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url != expectedURL {
		t.Errorf("expected:\n  %s\ngot:\n  %s", expectedURL, url)
	}
}
