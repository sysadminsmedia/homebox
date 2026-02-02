// Package v1 provides the API handlers for version 1 of the API.
package v1

import (
	"net/http"
	"time"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/app/api/providers"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/centrifuge"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

type Results[T any] struct {
	Items []T `json:"items"`
}

func WrapResults[T any](items []T) Results[T] {
	return Results[T]{Items: items}
}

type Wrapped struct {
	Item interface{} `json:"item"`
}

func Wrap(v any) Wrapped {
	return Wrapped{Item: v}
}

func WithMaxUploadSize(maxUploadSize int64) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.maxUploadSize = maxUploadSize
	}
}

func WithDemoStatus(demoStatus bool) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.isDemo = demoStatus
	}
}

func WithRegistration(allowRegistration bool) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.allowRegistration = allowRegistration
	}
}

func WithSecureCookies(secure bool) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.cookieSecure = secure
	}
}

func WithURL(url string) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.url = url
	}
}

func WithCentrifugeBroker(broker *centrifuge.Broker) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.broker = broker
	}
}

type V1Controller struct {
	cookieSecure      bool
	repo              *repo.AllRepos
	svc               *services.AllServices
	maxUploadSize     int64
	isDemo            bool
	allowRegistration bool
	bus               *eventbus.EventBus
	url               string
	config            *config.Config
	oidcProvider      *providers.OIDCProvider
	broker            *centrifuge.Broker
}

type (
	ReadyFunc func() bool

	Build struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		BuildTime string `json:"buildTime"`
	}

	APISummary struct {
		Healthy           bool            `json:"health"`
		Versions          []string        `json:"versions"`
		Title             string          `json:"title"`
		Message           string          `json:"message"`
		Build             Build           `json:"build"`
		Latest            services.Latest `json:"latest"`
		Demo              bool            `json:"demo"`
		AllowRegistration bool            `json:"allowRegistration"`
		LabelPrinting     bool            `json:"labelPrinting"`
		OIDC              OIDCStatus      `json:"oidc"`
	}

	OIDCStatus struct {
		Enabled      bool   `json:"enabled"`
		ButtonText   string `json:"buttonText,omitempty"`
		AutoRedirect bool   `json:"autoRedirect,omitempty"`
		AllowLocal   bool   `json:"allowLocal"`
	}
)

func NewControllerV1(svc *services.AllServices, repos *repo.AllRepos, bus *eventbus.EventBus, config *config.Config, options ...func(*V1Controller)) *V1Controller {
	ctrl := &V1Controller{
		repo:              repos,
		svc:               svc,
		allowRegistration: true,
		bus:               bus,
		config:            config,
	}

	for _, opt := range options {
		opt(ctrl)
	}

	ctrl.initOIDCProvider()

	return ctrl
}

func (ctrl *V1Controller) initOIDCProvider() {
	if ctrl.config.OIDC.Enabled {
		oidcProvider, err := providers.NewOIDCProvider(ctrl.svc.User, &ctrl.config.OIDC, &ctrl.config.Options, ctrl.cookieSecure)
		if err != nil {
			log.Err(err).Msg("failed to initialize OIDC provider at startup")
		} else {
			ctrl.oidcProvider = oidcProvider
			log.Info().Msg("OIDC provider initialized successfully at startup")
		}
	}
}

// HandleBase godoc
//
//	@Summary	Application Info
//	@Tags		Base
//	@Produce	json
//	@Success	200	{object}	APISummary
//	@Router		/v1/status [GET]
func (ctrl *V1Controller) HandleBase(ready ReadyFunc, build Build) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return server.JSON(w, http.StatusOK, APISummary{
			Healthy:           ready(),
			Title:             "Homebox",
			Message:           "Track, Manage, and Organize your Things",
			Build:             build,
			Latest:            ctrl.svc.BackgroundService.GetLatestVersion(),
			Demo:              ctrl.isDemo,
			AllowRegistration: ctrl.allowRegistration,
			LabelPrinting:     ctrl.config.LabelMaker.PrintCommand != nil,
			OIDC: OIDCStatus{
				Enabled:      ctrl.config.OIDC.Enabled,
				ButtonText:   ctrl.config.OIDC.ButtonText,
				AutoRedirect: ctrl.config.OIDC.AutoRedirect,
				AllowLocal:   ctrl.config.Options.AllowLocalLogin,
			},
		})
	}
}

// HandleCurrency godoc
//
//	@Summary	Currency
//	@Tags		Base
//	@Produce	json
//	@Success	200	{object}	currencies.Currency
//	@Router		/v1/currency [GET]
func (ctrl *V1Controller) HandleCurrency() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Set Cache for 10 Minutes
		w.Header().Set("Cache-Control", "max-age=600")

		return server.JSON(w, http.StatusOK, ctrl.svc.Currencies.Slice())
	}
}

// HandleWSEvents godoc
//
//	@Summary	WebSocket Events
//	@Tags		Base
//	@Router		/v1/ws/events [GET]
func (ctrl *V1Controller) HandleWSEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctrl.broker == nil {
			_ = server.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "websocket service not available",
			})
			return
		}

		ctrl.broker.HTTPHandler().ServeHTTP(w, r)
	}
}

// HandleWSToken godoc
//
//	@Summary	WebSocket Token
//	@Tags		Base
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/v1/ws/token [GET]
func (ctrl *V1Controller) HandleWSToken() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if ctrl.broker == nil {
			return server.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "websocket service not available",
			})
		}

		// Get user and tenant from context (set by middleware)
		ctx := services.NewContext(r.Context())

		// Generate token with 10 minute expiration
		token, err := ctrl.broker.GenerateToken(ctx.User.ID, ctx.GID, 10*time.Minute)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate centrifuge token")
			return server.JSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to generate token",
			})
		}

		return server.JSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}
