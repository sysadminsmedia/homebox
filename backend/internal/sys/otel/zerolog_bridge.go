package otel

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	otellog "go.opentelemetry.io/otel/log"
)

// ZerologOTelHook forwards zerolog events to OpenTelemetry logs.
type ZerologOTelHook struct {
	logger otellog.Logger
}

// NewZerologOTelHook creates a new hook using the provided LoggerProvider and logger name.
func NewZerologOTelHook(lp otellog.LoggerProvider, name string) *ZerologOTelHook {
	if lp == nil {
		return nil
	}
	return &ZerologOTelHook{logger: lp.Logger(name)}
}

// Run implements zerolog.Hook and emits an OTel log Record with severity and message.
func (h *ZerologOTelHook) Run(_ *zerolog.Event, level zerolog.Level, msg string) {
	if h == nil || h.logger == nil {
		return
	}

	// Create a record and map the severity.
	var rec otellog.Record
	rec.SetTimestamp(time.Now())
	rec.SetSeverity(mapZerologLevel(level))
	rec.SetSeverityText(level.String())
	rec.SetBody(otellog.StringValue(msg))

	// Emit with background context; if you have a request context, prefer passing it.
	h.logger.Emit(context.Background(), rec)
}

// mapZerologLevel converts zerolog levels to OTel log severities.
func mapZerologLevel(level zerolog.Level) otellog.Severity {
	switch level {
	case zerolog.TraceLevel:
		return otellog.SeverityTrace
	case zerolog.DebugLevel:
		return otellog.SeverityDebug
	case zerolog.InfoLevel:
		return otellog.SeverityInfo
	case zerolog.WarnLevel:
		return otellog.SeverityWarn
	case zerolog.ErrorLevel:
		return otellog.SeverityError
	case zerolog.FatalLevel:
		return otellog.SeverityFatal
	case zerolog.PanicLevel:
		return otellog.SeverityError
	default:
		return otellog.SeverityInfo
	}
}
