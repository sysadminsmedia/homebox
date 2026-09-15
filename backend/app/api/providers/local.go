package providers

import (
	"net/http"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type LocalProvider struct {
	service *services.UserService
}

func NewLocalProvider(service *services.UserService) *LocalProvider {
	return &LocalProvider{
		service: service,
	}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	ctx, span := otel.Tracer("provider").Start(r.Context(), "provider.LocalProvider.Authenticate",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.content_type", r.Header.Get("Content-Type")),
		))
	defer span.End()

	loginForm, err := getLoginForm(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("login.outcome", "form_decode_failed"))
		return services.UserAuthTokenDetail{}, err
	}
	span.SetAttributes(
		attribute.Int("login.username.length", len(loginForm.Username)),
		attribute.Int("login.password.length", len(loginForm.Password)),
		attribute.Bool("login.stay_logged_in", loginForm.StayLoggedIn),
	)

	out, err := p.service.Login(ctx, loginForm.Username, loginForm.Password, loginForm.StayLoggedIn)
	if err != nil {
		span.SetAttributes(attribute.String("login.outcome", "service_login_failed"))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return out, err
	}
	span.SetAttributes(attribute.String("login.outcome", "success"))
	return out, nil
}
