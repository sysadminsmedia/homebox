package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/analytics"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"

	"github.com/pressly/goose/v3"

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

	"gocloud.dev/pubsub"
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

//nolint:gocyclo
func run(cfg *config.Config) error {
	app := new(cfg)
	app.setupLogger()

	// =========================================================================
	// Initialize Database & Repos

	if strings.HasPrefix(cfg.Storage.ConnString, "file:///./") {
		raw := strings.TrimPrefix(cfg.Storage.ConnString, "file:///./")
		clean := filepath.Clean(raw)
		absBase, err := filepath.Abs(clean)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get absolute path for storage connection string")
		}
		// Construct and validate the full storage path
		storageDir := filepath.Join(absBase, cfg.Storage.PrefixPath)
		// Set windows paths to use forward slashes required by go-cloud
		storageDir = strings.ReplaceAll(storageDir, "\\", "/")
		if !strings.HasPrefix(storageDir, absBase+"/") && storageDir != absBase {
			log.Fatal().
				Str("path", storageDir).
				Msg("invalid storage path: you tried to use a prefix that is not a subdirectory of the base path")
		}
		// Create with more restrictive permissions
		if err := os.MkdirAll(storageDir, 0o750); err != nil {
			log.Fatal().
				Err(err).
				Msg("failed to create data directory")
		}
	}

	if strings.ToLower(cfg.Database.Driver) == "postgres" {
		if !validatePostgresSSLMode(cfg.Database.SslMode) {
			log.Fatal().Str("sslmode", cfg.Database.SslMode).Msg("invalid sslmode")
		}
	}

	// Set up the database URL based on the driver because for some reason a common URL format is not used
	databaseURL := ""
	switch strings.ToLower(cfg.Database.Driver) {
	case "sqlite3":
		databaseURL = cfg.Database.SqlitePath

		// Create directory for SQLite database if it doesn't exist
		dbFilePath := strings.Split(cfg.Database.SqlitePath, "?")[0] // Remove query parameters
		dbDir := filepath.Dir(dbFilePath)
		if err := os.MkdirAll(dbDir, 0o755); err != nil {
			log.Fatal().Err(err).Str("path", dbDir).Msg("failed to create SQLite database directory")
		}
	case "postgres":
		databaseURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Password, cfg.Database.Database, cfg.Database.SslMode)
	default:
		log.Fatal().Str("driver", cfg.Database.Driver).Msg("unsupported database driver")
	}

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

	collectFuncs := []currencies.CollectorFunc{
		currencies.CollectDefaults(),
	}

	if cfg.Options.CurrencyConfig != "" {
		log.Info().
			Str("path", cfg.Options.CurrencyConfig).
			Msg("loading currency config file")

		content, err := os.ReadFile(cfg.Options.CurrencyConfig)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", cfg.Options.CurrencyConfig).
				Msg("failed to read currency config file")
			return err
		}

		collectFuncs = append(collectFuncs, currencies.CollectJSON(bytes.NewReader(content)))
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

	// =========================================================================
	// Start Reoccurring Tasks

	runner.AddFunc("eventbus", app.bus.Run)

	runner.AddFunc("seed_database", func(ctx context.Context) error {
		// TODO: Remove through external API that does setup
		if cfg.Demo {
			log.Info().Msg("Running in demo mode, creating demo data")
			err := app.SetupDemo()
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}
		return nil
	})

	runner.AddPlugin(NewTask("purge-tokens", time.Duration(24)*time.Hour, func(ctx context.Context) {
		_, err := app.repos.AuthTokens.PurgeExpiredTokens(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to purge expired tokens")
		}
	}))

	runner.AddPlugin(NewTask("purge-invitations", time.Duration(24)*time.Hour, func(ctx context.Context) {
		_, err := app.repos.Groups.InvitationPurge(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to purge expired invitations")
		}
	}))

	runner.AddPlugin(NewTask("send-notifications", time.Duration(1)*time.Hour, func(ctx context.Context) {
		now := time.Now()

		if now.Hour() == 8 {
			fmt.Println("run notifiers")
			err := app.services.BackgroundService.SendNotifiersToday(context.Background())
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed to send notifiers")
			}
		}
	}))

	go runner.AddFunc("create-thumbnails-subscription", func(ctx context.Context) error {
		pubsubString, err := utils.GenerateSubPubConn(cfg.Database.PubSubConnString, "thumbnails")
		if err != nil {
			log.Error().Err(err).Msg("failed to generate pubsub connection string")
			return err
		}
		topic, err := pubsub.OpenTopic(ctx, pubsubString)
		if err != nil {
			return err
		}
		defer func(topic *pubsub.Topic, ctx context.Context) {
			err := topic.Shutdown(ctx)
			if err != nil {
				log.Err(err).Msg("fail to shutdown pubsub topic")
			}
		}(topic, ctx)

		subscription, err := pubsub.OpenSubscription(ctx, pubsubString)
		if err != nil {
			log.Err(err).Msg("failed to open pubsub topic")
			return err
		}
		defer func(topic *pubsub.Subscription, ctx context.Context) {
			err := topic.Shutdown(ctx)
			if err != nil {
				log.Err(err).Msg("fail to shutdown pubsub topic")
			}
		}(subscription, ctx)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				msg, err := subscription.Receive(ctx)
				log.Debug().Msg("received thumbnail generation request from pubsub topic")
				if err != nil {
					log.Err(err).Msg("failed to receive message from pubsub topic")
				}
				groupId, err := uuid.Parse(msg.Metadata["group_id"])
				if err != nil {
					log.Error().
						Err(err).
						Str("group_id", msg.Metadata["group_id"]).
						Msg("failed to parse group ID from message metadata")
				}
				attachmentId, err := uuid.Parse(msg.Metadata["attachment_id"])
				if err != nil {
					log.Error().
						Err(err).
						Str("attachment_id", msg.Metadata["attachment_id"]).
						Msg("failed to parse attachment ID from message metadata")
				}
				err = app.repos.Attachments.CreateThumbnail(ctx, groupId, attachmentId, msg.Metadata["title"], msg.Metadata["path"])
				if err != nil {
					log.Err(err).Msg("failed to create thumbnail")
				}
				msg.Ack()
			}
		}
	})

	if cfg.Options.GithubReleaseCheck {
		runner.AddPlugin(NewTask("get-latest-github-release", time.Hour, func(ctx context.Context) {
			log.Debug().Msg("running get latest github release")
			err := app.services.BackgroundService.GetLatestGithubRelease(context.Background())
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed to get latest github release")
			}
		}))
	}

	if cfg.Options.AllowAnalytics {
		runner.AddPlugin(NewTask("send-analytics", time.Duration(24)*time.Hour, func(ctx context.Context) {
			log.Debug().Msg("running send analytics")
			err := analytics.Send(version, build())
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed to send scheduled analytics")
			}
		}))
	}

	if cfg.Debug.Enabled {
		runner.AddFunc("debug", func(ctx context.Context) error {
			debugserver := http.Server{
				Addr:         fmt.Sprintf("%s:%s", cfg.Web.Host, cfg.Debug.Port),
				Handler:      app.debugRouter(),
				ReadTimeout:  cfg.Web.ReadTimeout,
				WriteTimeout: cfg.Web.WriteTimeout,
				IdleTimeout:  cfg.Web.IdleTimeout,
			}

			go func() {
				<-ctx.Done()
				_ = debugserver.Shutdown(context.Background())
			}()

			log.Info().Msgf("Debug server is running on %s:%s", cfg.Web.Host, cfg.Debug.Port)
			return debugserver.ListenAndServe()
		})
		// Print the configuration to the console
		cfg.Print()
	}

	return runner.Start(context.Background())
}
