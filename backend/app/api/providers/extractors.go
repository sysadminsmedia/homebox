package providers

import (
	"errors"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"net/http"

	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

type LoginForm struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	StayLoggedIn bool   `json:"stayLoggedIn"`
}

func getLoginForm(r *http.Request) (LoginForm, error) {
	loginForm := LoginForm{}

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			return loginForm, errors.New("failed to parse form")
		}

		loginForm.Username = r.PostFormValue("username")
		loginForm.Password = r.PostFormValue("password")
		loginForm.StayLoggedIn = r.PostFormValue("stayLoggedIn") == "true"
	case "application/json":
		err := server.Decode(r, &loginForm)
		if err != nil {
			log.Err(err).Msg("failed to decode login form")
			return loginForm, errors.New("failed to decode login form")
		}
	default:
		return loginForm, errors.New("invalid content type")
	}

	if loginForm.Username == "" || loginForm.Password == "" {
		return loginForm, validate.NewFieldErrors(
			validate.FieldError{
				Field: "username",
				Error: "username or password is empty",
			},
			validate.FieldError{
				Field: "password",
				Error: "username or password is empty",
			},
		)
	}

	return loginForm, nil
}

func getOAuthForm(r *http.Request) (services.OAuthValidate, error) {
	var oauthForm services.OAuthValidate
	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			return oauthForm, errors.New("failed to parse form")
		}

		oauthForm.Issuer = r.PostFormValue("issuer")
		oauthForm.Code = r.PostFormValue("code")
		oauthForm.State = r.PostFormValue("state")
	case "application/json":
		err := server.Decode(r, &oauthForm)
		if err != nil {
			log.Err(err).Msg("failed to decode OAuth form")
			return oauthForm, err
		}
	default:
		return oauthForm, errors.New("invalid content type")
	}

	if oauthForm.Issuer == "" || oauthForm.Code == "" {
		return oauthForm, validate.NewFieldErrors(
			validate.FieldError{
				Field: "iss",
				Error: "Issuer is empty",
			},
			validate.FieldError{
				Field: "code",
				Error: "Code is missing",
			},
		)
	}

	return oauthForm, nil
}
