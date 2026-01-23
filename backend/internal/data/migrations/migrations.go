// Package migrations
package migrations

import (
	"embed"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

//go:embed all:postgres
var postgresFiles embed.FS

//go:embed all:sqlite3
var sqliteFiles embed.FS

// Migrations returns the embedded file system containing the SQL migration files
// for the specified SQL dialect. It uses the "embed" package to include the
// migration files in the binary at build time. The function takes a string
// parameter "dialect" which specifies the SQL dialect to use. It returns an
// embedded file system containing the migration files for the specified dialect.
func Migrations(dialect string) (embed.FS, error) {
	switch dialect {
	case config.DriverPostgres:
		return postgresFiles, nil
	case config.DriverSqlite3:
		return sqliteFiles, nil
	default:
		log.Error().Str("dialect", dialect).Msg("unknown sql dialect")
		return embed.FS{}, fmt.Errorf("unknown sql dialect: %s", dialect)
	}
	// This should never get hit, but just in case
}
