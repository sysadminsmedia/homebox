package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v4/host"
	"net/http"
	"os"
	"path/filepath"
	"time"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/graceful"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	_ "github.com/shirou/gopsutil/v4/host"
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/web/mid"

	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"
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

// @title                      Homebox API
// @version                    1.0
// @description                Track, Manage, and Organize your Things.
// @contact.name               Don't
// @BasePath                   /api
// @securityDefinitions.apikey Bearer
// @in                         header
// @name                       Authorization
// @description                "Type 'Bearer TOKEN' to correctly set the API Key"
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

	if cfg.Options.AllowAnalytics {
		type analyticsData struct {
			Domain string                 `json:"domain"`
			Name   string                 `json:"name"`
			Url    string                 `json:"url"`
			Props  map[string]interface{} `json:"props"`
		}
		hostData, _ := host.Info()
		analytics := analyticsData{
			Domain: "homebox.software",
			Url:    "https://homebox.software",
			Name:   "stats",
			Props: map[string]interface{}{
				"os":               hostData.OS,
				"platform":         hostData.Platform,
				"platform_family":  hostData.PlatformFamily,
				"platform_version": hostData.PlatformVersion,
				"kernel_arch":      hostData.KernelArch,
			},
		}
		jsonBody, err := json.Marshal(analytics)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal analytics data")
		}
		bodyReader := bytes.NewReader(jsonBody)
		req, err := http.NewRequest("POST", "https://a.sysadmins.zone/api/event", bodyReader)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 OPR/71.0.3770.284")
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		log.Info().Msgf("Analytics Response: %v", res)
		log.Info().Msgf("Analytics Response: %v", json.NewEncoder(os.Stdout).Encode(analytics))
	}

	// =========================================================================
	// Initialize Database & Repos

	err := os.MkdirAll(cfg.Storage.Data, 0o755)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create data directory")
	}

	c, err := ent.Open("sqlite3", cfg.Storage.SqliteURL)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("driver", "sqlite").
			Str("url", cfg.Storage.SqliteURL).
			Msg("failed opening connection to sqlite")
	}
	defer func(c *ent.Client) {
		err := c.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to close database connection")
		}
	}(c)

	temp := filepath.Join(os.TempDir(), "migrations")

	err = migrations.Write(temp)
	if err != nil {
		return err
	}

	dir, err := atlas.NewLocalDir(temp)
	if err != nil {
		return err
	}

	options := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
	}

	err = c.Schema.Create(context.Background(), options...)
	if err != nil {
		log.Error().
			Err(err).
			Str("driver", "sqlite").
			Str("url", cfg.Storage.SqliteURL).
			Msg("failed creating schema resources")
		return err
	}

	err = os.RemoveAll(temp)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove temporary directory for database migrations")
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
	app.repos = repo.New(c, app.bus, cfg.Storage.Data)
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
	}

	return runner.Start(context.Background())
}
