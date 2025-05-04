// Package services provides the core business logic for the application.
package services

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

type AllServices struct {
	User              *UserService
	OAuth             *OAuthService
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

	user := &UserService{repos}

	return &AllServices{
		User:  user,
		OAuth: &OAuthService{repos: repos, user: user},
		Group: &GroupService{repos},
		Items: &ItemService{
			repo:                 repos,
			autoIncrementAssetID: options.autoIncrementAssetID,
		},
		BackgroundService: &BackgroundService{repos, Latest{}},
		Currencies:        currencies.NewCurrencyService(options.currencies),
	}
}
