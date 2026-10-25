package authz

import (
	"context"

	"github.com/google/uuid"
)

// Viewer is the resolved authorization identity for one request, scoped to a
// single tenant. It is attached to the request context by the viewer
// middleware and consumed by ent privacy policies and query interceptors.
type Viewer struct {
	UserID   uuid.UUID
	TenantID uuid.UUID

	// Superuser marks an instance administrator (users.is_superuser). The
	// privacy layer allows superusers everything; tenant owners get no such
	// bypass.
	Superuser bool

	// Perms is the effective tenant-wide permission set: the membership's
	// direct permissions unioned with the permissions of every permission
	// group (in this tenant) the user belongs to.
	Perms map[Permission]struct{}

	// PermGroupIDs lists the permission groups (this tenant) the user is a
	// member of. Row-level access-grant predicates match against these.
	PermGroupIDs []uuid.UUID
}

// NewViewer builds a viewer from permission strings, ignoring unknown values.
func NewViewer(userID, tenantID uuid.UUID, superuser bool, perms []string, permGroupIDs []uuid.UUID) *Viewer {
	v := &Viewer{
		UserID:       userID,
		TenantID:     tenantID,
		Superuser:    superuser,
		Perms:        make(map[Permission]struct{}, len(perms)),
		PermGroupIDs: permGroupIDs,
	}
	v.AddPerms(perms)
	return v
}

// AddPerms unions additional permission strings into the viewer's set.
// Wildcards ("*", "<resource>:*") are expanded against the CURRENT catalog,
// so a stored ["*"] membership automatically covers permissions added to the
// catalog after it was written. Unknown values are ignored.
func (v *Viewer) AddPerms(perms []string) {
	for _, p := range Expand(perms) {
		v.Perms[p] = struct{}{}
	}
}

// Has reports whether the viewer holds the tenant-wide permission p.
// Superusers hold every permission.
func (v *Viewer) Has(p Permission) bool {
	if v == nil {
		return false
	}
	if v.Superuser {
		return true
	}
	_, ok := v.Perms[p]
	return ok
}

// PermStrings returns the effective permission set in catalog order.
func (v *Viewer) PermStrings() []string {
	if v == nil {
		return nil
	}
	out := make([]string, 0, len(v.Perms))
	for _, p := range allPermissions {
		if v.Has(p) {
			out = append(out, string(p))
		}
	}
	return out
}

type viewerCtxKey struct{}

// NewContext returns a context carrying the viewer.
func NewContext(ctx context.Context, v *Viewer) context.Context {
	return context.WithValue(ctx, viewerCtxKey{}, v)
}

// FromContext returns the viewer attached to ctx, or nil if none.
func FromContext(ctx context.Context) *Viewer {
	v, _ := ctx.Value(viewerCtxKey{}).(*Viewer)
	return v
}

type systemCtxKey struct{}

// NewSystemContext returns a context that bypasses the privacy layer
// entirely. It may only be used for flows whose inputs are already
// authenticated or are not user-driven: authentication itself (login,
// registration, token lookup), startup tasks, background jobs, and test
// bootstrapping. Never apply it inside HTTP handlers.
func NewSystemContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, systemCtxKey{}, true)
}

// IsSystem reports whether ctx is a system (privacy-bypass) context.
func IsSystem(ctx context.Context) bool {
	ok, _ := ctx.Value(systemCtxKey{}).(bool)
	return ok
}
