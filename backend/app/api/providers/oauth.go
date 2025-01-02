package providers

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"golang.org/x/oauth2"
	"net/http"
)

type OAuthProvider struct {
	name    string
	service *services.OAuthService
	config  *services.OAuthConfig
}

func NewOAuthProvider(ctx context.Context, service *services.OAuthService, clientId string, clientSecret string, redirectUri string, providerUrl string) (*OAuthProvider, error) {
	// TODO: fallback for all variabnles if no well known is supported
	if providerUrl == "" {
		return nil, errors.New("Provider url not given")
	}
	provider, err := oidc.NewProvider(ctx, providerUrl)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("AuthUrl", provider.Endpoint().AuthURL).Msg("discovered oauth provider")

	return &OAuthProvider{
		name:    "OIDC",
		service: service,
		config: &services.OAuthConfig{
			Config: &oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Endpoint:     provider.Endpoint(),
				RedirectURL:  redirectUri,
				Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			},
			Provider: provider,
			Verifier: provider.Verifier(&oidc.Config{ClientID: clientId}),
		},
	}, nil
}

func (p *OAuthProvider) Name() string {
	return p.name
}

func (p *OAuthProvider) Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	oauthForm, err := getOAuthForm(r)
	if err != nil {
		return services.UserAuthTokenDetail{}, err
	}

	return p.service.Login(r.Context(), p.config, oauthForm)
}
