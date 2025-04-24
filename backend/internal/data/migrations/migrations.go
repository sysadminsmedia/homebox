package migrations

import (
	"embed"
	"github.com/rs/zerolog/log"
)

//go:embed all:postgres
var postgresFiles embed.FS

//go:embed all:sqlite3
var sqliteFiles embed.FS

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
