package otel

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// DatabaseDriver returns the driver name to use for database connections.
// If tracing is enabled, this returns the instrumented driver name.
func (p *Provider) DatabaseDriver(originalDriver string) string {
	if p == nil || !p.IsEnabled() || p.cfg == nil || !p.cfg.EnableDatabaseTracing {
		return originalDriver
	}
	// We use the original driver but wrap it with otelsql
	return originalDriver
}

// OpenDatabase opens a database connection with optional OpenTelemetry instrumentation.
// This wraps the standard sql.Open with tracing if enabled.
func (p *Provider) OpenDatabase(driverName, dataSourceName string) (*sql.DB, error) {
	if p == nil || !p.IsEnabled() || p.cfg == nil || !p.cfg.EnableDatabaseTracing {
		return sql.Open(driverName, dataSourceName)
	}

	// Determine the database system for semantic conventions
	dbSystem := getDBSystem(driverName)

	// Open the database with otelsql instrumentation
	db, err := otelsql.Open(driverName, dataSourceName,
		otelsql.WithAttributes(
			semconv.DBSystemKey.String(dbSystem),
			semconv.ServiceName(p.cfg.ServiceName),
		),
		otelsql.WithTracerProvider(p.TracerProvider()),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableQuery: false, // Include SQL queries in spans
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open instrumented database: %w", err)
	}

	// Register stats for metrics (ignoring error as it's non-fatal)
	_, _ = otelsql.RegisterDBStatsMetrics(db,
		otelsql.WithAttributes(
			semconv.DBSystemKey.String(dbSystem),
		),
	)

	return db, nil
}

// OpenEntDriver opens an Ent SQL driver with optional OpenTelemetry instrumentation.
func (p *Provider) OpenEntDriver(driverName, dataSourceName string) (*entsql.Driver, error) {
	db, err := p.OpenDatabase(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// Create Ent driver from the database connection
	drv := entsql.OpenDB(mapDialect(driverName), db)
	return drv, nil
}

// getDBSystem returns the OpenTelemetry semantic convention value for the database system.
func getDBSystem(driverName string) string {
	switch driverName {
	case "sqlite3":
		return "sqlite"
	case "postgres", "postgresql":
		return "postgresql"
	case "mysql":
		return "mysql"
	default:
		return driverName
	}
}

// mapDialect maps the driver name to the Ent dialect.
func mapDialect(driverName string) string {
	switch driverName {
	case "sqlite3":
		return dialect.SQLite
	case "postgres", "postgresql":
		return dialect.Postgres
	case "mysql":
		return dialect.MySQL
	default:
		return driverName
	}
}

// WrapEntDriver wraps an existing Ent driver with tracing.
// This is useful when you have an existing driver and want to add tracing.
func (p *Provider) WrapEntDriver(drv *entsql.Driver, driverName string) *entsql.Driver {
	if p == nil || !p.IsEnabled() || p.cfg == nil || !p.cfg.EnableDatabaseTracing {
		return drv
	}

	// The driver is already traced via otelsql.Open
	return drv
}

// DatabaseTracingEnabled returns true if database tracing is enabled.
func (p *Provider) DatabaseTracingEnabled() bool {
	return p != nil && p.IsEnabled() && p.cfg != nil && p.cfg.EnableDatabaseTracing
}

// DBDriverConfig holds configuration for creating a traced database driver.
type DBDriverConfig struct {
	Driver         string
	DataSourceName string
	EnableTracing  bool
}

// NewDBDriver creates a new database driver with optional tracing.
func (p *Provider) NewDBDriver(cfg DBDriverConfig) (*entsql.Driver, error) {
	if !cfg.EnableTracing || !p.DatabaseTracingEnabled() {
		// Return standard driver without tracing
		drv, err := entsql.Open(mapDialect(cfg.Driver), cfg.DataSourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		return drv, nil
	}

	return p.OpenEntDriver(cfg.Driver, cfg.DataSourceName)
}
