package main

import (
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
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

	return s
}
