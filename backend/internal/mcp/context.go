package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
)

// ErrNoAuthContext is returned when a tool runs without the auth/tenant
// middleware having populated the request context. The MCP route is mounted
// behind userMW, so this should only fire if a tool is invoked outside the
// HTTP transport (e.g. a misuse from tests).
var ErrNoAuthContext = errors.New("mcp: request has no authenticated user/tenant context")

// ErrGroupNotMember is returned when a tool requests a group_id the calling
// user is not a member of. This is the safety check that lets tools accept
// an optional group_id input — the request is rejected if the user has no
// access to the named group.
var ErrGroupNotMember = errors.New("mcp: user is not a member of the requested group")

// ServiceCtx returns the services.Context (UID, GID, User) for the calling
// MCP request. The GID is whatever the X-Tenant header resolved to (or the
// user's default group when the header is absent), so single-group users
// get the right scope automatically.
//
// Tools that accept an optional group_id input should use ResolveGroup
// instead, which validates membership before switching scope.
func ServiceCtx(ctx context.Context) (services.Context, error) {
	sctx := services.NewContext(ctx)
	if sctx.UID == uuid.Nil || sctx.GID == uuid.Nil {
		return services.Context{}, ErrNoAuthContext
	}
	return sctx, nil
}

// ResolveGroup returns a services.Context scoped to the requested group.
// If requested is uuid.Nil the user's tenant/default group is used. Otherwise
// the user must be a member of the requested group; ResolveGroup returns
// ErrGroupNotMember if not. This mirrors the membership check mwTenant
// performs for HTTP requests, but lets MCP tools take group_id as input so
// the LLM can pivot among the user's groups within a single session.
func ResolveGroup(ctx context.Context, requested uuid.UUID) (services.Context, error) {
	sctx, err := ServiceCtx(ctx)
	if err != nil {
		return services.Context{}, err
	}
	if requested == uuid.Nil {
		return sctx, nil
	}
	if sctx.User == nil {
		return services.Context{}, ErrNoAuthContext
	}
	for _, gid := range sctx.User.GroupIDs {
		if gid == requested {
			sctx.GID = requested
			return sctx, nil
		}
	}
	return services.Context{}, fmt.Errorf("%w: %s", ErrGroupNotMember, requested)
}
