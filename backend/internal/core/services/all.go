// Package services provides the core business logic for the application.
package services

import (
	"github.com/hay-kot/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
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
		BackgroundService: &BackgroundService{repos},
		Currencies:        currencies.NewCurrencyService(options.currencies),
	}
}
