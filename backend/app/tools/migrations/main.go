package main

import (
	"context"
	"fmt"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"log"
	"os"
	"strings"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/migrate"

	atlas "ariga.io/atlas/sql/migrate"
	_ "ariga.io/atlas/sql/sqlite"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.New(build(), "Homebox inventory management system")
	if err != nil {
		panic(err)
	}
	sqlDialect := ""
	switch strings.ToLower(cfg.Database.Driver) {
	case "sqlite3":
		sqlDialect = dialect.SQLite
	case "mysql":
		sqlDialect = dialect.MySQL
	case "postgres":
		sqlDialect = dialect.Postgres
	default:
		log.Fatalf("unsupported database driver: %s", cfg.Database.Driver)
	}
	ctx := context.Background()
	// Create a local migration directory able to understand Atlas migration file format for replay.
	dir, err := atlas.NewLocalDir(fmt.Sprintf("internal/data/migrations/%s", sqlDialect))
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}
	// Migrate diff options.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(sqlDialect),              // Ent dialect to use
		schema.WithFormatter(atlas.DefaultFormatter),
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
	}
	if len(os.Args) != 2 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod ent/migrate/main.go <name>'")
	}

	databaseURL := ""
	switch {
	case cfg.Database.Driver == "sqlite3":
		databaseURL = fmt.Sprintf("sqlite://%s", cfg.Database.SqlitePath)
	case cfg.Database.Driver == "postgres":
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, cfg.Database.SslMode)
	default:
		log.Fatalf("unsupported database driver: %s", cfg.Database.Driver)
	}

	// Generate migrations using Atlas support for MySQL (note the Ent dialect option passed above).
	err = migrate.NamedDiff(ctx, databaseURL, os.Args[1], opts...)
	if err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}

	fmt.Println("Migration file generated successfully.")
}

var (
	version   = "nightly"
	commit    = "HEAD"
	buildTime = "now"
)

func build() string {
	short := commit
	if len(short) > 7 {
		short = short[:7]
	}

	return fmt.Sprintf("%s, commit %s, built at %s", version, short, buildTime)
}
