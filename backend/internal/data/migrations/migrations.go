// Package migrations
package migrations

import (
	"embed"
	"github.com/rs/zerolog/log"
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
func Migrations(dialect string) embed.FS {
	switch dialect {
	case "postgres":
		return postgresFiles
	case "sqlite3":
		return sqliteFiles
	default:
		log.Fatal().Str("dialect", dialect).Msg("unknown sql dialect")
	}
	// This should never get hit, but just in case
	return sqliteFiles
}
