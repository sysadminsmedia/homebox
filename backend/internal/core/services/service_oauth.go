package services

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"golang.org/x/oauth2"
)

type OAuthService struct {
	repos *repo.AllRepos
	user  *UserService
}

type (
	OAuthConfig struct {
		Config   *oauth2.Config
		Provider *oidc.Provider
		Verifier *oidc.IDTokenVerifier
	}
	OAuthValidate struct {
		Issuer string `json:"iss"`
		Code   string `json:"code"`
		State  string `json:"state"`
	}
	OAuthUserRegistration struct {
		Issuer  string `json:"iss"`
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}
	OAuthIdClaims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
)

func (svc *OAuthService) Login(ctx context.Context, config *OAuthConfig, data OAuthValidate) (UserAuthTokenDetail, error) {
	usr, err := svc.ValidateCode(ctx, config, data)
	if err != nil {
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	return svc.user.createSessionToken(ctx, usr.ID, false)
}

func (svc *OAuthService) ValidateCode(ctx context.Context, config *OAuthConfig, data OAuthValidate) (repo.UserOut, error) {
	log.Debug().Str("ClientId", config.Config.ClientID).Msg("Exchanging code")
	token, err := config.Config.Exchange(ctx, data.Code)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to exchange code")
		return repo.UserOut{}, err
	}

	user, ok, err := svc.LoginWithIdToken(ctx, config, token)
	if err != nil {
		return repo.UserOut{}, err
	}
	if !ok {
		panic("Id token check not ok") // TODO: fallback to user info
	}
	return user, nil
}

func (svc *OAuthService) LoginWithIdToken(ctx context.Context, config *OAuthConfig, token *oauth2.Token) (repo.UserOut, bool, error) {
	if config.Verifier == nil {
		return repo.UserOut{}, false, nil
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return repo.UserOut{}, ok, nil
	}

	idToken, err := config.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return repo.UserOut{}, false, err
	}

	log.Debug().
		Str("iss", idToken.Issuer).
		Str("sub", idToken.Subject).
		Msg("Searching for user")

	user, err := svc.repos.OAuth.GetUserFromToken(ctx, idToken.Issuer, idToken.Subject)
	if err != nil {
		var notFoundError *ent.NotFoundError
		notFound := errors.As(err, &notFoundError)
		if notFound {
			// User does not exist, create and link a new one
			user, err := svc.CreateUserIdToken(ctx, idToken)
			return user, true, err
		}
	}
	return user, true, err
}

func (svc *OAuthService) CreateUserIdToken(ctx context.Context, token *oidc.IDToken) (repo.UserOut, error) {
	var claims OAuthIdClaims
	if err := token.Claims(&claims); err != nil {
		return repo.UserOut{}, err
	}

	// Check that user does not yet exist so that we don't clash
	if _, err := svc.repos.Users.GetOneEmail(ctx, claims.Email); err != nil {
		var notFoundError *ent.NotFoundError
		if notFound := errors.As(err, &notFoundError); !notFound {
			return repo.UserOut{}, err
		}
	} else {
		return repo.UserOut{}, errors.New("cannot create OAuth connection")
	}

	registration := OAuthUserRegistration{
		Issuer:  token.Issuer,
		Subject: token.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
	}
	return svc.CreateUser(ctx, registration)
}

func (svc *OAuthService) CreateUser(ctx context.Context, registration OAuthUserRegistration) (repo.UserOut, error) {
	log.Debug().
		Str("Subject", registration.Subject).
		Str("Issuer", registration.Issuer).
		Str("name", registration.Name).
		Msg("Registering new OAuth user")

	var groupId uuid.UUID
	if group, err := svc.repos.Groups.GroupByName(ctx, "OAuth"); err == nil {
		log.Debug().Msg("joining existing oauth group")
		groupId = group.ID
	} else {
		var notFoundError *ent.NotFoundError
		if notFound := errors.As(err, &notFoundError); notFound {
			log.Debug().Msg("Creating new oauth group")
			group, err := svc.repos.Groups.GroupCreate(ctx, "OAuth")
			if err != nil {
				log.Err(err).Msg("Failed to create group")
				return repo.UserOut{}, err
			}
			groupId = group.ID
			err = createDefaultLabels(ctx, svc.repos, groupId)
			if err != nil {
				return repo.UserOut{}, err
			}
		} else {
			return repo.UserOut{}, err
		}
	}

	usrCreate := repo.UserCreate{
		Name:        registration.Name,
		Email:       registration.Email,
		IsSuperuser: false, // TODO: use role to check if superuser
		GroupID:     groupId,
		IsOwner:     false, // TODO: use role to check if owner?
	}
	usr, err := svc.repos.Users.Create(ctx, usrCreate)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create user")
		return repo.UserOut{}, err
	}

	oauthCreate := repo.OAuthCreate{
		Provider: registration.Issuer,
		Subject:  registration.Subject,
		UserId:   usr.ID,
	}
	_, err = svc.repos.OAuth.Create(ctx, oauthCreate)
	if err != nil {
		return repo.UserOut{}, err
	}

	log.Debug().Msg("OAuth User created")
	return usr, nil
}
