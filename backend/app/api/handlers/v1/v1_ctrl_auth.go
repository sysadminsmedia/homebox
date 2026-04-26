package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"go.opentelemetry.io/otel/attribute"
)

const (
	cookieNameToken    = "hb.auth.token"
	cookieNameRemember = "hb.auth.remember"
	cookieNameSession  = "hb.auth.session"
)

type (
	TokenResponse struct {
		Token           string    `json:"token"`
		ExpiresAt       time.Time `json:"expiresAt"`
		AttachmentToken string    `json:"attachmentToken"`
	}

	LoginForm struct {
		Username     string `json:"username"     example:"admin@admin.com"`
		Password     string `json:"password"     example:"admin"`
		StayLoggedIn bool   `json:"stayLoggedIn"`
	}
)

type CookieContents struct {
	Token     string
	ExpiresAt time.Time
	Remember  bool
}

func GetCookies(r *http.Request) (*CookieContents, error) {
	cookie, err := r.Cookie(cookieNameToken)
	if err != nil {
		return nil, errors.New("authorization cookie is required")
	}

	rememberCookie, err := r.Cookie(cookieNameRemember)
	if err != nil {
		return nil, errors.New("remember cookie is required")
	}

	return &CookieContents{
		Token:     cookie.Value,
		ExpiresAt: cookie.Expires,
		Remember:  rememberCookie.Value == "true",
	}, nil
}

// AuthProvider is an interface that can be implemented by any authentication provider.
// to extend authentication methods for the API.
type AuthProvider interface {
	// Name returns the name of the authentication provider. This should be a unique name.
	// that is URL friendly.
	//
	// Example: "local", "ldap"
	Name() string
	// Authenticate is called when a user attempts to login to the API. The implementation
	// should return an error if the user cannot be authenticated. If an error is returned
	// the API controller will return a vague error message to the user.
	//
	// Authenticate should do the following:
	//
	// 1. Ensure that the user exists within the database (either create, or get)
	// 2. On successful authentication, they must set the user cookies.
	Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error)
}

// HandleAuthLogin godoc
//
//	@Summary	User Login
//	@Tags		Authentication
//	@Accept		x-www-form-urlencoded
//	@Accept		application/json
//	@Param		payload		body	LoginForm	true	"Login Data"
//	@Param		provider	query	string		false	"auth provider"
//	@Produce	json
//	@Success	200	{object}	TokenResponse
//	@Router		/v1/users/login [POST]
func (ctrl *V1Controller) HandleAuthLogin(ps ...AuthProvider) errchain.HandlerFunc {
	if len(ps) == 0 {
		panic("no auth providers provided")
	}

	providers := lo.SliceToMap(ps, func(p AuthProvider) (string, AuthProvider) {
		log.Info().Str("name", p.Name()).Msg("registering auth provider")
		return p.Name(), p
	})

	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleAuthLogin")
		defer span.End()

		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "local"
		}
		span.SetAttributes(attribute.String("auth.provider", provider))

		if provider == "local" && !ctrl.config.Options.AllowLocalLogin {
			span.SetAttributes(attribute.String("auth.outcome", "local_disabled"))
			return validate.NewRequestError(fmt.Errorf("local login is not enabled"), http.StatusForbidden)
		}

		p, ok := providers[provider]
		if !ok {
			span.SetAttributes(attribute.String("auth.outcome", "unknown_provider"))
			return validate.NewRequestError(errors.New("invalid auth provider"), http.StatusBadRequest)
		}

		newToken, err := p.Authenticate(w, r.WithContext(spanCtx))
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("auth.outcome", "authenticate_failed"))
			log.Warn().Err(err).Msg("authentication failed")
			return validate.NewUnauthorizedError()
		}
		span.SetAttributes(
			attribute.String("auth.outcome", "success"),
			attribute.String("auth.session.expires_at", newToken.ExpiresAt.Format(time.RFC3339)),
		)

		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, true, newToken.AttachmentToken)
		return server.JSON(w, http.StatusOK, TokenResponse{
			Token:           "Bearer " + newToken.Raw,
			ExpiresAt:       newToken.ExpiresAt,
			AttachmentToken: newToken.AttachmentToken,
		})
	}
}

// HandleAuthLogout godoc
//
//	@Summary	User Logout
//	@Tags		Authentication
//	@Success	204
//	@Router		/v1/users/logout [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAuthLogout() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleAuthLogout")
		defer span.End()

		token := services.UseTokenCtx(spanCtx)
		if token == "" {
			span.SetAttributes(attribute.String("logout.outcome", "no_token"))
			return validate.NewRequestError(errors.New("no token within request context"), http.StatusUnauthorized)
		}

		err := ctrl.svc.User.Logout(spanCtx, token)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("logout.outcome", "delete_failed"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		span.SetAttributes(attribute.String("logout.outcome", "success"))
		ctrl.unsetCookies(w, noPort(r.Host))
		return server.JSON(w, http.StatusNoContent, nil)
	}
}

// HandleAuthRefresh godoc
//
//	@Summary		User Token Refresh
//	@Description	handleAuthRefresh returns a handler that will issue a new token from an existing token.
//	@Description	This does not validate that the user still exists within the database.
//	@Tags			Authentication
//	@Success		200
//	@Router			/v1/users/refresh [GET]
//	@Security		Bearer
func (ctrl *V1Controller) HandleAuthRefresh() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleAuthRefresh")
		defer span.End()

		requestToken := services.UseTokenCtx(spanCtx)
		if requestToken == "" {
			span.SetAttributes(attribute.String("refresh.outcome", "no_token"))
			return validate.NewRequestError(errors.New("no token within request context"), http.StatusUnauthorized)
		}

		newToken, err := ctrl.svc.User.RenewToken(spanCtx, requestToken)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("refresh.outcome", "renew_failed"))
			return validate.NewUnauthorizedError()
		}

		span.SetAttributes(
			attribute.String("refresh.outcome", "success"),
			attribute.String("refresh.expires_at", newToken.ExpiresAt.Format(time.RFC3339)),
		)
		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, false, newToken.AttachmentToken)
		return server.JSON(w, http.StatusOK, newToken)
	}
}

func noPort(host string) string {
	return strings.Split(host, ":")[0]
}

func (ctrl *V1Controller) setCookies(w http.ResponseWriter, domain, token string, expires time.Time, remember bool, attachmentToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameRemember,
		Value:    strconv.FormatBool(remember),
		Expires:  expires,
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Set HTTP only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameToken,
		Value:    token,
		Expires:  expires,
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Set Fake Session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameSession,
		Value:    "true",
		Expires:  expires,
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Set attachment token cookie (accessible to frontend, not HttpOnly)
	if attachmentToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "hb.auth.attachment_token",
			Value:    attachmentToken,
			Expires:  expires,
			Domain:   domain,
			Secure:   ctrl.cookieSecure,
			HttpOnly: false,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func (ctrl *V1Controller) unsetCookies(w http.ResponseWriter, domain string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameToken,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameRemember,
		Value:    "false",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Set Fake Session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameSession,
		Value:    "false",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Unset attachment token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "hb.auth.attachment_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

// HandleOIDCLogin godoc
//
//	@Summary	OIDC Login Initiation
//	@Tags		Authentication
//	@Produce	json
//	@Success	302
//	@Router		/v1/users/login/oidc [GET]
func (ctrl *V1Controller) HandleOIDCLogin() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleOIDCLogin")
		defer span.End()

		if !ctrl.config.OIDC.Enabled {
			span.SetAttributes(attribute.String("oidc.outcome", "disabled"))
			return validate.NewRequestError(fmt.Errorf("OIDC is not enabled"), http.StatusForbidden)
		}

		if ctrl.oidcProvider == nil {
			span.SetAttributes(attribute.String("oidc.outcome", "provider_unavailable"))
			log.Error().Msg("OIDC provider not initialized")
			return validate.NewRequestError(errors.New("OIDC provider not available"), http.StatusInternalServerError)
		}

		_, err := ctrl.oidcProvider.InitiateOIDCFlow(w, r.WithContext(spanCtx))
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("oidc.outcome", "initiate_failed"))
		} else {
			span.SetAttributes(attribute.String("oidc.outcome", "initiated"))
		}
		return err
	}
}

// HandleOIDCCallback godoc
//
//	@Summary	OIDC Callback Handler
//	@Tags		Authentication
//	@Param		code	query	string	true	"Authorization code"
//	@Param		state	query	string	true	"State parameter"
//	@Success	302
//	@Router		/v1/users/login/oidc/callback [GET]
func (ctrl *V1Controller) HandleOIDCCallback() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleOIDCCallback",
			attribute.Bool("oidc.has_code", r.URL.Query().Get("code") != ""),
			attribute.Bool("oidc.has_state", r.URL.Query().Get("state") != ""),
		)
		defer span.End()

		if !ctrl.config.OIDC.Enabled {
			span.SetAttributes(attribute.String("oidc.outcome", "disabled"))
			return validate.NewRequestError(fmt.Errorf("OIDC is not enabled"), http.StatusForbidden)
		}

		if ctrl.oidcProvider == nil {
			span.SetAttributes(attribute.String("oidc.outcome", "provider_unavailable"))
			log.Error().Msg("OIDC provider not initialized")
			return validate.NewRequestError(errors.New("OIDC provider not available"), http.StatusInternalServerError)
		}

		newToken, err := ctrl.oidcProvider.HandleCallback(w, r.WithContext(spanCtx))
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("oidc.outcome", "callback_failed"))
			log.Err(err).Msg("OIDC callback failed")
			http.Redirect(w, r, ctrl.config.Web.AppBase+"?oidc_error=oidc_auth_failed", http.StatusFound)
			return nil
		}

		span.SetAttributes(
			attribute.String("oidc.outcome", "success"),
			attribute.String("session.expires_at", newToken.ExpiresAt.Format(time.RFC3339)),
		)
		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, true, newToken.AttachmentToken)
		http.Redirect(w, r, ctrl.config.Web.AppBase+"home", http.StatusFound)
		return nil
	}
}
