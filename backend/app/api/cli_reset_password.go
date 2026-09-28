package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// runResetPasswordCLI handles `homebox reset-password --email=...`. It mints a
// one-time reset link and prints it to stdout. This is the escape hatch for
// installations without SMTP, or for debugging the password matching path
// itself — the cases where the email-based flow can't help.
//
// Returns true when it consumed the command (and the caller should exit), so
// `homebox` with no subcommand still falls through to the server.
func runResetPasswordCLI(args []string) (handled bool, exitCode int) {
	if len(args) < 2 || args[1] != "reset-password" {
		return false, 0
	}

	fs := flag.NewFlagSet("reset-password", flag.ContinueOnError)
	email := fs.String("email", "", "Email address of the account to reset")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: homebox reset-password --email=<address>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Generates a one-time password reset link for the given account and")
		fmt.Fprintln(os.Stderr, "prints it to stdout. The link expires in one hour and can be used once.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "All HBOX_* environment variables (database, hostname) are honored.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args[2:]); err != nil {
		return true, 2
	}
	trimmedEmail := strings.TrimSpace(*email)
	if trimmedEmail == "" {
		fs.Usage()
		return true, 2
	}

	cfg, err := config.New(build(), "Homebox inventory management system")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return true, 1
	}

	link, err := generateResetLinkOffline(cfg, trimmedEmail)
	if err != nil {
		if ent.IsNotFound(err) {
			fmt.Fprintf(os.Stderr, "no account found for %s\n", trimmedEmail)
			return true, 1
		}
		fmt.Fprintf(os.Stderr, "failed to generate reset link: %v\n", err)
		return true, 1
	}

	fmt.Println(link)
	return true, 0
}

// generateResetLinkOffline opens the database, runs migrations if needed, and
// mints a token without starting the HTTP server. The returned link uses
// HBOX_OPTIONS_HOSTNAME if set; otherwise it's emitted as a path the operator
// can append to whatever URL their instance is reachable at.
func generateResetLinkOffline(cfg *config.Config, email string) (string, error) {
	databaseURL, err := setupDatabaseURL(cfg)
	if err != nil {
		return "", fmt.Errorf("setup database url: %w", err)
	}

	driver := strings.ToLower(cfg.Database.Driver)
	var driverName, dialectName string
	switch driver {
	case config.DriverPostgres:
		driverName = "pgx"
		dialectName = dialect.Postgres
	case config.DriverSqlite3, "sqlite":
		driverName = "sqlite3"
		dialectName = dialect.SQLite
	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		return "", fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	drv := entsql.OpenDB(dialectName, db)
	c := ent.NewClient(ent.Driver(drv))
	defer func() { _ = c.Close() }()

	migrationsFs, err := migrations.Migrations(driver)
	if err != nil {
		return "", fmt.Errorf("load migrations: %w", err)
	}
	goose.SetBaseFS(migrationsFs)
	if err := goose.SetDialect(driver); err != nil {
		return "", fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(c.Sql(), driver); err != nil {
		return "", fmt.Errorf("apply migrations: %w", err)
	}

	bus := eventbus.New()
	repos := repo.New(c, bus, cfg.Storage, cfg.Database.PubSubConnString, cfg.Thumbnail, nil)
	svc := services.New(repos)

	baseURL := strings.TrimSuffix(cfg.Options.Hostname, "/")
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		// Hostname without a scheme: assume https. The CLI has no request to
		// inspect for r.TLS, and https is the safer default to print.
		baseURL = "https://" + baseURL
	}

	link, err := svc.User.GenerateResetLink(context.Background(), email, baseURL)
	if err != nil {
		if errors.Is(err, services.ErrorMailerNotConfigured) {
			// Should never happen here; GenerateResetLink doesn't touch the
			// mailer. Defensive in case future refactors reorder things.
			return "", err
		}
		return "", err
	}
	return link, nil
}
