package main

import (
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/otel"
	"github.com/sysadminsmedia/homebox/backend/pkgs/mailer"
)

type app struct {
	conf                *config.Config
	mailer              mailer.Mailer
	db                  *ent.Client
	repos               *repo.AllRepos
	services            *services.AllServices
	bus                 *eventbus.EventBus
	authLimiter         *authRateLimiter
	notifierTestLimiter *simpleRateLimiter
	foundLimiter        *simpleRateLimiter
	foundSendLimiter    *simpleRateLimiter
	otel                *otel.Provider
}

func new(conf *config.Config) *app {
	s := &app{
		conf: conf,
	}

	s.mailer = mailer.Mailer{
		Host:     s.conf.Mailer.Host,
		Port:     s.conf.Mailer.Port,
		Username: s.conf.Mailer.Username,
		Password: s.conf.Mailer.Password,
		From:     s.conf.Mailer.From,
	}

	s.authLimiter = newAuthRateLimiter(s.conf.Auth.RateLimit)
	s.notifierTestLimiter = newSimpleRateLimiter(10, time.Minute, s.conf.Options.TrustProxy) // 10 requests per minute
	// Public found-item endpoints. foundLimiter is a per-IP request cap on
	// both routes (30/min/IP) that throttles asset-ID enumeration; unlike
	// mwAuthRateLimit it is not a failed-auth backoff and does not reset on a
	// successful response. foundSendLimiter is a per-item email cap
	// (3 per 10 min) that bounds owner-directed mailbombing even from
	// distributed source IPs.
	s.foundLimiter = newSimpleRateLimiter(30, time.Minute, s.conf.Options.TrustProxy)
	s.foundSendLimiter = newSimpleRateLimiter(3, 10*time.Minute, s.conf.Options.TrustProxy)

	return s
}
