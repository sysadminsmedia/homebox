package providers

import (
	"net/http"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
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
	loginForm, err := getLoginForm(r)
	if err != nil {
		return services.UserAuthTokenDetail{}, err
	}

	return p.service.Login(r.Context(), loginForm.Username, loginForm.Password, loginForm.StayLoggedIn)
}
