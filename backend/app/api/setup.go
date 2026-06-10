package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/otel"
)

// setupDatabase prepares the storage directory, validates the driver/SSL
// config, opens the database (through the OTel provider for tracing), runs
// any pending goose migrations, and returns an ent client along with the
// resolved ent dialect name. Caller owns the client and must Close() it.
// Extracted from run() to keep that function under the gocyclo threshold.
func setupDatabase(cfg *config.Config, otelProvider *otel.Provider) (*ent.Client, string, error) {
	if err := setupStorageDir(cfg); err != nil {
		return nil, "", err
	}

	driver := strings.ToLower(cfg.Database.Driver)
	if driver == config.DriverPostgres {
		if !validatePostgresSSLMode(cfg.Database.SslMode) {
			log.Error().Str("sslmode", cfg.Database.SslMode).Msg("invalid sslmode")
			return nil, "", fmt.Errorf("invalid sslmode: %s", cfg.Database.SslMode)
		}
	}

	databaseURL, err := setupDatabaseURL(cfg)
	if err != nil {
		return nil, "", err
	}

	driverName, dialectName, err := resolveDriver(driver)
	if err != nil {
		return nil, "", err
	}

	db, err := otelProvider.OpenDatabase(driverName, databaseURL)
	if err != nil {
		log.Error().
			Err(err).
			Str("driver", driver).
			Str("host", cfg.Database.Host).
			Str("port", cfg.Database.Port).
			Str("database", cfg.Database.Database).
			Msg("failed opening connection to {driver} database at {host}:{port}/{database}")
		return nil, "", fmt.Errorf("failed opening connection to %s database at %s:%s/%s: %w",
			driver, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, err)
	}

	c := ent.NewClient(ent.Driver(entsql.OpenDB(dialectName, db)))

	if err := runMigrations(c, driver); err != nil {
		_ = c.Close()
		return nil, "", err
	}
	return c, dialectName, nil
}

// resolveDriver maps the configured driver name onto the (database/sql, ent
// dialect) pair we use throughout the app.
func resolveDriver(driver string) (sqlDriverName, entDialect string, err error) {
	switch driver {
	case config.DriverPostgres:
		return "pgx", dialect.Postgres, nil
	case config.DriverSqlite3, "sqlite":
		return "sqlite3", dialect.SQLite, nil
	default:
		return "", "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

// runMigrations applies all pending goose migrations for the given driver.
func runMigrations(c *ent.Client, driver string) error {
	migrationsFs, err := migrations.Migrations(driver)
	if err != nil {
		return fmt.Errorf("failed to get migrations for %s: %w", driver, err)
	}
	goose.SetBaseFS(migrationsFs)
	if err := goose.SetDialect(driver); err != nil {
		log.Error().Str("driver", driver).Msg("unsupported database driver")
		return fmt.Errorf("unsupported database driver: %s", driver)
	}
	if err := goose.Up(c.Sql(), driver); err != nil {
		log.Error().Err(err).Msg("failed to migrate database")
		return err
	}
	return nil
}

// setupStorageDir handles the creation and validation of the storage directory.
func setupStorageDir(cfg *config.Config) error {
	if strings.HasPrefix(cfg.Storage.ConnString, "file:///./") {
		raw := strings.TrimPrefix(cfg.Storage.ConnString, "file:///./")
		clean := filepath.Clean(raw)
		absBase, err := filepath.Abs(clean)
		if err != nil {
			log.Error().Err(err).Msg("failed to get absolute path for storage connection string")
			return fmt.Errorf("failed to get absolute path for storage connection string: %w", err)
		}
		absBase = strings.ReplaceAll(absBase, "\\", "/")
		storageDir := filepath.Join(absBase, cfg.Storage.PrefixPath)
		storageDir = strings.ReplaceAll(storageDir, "\\", "/")
		if !strings.HasPrefix(storageDir, absBase+"/") && storageDir != absBase {
			log.Error().Str("path", storageDir).Msg("invalid storage path: you tried to use a prefix that is not a subdirectory of the base path")
			return fmt.Errorf("invalid storage path: you tried to use a prefix that is not a subdirectory of the base path")
		}
		if err := os.MkdirAll(storageDir, 0o750); err != nil {
			log.Error().Err(err).Msg("failed to create data directory")
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}
	return nil
}

// setupDatabaseURL returns the database URL and ensures any required directories exist.
func setupDatabaseURL(cfg *config.Config) (string, error) {
	databaseURL := ""
	switch strings.ToLower(cfg.Database.Driver) {
	case config.DriverSqlite3:
		databaseURL = cfg.Database.SqlitePath
		dbFilePath := strings.Split(cfg.Database.SqlitePath, "?")[0]
		dbDir := filepath.Dir(dbFilePath)
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			log.Error().Err(err).Str("path", dbDir).Msg("failed to create SQLite database directory")
			return "", fmt.Errorf("failed to create SQLite database directory: %w", err)
		}
	case config.DriverPostgres:
		databaseURL = fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, cfg.Database.SslMode)
		if cfg.Database.Username != "" {
			databaseURL += fmt.Sprintf(" user=%s", cfg.Database.Username)
		}
		if cfg.Database.Password != "" {
			databaseURL += fmt.Sprintf(" password=%s", cfg.Database.Password)
		}
		if cfg.Database.SslRootCert != "" {
			if _, err := os.Stat(cfg.Database.SslRootCert); err != nil {
				log.Error().Err(err).Str("path", cfg.Database.SslRootCert).Msg("SSL root certificate file is not accessible")
				return "", fmt.Errorf("SSL root certificate file is not accessible: %w", err)
			}
			databaseURL += fmt.Sprintf(" sslrootcert=%s", cfg.Database.SslRootCert)
		}
		if cfg.Database.SslCert != "" {
			if _, err := os.Stat(cfg.Database.SslCert); err != nil {
				log.Error().Err(err).Str("path", cfg.Database.SslCert).Msg("SSL certificate file is not accessible")
				return "", fmt.Errorf("SSL certificate file is not accessible: %w", err)
			}
			databaseURL += fmt.Sprintf(" sslcert=%s", cfg.Database.SslCert)
		}
		if cfg.Database.SslKey != "" {
			if _, err := os.Stat(cfg.Database.SslKey); err != nil {
				log.Error().Err(err).Str("path", cfg.Database.SslKey).Msg("SSL key file is not accessible")
				return "", fmt.Errorf("SSL key file is not accessible: %w", err)
			}
			databaseURL += fmt.Sprintf(" sslkey=%s", cfg.Database.SslKey)
		}
	default:
		log.Error().Str("driver", cfg.Database.Driver).Msg("unsupported database driver")
		return "", fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
	return databaseURL, nil
}

// loadCurrencies loads currency data from config if provided.
func loadCurrencies(cfg *config.Config) ([]currencies.CollectorFunc, error) {
	collectFuncs := []currencies.CollectorFunc{
		currencies.CollectDefaults(),
	}
	if cfg.Options.CurrencyConfig != "" {
		log.Info().Str("path", cfg.Options.CurrencyConfig).Msg("loading currency config file")
		content, err := os.ReadFile(cfg.Options.CurrencyConfig)
		if err != nil {
			log.Error().Err(err).Str("path", cfg.Options.CurrencyConfig).Msg("failed to read currency config file")
			return nil, err
		}
		collectFuncs = append(collectFuncs, currencies.CollectJSON(bytes.NewReader(content)))
	}
	return collectFuncs, nil
}
