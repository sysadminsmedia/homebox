package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/pressly/goose/v3"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/graceful"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/web/mid"
	"go.balki.me/anyhttp"

	_ "github.com/lib/pq"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/postgres"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/sqlite3"
	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"

	_ "gocloud.dev/pubsub/awssnssqs"
	_ "gocloud.dev/pubsub/azuresb"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
	_ "gocloud.dev/pubsub/mempubsub"
	_ "gocloud.dev/pubsub/natspubsub"
	_ "gocloud.dev/pubsub/rabbitpubsub"
)

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

func validatePostgresSSLMode(sslMode string) bool {
	validModes := map[string]bool{
		"":            true,
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	return validModes[strings.ToLower(strings.TrimSpace(sslMode))]
}

//	@title						Homebox API
//	@version					1.0
//	@description				Track, Manage, and Organize your Things.
//	@contact.name				Homebox Team
//	@contact.url				https://discord.homebox.software
//	@host						demo.homebox.software
//	@schemes					https http
//	@BasePath					/api
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				"Type 'Bearer TOKEN' to correctly set the API Key"
//	@externalDocs.url			https://homebox.software/en/api

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	cfg, err := config.New(build(), "Homebox inventory management system")
	if err != nil {
		panic(err)
	}

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	app := new(cfg)
	app.setupLogger()

	// =========================================================================
	// Initialize Database & Repos

	setupStorageDir(cfg)

	if strings.ToLower(cfg.Database.Driver) == "postgres" {
		if !validatePostgresSSLMode(cfg.Database.SslMode) {
			log.Fatal().Str("sslmode", cfg.Database.SslMode).Msg("invalid sslmode")
		}
	}

	databaseURL := setupDatabaseURL(cfg)

	c, err := ent.Open(strings.ToLower(cfg.Database.Driver), databaseURL)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("driver", strings.ToLower(cfg.Database.Driver)).
			Str("host", cfg.Database.Host).
			Str("port", cfg.Database.Port).
			Str("database", cfg.Database.Database).
			Msg("failed opening connection to {driver} database at {host}:{port}/{database}")
	}

	goose.SetBaseFS(migrations.Migrations(strings.ToLower(cfg.Database.Driver)))
	err = goose.SetDialect(strings.ToLower(cfg.Database.Driver))
	if err != nil {
		log.Fatal().Str("driver", cfg.Database.Driver).Msg("unsupported database driver")
		return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	err = goose.Up(c.Sql(), strings.ToLower(cfg.Database.Driver))
	if err != nil {
		log.Error().Err(err).Msg("failed to migrate database")
		return err
	}

	collectFuncs, err := loadCurrencies(cfg)
	if err != nil {
		return err
	}

	currencies, err := currencies.CollectionCurrencies(collectFuncs...)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to collect currencies")
		return err
	}

	app.bus = eventbus.New()
	app.db = c
	app.repos = repo.New(c, app.bus, cfg.Storage, cfg.Database.PubSubConnString, cfg.Thumbnail)
	app.services = services.New(
		app.repos,
		services.WithAutoIncrementAssetID(cfg.Options.AutoIncrementAssetID),
		services.WithCurrencies(currencies),
	)

	// =========================================================================
	// Start Server

	logger := log.With().Caller().Logger()

	router := chi.NewMux()
	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		mid.Logger(logger),
		middleware.Recoverer,
		middleware.StripSlashes,
	)

	chain := errchain.New(mid.Errors(logger))

	app.mountRoutes(router, chain, app.repos)

	runner := graceful.NewRunner()

	runner.AddFunc("server", func(ctx context.Context) error {
		httpserver := http.Server{
			Addr:         fmt.Sprintf("%s:%s", cfg.Web.Host, cfg.Web.Port),
			Handler:      router,
			ReadTimeout:  cfg.Web.ReadTimeout,
			WriteTimeout: cfg.Web.WriteTimeout,
			IdleTimeout:  cfg.Web.IdleTimeout,
		}

		go func() {
			<-ctx.Done()
			_ = httpserver.Shutdown(context.Background())
		}()

		listener, addrType, addrCfg, err := anyhttp.GetListener(cfg.Web.Host)
		if err == nil {
			switch addrType {
			case anyhttp.SystemdFD:
				sysdCfg := addrCfg.(*anyhttp.SysdConfig)
				if sysdCfg.IdleTimeout != nil {
					log.Error().Msg("idle timeout not yet supported. Please remove and try again")
					return errors.New("idle timeout not yet supported. Please remove and try again")
				}
				fallthrough
			case anyhttp.UnixSocket:
				log.Info().Msgf("Server is running on %s", cfg.Web.Host)
				return httpserver.Serve(listener)
			}
		} else {
			log.Debug().Msgf("anyhttp error: %v", err)
		}
		log.Info().Msgf("Server is running on %s:%s", cfg.Web.Host, cfg.Web.Port)
		return httpserver.ListenAndServe()
	})

	// Start Reoccurring Tasks
	registerRecurringTasks(app, cfg, runner)

	return runner.Start(context.Background())
}
