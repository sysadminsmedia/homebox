// Package services provides the core business logic for the application.
package services

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

type AllServices struct {
	User              *UserService
	Group             *GroupService
	Entities          *EntityService
	BackgroundService *BackgroundService
	Exports           *ExportService
	Currencies        *currencies.CurrencyRegistry
}

type OptionsFunc func(*options)

type options struct {
	autoIncrementAssetID bool
	currencies           []currencies.Currency
	notifierConfig       *config.NotifierConf
	bus                  *eventbus.EventBus
	db                   *ent.Client
	storage              config.Storage
	pubSubConn           string
	dialect              string
}

func WithAutoIncrementAssetID(v bool) func(*options) {
	return func(o *options) {
		o.autoIncrementAssetID = v
	}
}

func WithCurrencies(v []currencies.Currency) func(*options) {
	return func(o *options) {
		o.currencies = v
	}
}

func WithNotifierConfig(v *config.NotifierConf) func(*options) {
	return func(o *options) {
		if v != nil {
			o.notifierConfig = v
		}
	}
}

// WithExportPlumbing wires the dependencies the ExportService needs to dump
// raw SQL through the ent client and to publish job messages.
func WithExportPlumbing(bus *eventbus.EventBus, db *ent.Client, storage config.Storage, pubSubConn, dialect string) func(*options) {
	return func(o *options) {
		o.bus = bus
		o.db = db
		o.storage = storage
		o.pubSubConn = pubSubConn
		o.dialect = dialect
	}
}

// defaultNotifierConf returns a NotifierConf with safe defaults matching the conf tags.
// This ensures SSRF protections are enabled when WithNotifierConfig is not provided.
func defaultNotifierConf() *config.NotifierConf {
	return &config.NotifierConf{
		BlockBogonNets:     true, // default:true per conf tag
		BlockCloudMetadata: true, // default:true per conf tag
	}
}

func New(repos *repo.AllRepos, opts ...OptionsFunc) *AllServices {
	if repos == nil {
		panic("repos cannot be nil")
	}

	defaultCurrencies, err := currencies.CollectionCurrencies(
		currencies.CollectDefaults(),
	)
	if err != nil {
		panic("failed to collect default currencies")
	}

	options := &options{
		autoIncrementAssetID: true,
		currencies:           defaultCurrencies,
		notifierConfig:       defaultNotifierConf(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &AllServices{
		User:  &UserService{repos},
		Group: &GroupService{repos},
		Entities: &EntityService{
			repo:                 repos,
			autoIncrementAssetID: options.autoIncrementAssetID,
		},
		BackgroundService: &BackgroundService{
			repos:          repos,
			latest:         Latest{},
			notifierConfig: options.notifierConfig,
		},
		Exports: &ExportService{
			db:         options.db,
			repos:      repos,
			bus:        options.bus,
			storage:    options.storage,
			pubSubConn: options.pubSubConn,
			dialect:    options.dialect,
		},
		Currencies: currencies.NewCurrencyService(options.currencies),
	}
}
