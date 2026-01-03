package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	v1 "github.com/sysadminsmedia/homebox/backend/app/api/handlers/v1"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

type tokenHasKey struct {
	key string
}

var hashedToken = tokenHasKey{key: "hashedToken"}

type RoleMode int

const (
	RoleModeOr  RoleMode = 0
	RoleModeAnd RoleMode = 1
)

// mwRoles is a middleware that will validate the required roles are met. All roles
// are required to be met for the request to be allowed. If the user does not have
// the required roles, a 403 Forbidden will be returned.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken or else it will panic
func (a *app) mwRoles(rm RoleMode, required ...string) errchain.Middleware {
	return func(next errchain.Handler) errchain.Handler {
		return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()

			maybeToken := ctx.Value(hashedToken)
			if maybeToken == nil {
				panic("mwRoles: token not found in context, you must call mwAuthToken before mwRoles")
			}

			token := maybeToken.(string)

			roles, err := a.repos.AuthTokens.GetRoles(r.Context(), token)
			if err != nil {
				return err
			}

		outer:
			switch rm {
			case RoleModeOr:
				for _, role := range required {
					if roles.Contains(role) {
						break outer
					}
				}
				return validate.NewRequestError(errors.New("Forbidden"), http.StatusForbidden)
			case RoleModeAnd:
				for _, req := range required {
					if !roles.Contains(req) {
						return validate.NewRequestError(errors.New("Unauthorized"), http.StatusForbidden)
					}
				}
			}

			return next.ServeHTTP(w, r)
		})
	}
}

type KeyFunc func(r *http.Request) (string, error)

func getBearer(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("authorization header is required")
	}

	return auth, nil
}

func getQuery(r *http.Request) (string, error) {
	token := r.URL.Query().Get("access_token")
	if token == "" {
		return "", errors.New("access_token query is required")
	}

	token, err := url.QueryUnescape(token)
	if err != nil {
		return "", errors.New("access_token query is required")
	}

	return token, nil
}

// mwAuthToken is a middleware that will check the database for a stateful token
// and attach it's user to the request context, or return an appropriate error.
// Authorization support is by token via Headers or Query Parameter
//
// Example:
//   - header = "Bearer 1234567890"
//   - query = "?access_token=1234567890"
func (a *app) mwAuthToken(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var requestToken string

		// We ignore the error to allow the next strategy to be attempted
		{
			cookies, _ := v1.GetCookies(r)
			if cookies != nil {
				requestToken = cookies.Token
			}
		}

		if requestToken == "" {
			keyFuncs := [...]KeyFunc{
				getBearer,
				getQuery,
			}

			for _, keyFunc := range keyFuncs {
				token, err := keyFunc(r)
				if err == nil {
					requestToken = token
					break
				}
			}
		}

		if requestToken == "" {
			return validate.NewRequestError(errors.New("authorization header or query is required"), http.StatusUnauthorized)
		}

		requestToken = strings.TrimPrefix(requestToken, "Bearer ")

		r = r.WithContext(context.WithValue(r.Context(), hashedToken, requestToken))

		usr, err := a.services.User.GetSelf(r.Context(), requestToken)
		// Check the database for the token
		if err != nil {
			if ent.IsNotFound(err) {
				return validate.NewRequestError(errors.New("valid authorization token is required"), http.StatusUnauthorized)
			}

			return err
		}

		r = r.WithContext(services.SetUserCtx(r.Context(), &usr, requestToken))
		return next.ServeHTTP(w, r)
	})
}

// mwTenant is a middleware that will parse the X-Tenant header and validate the user has access
// to the requested tenant. If no header is provided, the user's default group is used.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken
func (a *app) mwTenant(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		// Get the user from context (set by mwAuthToken)
		user := services.UseUserCtx(ctx)
		if user == nil {
			return validate.NewRequestError(errors.New("user context not found"), http.StatusInternalServerError)
		}

		tenantID := user.DefaultGroupID

		// Check for X-Tenant header or tenant query parameter
		tenantHeader := r.Header.Get("X-Tenant")
		if tenantHeader == "" {
			tenantHeader = r.URL.Query().Get("tenant")
		}

		if tenantHeader != "" {
			parsedTenantID, err := uuid.Parse(tenantHeader)
			if err != nil {
				return validate.NewRequestError(errors.New("invalid X-Tenant header format"), http.StatusBadRequest)
			}

			// Validate user has access to the requested tenant
			hasAccess := false
			for _, gid := range user.GroupIDs {
				if gid == parsedTenantID {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				return validate.NewRequestError(errors.New("user does not have access to the requested tenant"), http.StatusForbidden)
			}

			tenantID = parsedTenantID
		}

		// Set the tenant in context
		r = r.WithContext(services.SetTenantCtx(ctx, tenantID))
		return next.ServeHTTP(w, r)
	})
}
