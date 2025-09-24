package v1

import (
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/app/api/providers"
)

type OIDCConfigResponse struct {
	Enabled bool   `json:"enabled"`
	AuthURL string `json:"authUrl,omitempty"`
}

// HandleOIDCConfig godoc
//
//	@Summary	Get OIDC Configuration
//	@Tags		Authentication
//	@Produce	json
//	@Success	200	{object}	OIDCConfigResponse
//	@Router		/v1/auth/oidc/config [GET]
func (ctrl *V1Controller) HandleOIDCConfig() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		response := OIDCConfigResponse{
			Enabled: ctrl.config.OIDC.Enabled,
		}

		if ctrl.config.OIDC.Enabled {
			response.AuthURL = "/api/v1/auth/oidc/login"
		}

		return server.JSON(w, http.StatusOK, response)
	}
}

// HandleOIDCLogin godoc
//
//	@Summary	Initiate OIDC Login
//	@Tags		Authentication
//	@Success	302	"Redirect to OIDC provider"
//	@Router		/v1/auth/oidc/login [GET]
func (ctrl *V1Controller) HandleOIDCLogin(oidcProvider *providers.OIDCProvider) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		_, err := oidcProvider.InitiateAuth(w, r)
		if err != nil {
			return server.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		// InitiateAuth handles the redirect, so we don't need to return anything
		return nil
	}
}

// HandleOIDCCallback godoc
//
//	@Summary	Handle OIDC Callback
//	@Tags		Authentication
//	@Param		code	query	string	true	"Authorization code"
//	@Param		state	query	string	true	"State parameter"
//	@Success	302		"Redirect to home page"
//	@Router		/v1/auth/oidc/callback [GET]
func (ctrl *V1Controller) HandleOIDCCallback(oidcProvider *providers.OIDCProvider) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		newToken, err := oidcProvider.HandleCallback(w, r)
		if err != nil {
			return server.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		ctrl.setCookies(w, noPort(r.Host), newToken.Raw, newToken.ExpiresAt, true)
		
		// Redirect to home page instead of returning JSON
		http.Redirect(w, r, "/home", http.StatusFound)
		return nil
	}
}