package providers

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"strings"
)

type OAuthProvider struct {
	name    string
	service *services.OAuthService
	config  *services.OAuthConfig
}

func NewOAuthProvider(ctx context.Context, service *services.OAuthService, name string) (*OAuthProvider, error) {
	upperName := strings.ToUpper(name)
	clientId := os.Getenv(fmt.Sprintf("HBOX_OAUTH_%S_ID", upperName))
	clientSecret := os.Getenv(fmt.Sprintf("HBOX_OAUTH_%s_SECRET", upperName))
	redirectUri := os.Getenv(fmt.Sprintf("HBOX_OAUTH_%s_REDIRECT", upperName))

	providerUrl := os.Getenv(fmt.Sprintf("HBOX_OAUTH_%s_URL", upperName))
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
		name:    name,
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
