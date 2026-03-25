package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

type contextKeys struct {
	name string
}

var (
	ContextUser      = &contextKeys{name: "User"}
	ContextUserToken = &contextKeys{name: "UserToken"}
	ContextTenant    = &contextKeys{name: "Tenant"}
)

type Context struct {
	context.Context

	// UID is a unique identifier for the acting user.
	UID uuid.UUID

	// GID is a unique identifier for the acting users group.
	GID uuid.UUID

	// User is the acting user.
	User *repo.UserOut
}

// NewContext is a helper function that returns the service context from the context.
// This extracts the users from the context and embeds it into the ServiceContext struct
func NewContext(ctx context.Context) Context {
	user := UseUserCtx(ctx)
	gid := UseTenantCtx(ctx)

	var uid uuid.UUID
	if user != nil {
		uid = user.ID
		if gid == uuid.Nil {
			gid = user.DefaultGroupID
		}
	}

	return Context{
		Context: ctx,
		UID:     uid,
		GID:     gid,
		User:    user,
	}
}

// SetUserCtx is a helper function that sets the ContextUser and ContextUserToken
// values within the context of a web request (or any context).
func SetUserCtx(ctx context.Context, user *repo.UserOut, token string) context.Context {
	ctx = context.WithValue(ctx, ContextUser, user)
	ctx = context.WithValue(ctx, ContextUserToken, token)
	return ctx
}

// UseUserCtx is a helper function that returns the user from the context.
func UseUserCtx(ctx context.Context) *repo.UserOut {
	if val := ctx.Value(ContextUser); val != nil {
		return val.(*repo.UserOut)
	}
	return nil
}

// UseTokenCtx is a helper function that returns the user token from the context.
func UseTokenCtx(ctx context.Context) string {
	if val := ctx.Value(ContextUserToken); val != nil {
		return val.(string)
	}
	return ""
}

// UseTenantCtx is a helper function that returns the tenant group ID from the context.
// Returns uuid.Nil if not set.
func UseTenantCtx(ctx context.Context) uuid.UUID {
	if val := ctx.Value(ContextTenant); val != nil {
		return val.(uuid.UUID)
	}
	return uuid.Nil
}

// SetTenantCtx is a helper function that sets the ContextTenant in the context.
func SetTenantCtx(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, ContextTenant, tenantID)
}
