// Package services provides the core business logic for the application.
package services

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

type AllServices struct {
	User              *UserService
	Group             *GroupService
	Items             *ItemService
	BackgroundService *BackgroundService
	Currencies        *currencies.CurrencyRegistry
}

type OptionsFunc func(*options)

type options struct {
	autoIncrementAssetID bool
	currencies           []currencies.Currency
	notifierConfig       *config.NotifierConf
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
		Items: &ItemService{
			repo:                 repos,
			autoIncrementAssetID: options.autoIncrementAssetID,
		},
		BackgroundService: &BackgroundService{
			repos:          repos,
			latest:         Latest{},
			notifierConfig: options.notifierConfig,
		},
		Currencies: currencies.NewCurrencyService(options.currencies),
	}
}
