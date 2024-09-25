// Package v1 provides the API handlers for version 1 of the API.
package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"

	"github.com/olahol/melody"
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

func WithHeaderAuthEnabled(headerAuthEnabled bool) func(*V1Controller) {
	return func(ctrl *V1Controller) {
		ctrl.headerAuthEnabled = headerAuthEnabled
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

type V1Controller struct {
	cookieSecure      bool
	repo              *repo.AllRepos
	svc               *services.AllServices
	maxUploadSize     int64
	isDemo            bool
	allowRegistration bool
	headerAuthEnabled bool
	bus               *eventbus.EventBus
	url               string
}

type (
	ReadyFunc func() bool

	Build struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		BuildTime string `json:"buildTime"`
	}

	APISummary struct {
		Healthy           bool     `json:"health"`
		Versions          []string `json:"versions"`
		Title             string   `json:"title"`
		Message           string   `json:"message"`
		Build             Build    `json:"build"`
		Demo              bool     `json:"demo"`
		AllowRegistration bool     `json:"allowRegistration"`
		HeaderAuthEnabled bool     `json:"headerAuthEnabled"`
	}
)

func NewControllerV1(svc *services.AllServices, repos *repo.AllRepos, bus *eventbus.EventBus, options ...func(*V1Controller)) *V1Controller {
	ctrl := &V1Controller{
		repo:              repos,
		svc:               svc,
		allowRegistration: true,
		bus:               bus,
	}

	for _, opt := range options {
		opt(ctrl)
	}

	return ctrl
}

// HandleBase godoc
//
//	@Summary Application Info
//	@Tags    Base
//	@Produce json
//	@Success 200 {object} APISummary
//	@Router  /v1/status [GET]
func (ctrl *V1Controller) HandleBase(ready ReadyFunc, build Build) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return server.JSON(w, http.StatusOK, APISummary{
			Healthy:           ready(),
			Title:             "Homebox",
			Message:           "Track, Manage, and Organize your Things",
			Build:             build,
			Demo:              ctrl.isDemo,
			AllowRegistration: ctrl.allowRegistration,
			HeaderAuthEnabled: ctrl.headerAuthEnabled,
		})
	}
}

// HandleCurrency godoc
//
// @Summary Currency
// @Tags    Base
// @Produce json
// @Success 200 {object} currencies.Currency
// @Router  /v1/currency [GET]
func (ctrl *V1Controller) HandleCurrency() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Set Cache for 10 Minutes
		w.Header().Set("Cache-Control", "max-age=600")

		return server.JSON(w, http.StatusOK, ctrl.svc.Currencies.Slice())
	}
}

func (ctrl *V1Controller) HandleCacheWS() errchain.HandlerFunc {
	type eventMsg struct {
		Event string `json:"event"`
	}

	m := melody.New()

	m.HandleConnect(func(s *melody.Session) {
		auth := services.NewContext(s.Request.Context())
		s.Set("gid", auth.GID)
	})

	factory := func(e string) func(data any) {
		return func(data any) {
			eventData, ok := data.(eventbus.GroupMutationEvent)
			if !ok {
				log.Log().Msgf("invalid event data: %v", data)
				return
			}

			msg := &eventMsg{Event: e}

			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				log.Log().Msgf("error marshling event data %v: %v", data, err)
				return
			}

			_ = m.BroadcastFilter(jsonBytes, func(s *melody.Session) bool {
				groupIDStr, ok := s.Get("gid")
				if !ok {
					return false
				}

				GID := groupIDStr.(uuid.UUID)
				return GID == eventData.GID
			})
		}
	}

	ctrl.bus.Subscribe(eventbus.EventLabelMutation, factory("label.mutation"))
	ctrl.bus.Subscribe(eventbus.EventLocationMutation, factory("location.mutation"))
	ctrl.bus.Subscribe(eventbus.EventItemMutation, factory("item.mutation"))

	// Persistent asynchronous ticker that keeps all websocket connections alive with periodic pings.
	go func() {
		const interval = 10 * time.Second

		ping := time.NewTicker(interval)
		defer ping.Stop()

		for range ping.C {
			msg := &eventMsg{Event: "ping"}

			pingBytes, err := json.Marshal(msg)
			if err != nil {
				log.Log().Msgf("error marshaling ping: %v", err)
			} else {
				_ = m.Broadcast(pingBytes)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) error {
		return m.HandleRequest(w, r)
	}
}
