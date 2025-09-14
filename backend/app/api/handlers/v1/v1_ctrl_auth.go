package v1

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/app/api/providers"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
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

	OIDCAuthURLResponse struct {
		AuthURL string `json:"authUrl"`
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

	providers := make(map[string]AuthProvider)
	for _, p := range ps {
		log.Info().Str("name", p.Name()).Msg("registering auth provider")
		providers[p.Name()] = p
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		// Extract provider query
		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "local"
		}

		// Get the provider
		p, ok := providers[provider]
		if !ok {
			return validate.NewRequestError(errors.New("invalid auth provider"), http.StatusBadRequest)
		}

		newToken, err := p.Authenticate(w, r)
		if err != nil {
			log.Err(err).Msg("failed to authenticate")
			return server.JSON(w, http.StatusInternalServerError, err.Error())
		}

		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, true)
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
		token := services.UseTokenCtx(r.Context())
		if token == "" {
			return validate.NewRequestError(errors.New("no token within request context"), http.StatusUnauthorized)
		}

		err := ctrl.svc.User.Logout(r.Context(), token)
		if err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

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
		requestToken := services.UseTokenCtx(r.Context())
		if requestToken == "" {
			return validate.NewRequestError(errors.New("no token within request context"), http.StatusUnauthorized)
		}

		newToken, err := ctrl.svc.User.RenewToken(r.Context(), requestToken)
		if err != nil {
			return validate.NewUnauthorizedError()
		}

		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, false)
		return server.JSON(w, http.StatusOK, newToken)
	}
}

// HandleOIDCAuthURL godoc
//
//	@Summary		Get OIDC Authorization URL
//	@Description	Returns the OIDC authorization URL for SSO login
//	@Tags			Authentication
//	@Success		200 {object} OIDCAuthURLResponse
//	@Failure		404
//	@Router			/v1/users/oidc/auth-url [GET]
func (ctrl *V1Controller) HandleOIDCAuthURL(provider AuthProvider) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Check if OIDC provider is available
		if provider == nil || provider.Name() != "OIDC" {
			return validate.NewRequestError(errors.New("OIDC provider not configured"), http.StatusNotFound)
		}

		// Get the OIDC provider and generate auth URL
		oidcProvider, ok := provider.(*providers.OAuthProvider)
		if !ok {
			return validate.NewRequestError(errors.New("invalid OIDC provider"), http.StatusInternalServerError)
		}

		// Generate state parameter for security
		state := generateRandomState()
		
		// Store state in session/cookie for validation
		http.SetCookie(w, &http.Cookie{
			Name:     "oidc_state",
			Value:    state,
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			Secure:   ctrl.cookieSecure,
			Path:     "/",
		})
		
		// Get the authorization URL
		authURL := oidcProvider.GetAuthURL(state)
		
		return server.JSON(w, http.StatusOK, OIDCAuthURLResponse{
			AuthURL: authURL,
		})
	}
}

func noPort(host string) string {
	return strings.Split(host, ":")[0]
}

func generateRandomState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (ctrl *V1Controller) setCookies(w http.ResponseWriter, domain, token string, expires time.Time, remember bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameRemember,
		Value:    strconv.FormatBool(remember),
		Expires:  expires,
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
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
	})
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
	})

	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameRemember,
		Value:    "false",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		Secure:   ctrl.cookieSecure,
		HttpOnly: true,
		Path:     "/",
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
	})
}
