// Package authzrules implements the ent privacy policies and query
// interceptors that enforce the Homebox permission system at the ORM layer.
//
// Design:
//   - Query side: every schema has an interceptor that injects a WHERE clause
//     scoping results to what the request viewer may read (tenant membership,
//     tenant-wide permissions, and row-level access grants). Inaccessible rows
//     are invisible, so single-row lookups fail with NotFoundError (404, no
//     existence leak).
//   - Mutation side: every schema has a privacy policy that maps the mutation
//     to a required permission and *pins* update/delete statements to the
//     viewer's tenant (or to row-grant-covered ids) by adding predicates to
//     the mutation itself. A cross-tenant write therefore matches zero rows
//     even if a repository forgets a scope filter.
//   - System contexts (authz.NewSystemContext) bypass everything; they are
//     reserved for authentication flows, startup tasks, and background jobs.
//   - There is no superuser bypass rule: authz.Viewer.Has reports true for
//     every permission when the viewer is an instance superuser, so
//     superusers pass permission checks but remain tenant-scoped on reads.
//
// This package may import the generated ent packages (the runtime stitching
// lives in ent/runtime, so there is no import cycle with ent/schema), but it
// must never import ent/runtime itself.
package authzrules

import (
	"context"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
)

// AllowIfSystem allows queries and mutations running under a system context.
func AllowIfSystem() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if authz.IsSystem(ctx) {
			return privacy.Allow
		}
		return privacy.Skip
	})
}

// DenyIfNoViewer denies queries and mutations that carry no viewer. Combined
// with AllowIfSystem this makes "forgot to attach a viewer" fail closed.
func DenyIfNoViewer() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if authz.FromContext(ctx) == nil {
			return privacy.Denyf("authz: no viewer in context")
		}
		return privacy.Skip
	})
}

// NewPolicy builds the uniform policy shape used by every schema:
//
//	Query:    system bypass -> deny if no viewer -> allow (row filtering is
//	          the job of the schema's query interceptor)
//	Mutation: system bypass -> deny if no viewer -> schema rules -> deny
//
// The trailing AlwaysDenyRule makes mutations default-deny: a schema rule
// must explicitly Allow.
func NewPolicy(mutation ...privacy.MutationRule) privacy.Policy {
	mutations := privacy.MutationPolicy{AllowIfSystem(), DenyIfNoViewer()}
	mutations = append(mutations, mutation...)
	mutations = append(mutations, privacy.AlwaysDenyRule())

	return privacy.Policy{
		Query: privacy.QueryPolicy{
			AllowIfSystem(),
			DenyIfNoViewer(),
			privacy.AlwaysAllowRule(),
		},
		Mutation: mutations,
	}
}

// viewerFor resolves the viewer for rule evaluation. It returns (nil, nil)
// for system contexts (caller should not filter) and (nil, deny) when no
// viewer is attached.
func viewerFor(ctx context.Context) (*authz.Viewer, error) {
	if authz.IsSystem(ctx) {
		return nil, nil
	}
	v := authz.FromContext(ctx)
	if v == nil {
		return nil, privacy.Denyf("authz: no viewer in context")
	}
	return v, nil
}
