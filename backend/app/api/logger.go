package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// setupLogger initializes the zerolog config
// for the shared logger.
func (a *app) setupLogger() {
	// Logger Init
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if a.conf.Log.Format != config.LogFormatJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	}

	level, err := zerolog.ParseLevel(a.conf.Log.Level)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}
}
