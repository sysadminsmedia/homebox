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

type (
	ForgotPasswordRequest struct {
		Email string `json:"email" validate:"required" example:"user@example.com"`
	}

	// ResetPasswordRequest carries the token from the email link and the new
	// password. The constraints below feed the OpenAPI spec via swaggo and
	// are also enforced by the handler so the spec doesn't over-promise.
	//
	// password min=6 matches the frontend's PASSWORD_MIN_LENGTH. No max:
	// argon2id has no practical input limit, and the inbound body is already
	// bounded by mid.MaxBodySize. Token min=20 fits the 26-char base32 output
	// of hasher.GenerateToken with a little slack; it's spec-only since the
	// lookup itself rejects anything that doesn't match a stored hash.
	ResetPasswordRequest struct {
		Token       string `json:"token"    validate:"required,min=20"`
		NewPassword string `json:"password" validate:"required,min=6"`
	}
)

// resetPasswordMinLength is enforced server-side to keep the OpenAPI spec
// honest about its minLength constraint.
const resetPasswordMinLength = 6

// HandleForgotPassword godoc
//
//	@Summary		Request Password Reset
//	@Description	Sends a password reset email if the address is associated with a local account.
//	@Description	Always returns 204 on success to avoid leaking whether the email is registered.
//	@Tags			Authentication
//	@Accept			application/json
//	@Produce		json
//	@Param			payload	body	ForgotPasswordRequest	true	"Email"
//	@Success		204
//	@Failure		400	{string}	string	"missing or invalid request body, or empty email field"
//	@Failure		403	{string}	string	"demo mode is enabled or local login is disabled"
//	@Failure		500	{string}	string	"internal error while processing the request"
//	@Router			/v1/users/forgot-password [POST]
func (ctrl *V1Controller) HandleForgotPassword() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleForgotPassword")
		defer span.End()

		if ctrl.isDemo {
			span.SetAttributes(attribute.String("forgot.outcome", "demo_blocked"))
			return validate.NewRequestError(nil, http.StatusForbidden)
		}

		if !ctrl.config.Options.AllowLocalLogin {
			span.SetAttributes(attribute.String("forgot.outcome", "local_login_disabled"))
			return validate.NewRequestError(errors.New("local login is not enabled"), http.StatusForbidden)
		}

		var body ForgotPasswordRequest
		if err := server.Decode(r, &body); err != nil {
			span.SetAttributes(attribute.String("forgot.outcome", "decode_failed"))
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		body.Email = strings.TrimSpace(body.Email)
		span.SetAttributes(attribute.Int("user.email.length", len(body.Email)))
		if body.Email == "" {
			span.SetAttributes(attribute.String("forgot.outcome", "missing_email"))
			return validate.NewRequestError(errors.New("email is required"), http.StatusBadRequest)
		}

		// SECURITY: The two configuration failures below (SMTP not ready,
		// no safe base URL) MUST NOT be reported to the client. The whole
		// point of the "always 204" response is that an unauthenticated
		// caller can't probe the server for state — including configuration
		// state. Returning 503 here would let an attacker fingerprint
		// whether the instance has SMTP configured or what HBOX_OPTIONS_*
		// values are set, which is reconnaissance info we deliberately
		// withhold. Operators discover these via server logs instead.
		if !ctrl.svc.User.MailerReady() {
			span.SetAttributes(attribute.String("forgot.outcome", "mailer_not_configured"))
			log.Warn().Msg("forgot-password requested but SMTP mailer is not configured; no email will be sent")
			return server.JSON(w, http.StatusNoContent, nil)
		}

		// SecureBaseURL refuses Referer-based fallback so an attacker can't
		// poison the link in the victim's email by sending a forged Referer.
		baseURL := SecureBaseURL(r, &ctrl.config.Options)
		if baseURL == "" {
			span.SetAttributes(attribute.String("forgot.outcome", "no_safe_base_url"))
			log.Warn().Msg("forgot-password requested but no safe base URL is available; set HBOX_OPTIONS_HOSTNAME to enable")
			return server.JSON(w, http.StatusNoContent, nil)
		}

		if err := ctrl.svc.User.RequestPasswordReset(spanCtx, body.Email, baseURL); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("forgot.outcome", "service_failed"))
			log.Err(err).Msg("password reset request failed")
			// Don't leak the underlying error to the client; respond with 500
			// but no details.
			return validate.NewRequestError(errors.New("internal error"), http.StatusInternalServerError)
		}

		span.SetAttributes(attribute.String("forgot.outcome", "ok"))
		return server.JSON(w, http.StatusNoContent, nil)
	}
}

// HandleResetPassword godoc
//
//	@Summary		Reset Password
//	@Description	Consumes a single-use reset token and changes the user's password.
//	@Description	On success, all existing sessions for the user are revoked.
//	@Tags			Authentication
//	@Accept			application/json
//	@Produce		json
//	@Param			payload	body	ResetPasswordRequest	true	"Token + new password"
//	@Success		204
//	@Failure		400	{string}	string	"invalid request body, password shorter than the minimum length, or token that is invalid / expired / already used"
//	@Failure		403	{string}	string	"demo mode is enabled or local login is disabled"
//	@Failure		500	{string}	string	"internal error while processing the request"
//	@Router			/v1/users/reset-password [POST]
func (ctrl *V1Controller) HandleResetPassword() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleResetPassword")
		defer span.End()

		if ctrl.isDemo {
			span.SetAttributes(attribute.String("reset.outcome", "demo_blocked"))
			return validate.NewRequestError(nil, http.StatusForbidden)
		}

		if !ctrl.config.Options.AllowLocalLogin {
			span.SetAttributes(attribute.String("reset.outcome", "local_login_disabled"))
			return validate.NewRequestError(errors.New("local login is not enabled"), http.StatusForbidden)
		}

		var body ResetPasswordRequest
		if err := server.Decode(r, &body); err != nil {
			span.SetAttributes(attribute.String("reset.outcome", "decode_failed"))
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		span.SetAttributes(
			attribute.Int("token.length", len(body.Token)),
			attribute.Int("password.new.length", len(body.NewPassword)),
		)

		// Enforce the documented minimum so the OpenAPI spec's minLength on
		// `password` is honored. Token bounds are spec-only (the hash lookup
		// naturally rejects bad tokens), so we don't re-check them here.
		if len(body.NewPassword) < resetPasswordMinLength {
			span.SetAttributes(attribute.String("reset.outcome", "password_too_short"))
			return validate.NewRequestError(
				fmt.Errorf("password must be at least %d characters", resetPasswordMinLength),
				http.StatusBadRequest,
			)
		}

		if err := ctrl.svc.User.ResetPassword(spanCtx, body.Token, body.NewPassword); err != nil {
			if errors.Is(err, services.ErrorPasswordResetInvalid) {
				span.SetAttributes(attribute.String("reset.outcome", "token_invalid"))
				return validate.NewRequestError(err, http.StatusBadRequest)
			}
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("reset.outcome", "service_failed"))
			log.Err(err).Msg("password reset failed")
			return validate.NewRequestError(errors.New("internal error"), http.StatusInternalServerError)
		}

		span.SetAttributes(attribute.String("reset.outcome", "success"))
		return server.JSON(w, http.StatusNoContent, nil)
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

		// API keys are not session tokens — they live in their own table and
		// are managed from the profile page. Returning 204 here would silently
		// delete zero rows from auth_tokens and mislead the caller.
		if services.IsAPIKeyAuth(spanCtx) {
			span.SetAttributes(attribute.String("logout.outcome", "api_key_rejected"))
			return validate.NewRequestError(
				errors.New("API keys cannot be logged out; revoke them from the API keys page"),
				http.StatusBadRequest,
			)
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

		// API keys are long-lived and don't go through the session refresh
		// flow. Reject the request with a clear error rather than the generic
		// 401 RenewToken would otherwise produce.
		if services.IsAPIKeyAuth(spanCtx) {
			span.SetAttributes(attribute.String("refresh.outcome", "api_key_rejected"))
			return validate.NewRequestError(
				errors.New("API keys do not require refresh"),
				http.StatusBadRequest,
			)
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
		Path:     ctrl.config.Web.AppBase,
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
		Path:     ctrl.config.Web.AppBase,
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
		Path:     ctrl.config.Web.AppBase,
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
			Path:     ctrl.config.Web.AppBase,
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
		Path:     ctrl.config.Web.AppBase,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameRemember,
		Value:    "false",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     ctrl.config.Web.AppBase,
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
		Path:     ctrl.config.Web.AppBase,
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
		Path:     ctrl.config.Web.AppBase,
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
