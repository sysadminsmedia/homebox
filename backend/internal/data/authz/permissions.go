// Package authz defines the permission catalog and the request viewer used by
// the ent privacy layer. It is intentionally dependency-free (stdlib + uuid)
// so it can be imported from ent schemas, generated-code helpers, services,
// and handlers without import cycles.
package authz

import "strings"

// Permission is a single grantable capability, expressed as "resource:action".
type Permission string

const (
	// Entity (inventory item / location tree) permissions. These also govern
	// child resources that hang off an entity: attachments, maintenance
	// entries, and custom fields.
	PermEntityRead   Permission = "entity:read"
	PermEntityCreate Permission = "entity:create"
	PermEntityUpdate Permission = "entity:update"
	PermEntityDelete Permission = "entity:delete"

	// Reference-data management. Reading reference data (tags, entity types,
	// templates) requires only tenant membership.
	PermTagManage        Permission = "tag:manage"
	PermEntityTypeManage Permission = "entitytype:manage"
	PermTemplateManage   Permission = "template:manage"
	PermNotifierManage   Permission = "notifier:manage"

	// Bulk data movement. NOTE: data:export is a coarse "read everything in the
	// tenant" capability — the export job dumps all tenant rows via raw SQL and
	// intentionally bypasses row-level access grants. Do not grant it to members
	// who are meant to be restricted to a subset of entities.
	PermDataExport Permission = "data:export"
	PermDataImport Permission = "data:import"

	// Tenant administration.
	PermSettingsManage    Permission = "settings:manage"
	PermMembersManage     Permission = "members:manage"
	PermPermissionsManage Permission = "permissions:manage"
)

// allPermissions is the canonical ordered catalog. Order is stable and used
// for API output; append new permissions at the end of their resource block.
var allPermissions = []Permission{
	PermEntityRead,
	PermEntityCreate,
	PermEntityUpdate,
	PermEntityDelete,
	PermTagManage,
	PermEntityTypeManage,
	PermTemplateManage,
	PermNotifierManage,
	PermDataExport,
	PermDataImport,
	PermSettingsManage,
	PermMembersManage,
	PermPermissionsManage,
}

// All returns a copy of the full permission catalog.
func All() []Permission {
	out := make([]Permission, len(allPermissions))
	copy(out, allPermissions)
	return out
}

// AllStrings returns the full catalog as plain strings, the storage shape
// used by JSON permission-list columns.
func AllStrings() []string {
	out := make([]string, len(allPermissions))
	for i, p := range allPermissions {
		out[i] = string(p)
	}
	return out
}

// --- Wildcards ---------------------------------------------------------------
//
// Stored permission lists support wildcards so that grants stay valid as the
// catalog grows. "*" means every permission, present and future; "<resource>:*"
// means every action on one resource. Full-access memberships are stored as
// ["*"] (never as an enumerated snapshot), so adding a new permission to the
// catalog automatically reaches them. Explicitly restricted lists do NOT grow
// — new permissions are denied until granted, which is the fail-closed
// direction.

// Wildcard grants every permission, present and future.
const Wildcard = "*"

// FullAccess is the canonical stored shape of an unrestricted permission
// list. Use this — not AllStrings — wherever "everything" is persisted.
func FullAccess() []string {
	return []string{Wildcard}
}

// Resource returns p's resource segment ("entity:read" -> "entity").
func (p Permission) Resource() string {
	if i := strings.IndexByte(string(p), ':'); i >= 0 {
		return string(p)[:i]
	}
	return string(p)
}

// resourceWildcard reports whether s is a "<resource>:*" wildcard for a
// resource that exists in the catalog.
func resourceWildcard(s string) (string, bool) {
	const suffix = ":*"
	if len(s) <= len(suffix) || s[len(s)-len(suffix):] != suffix {
		return "", false
	}
	resource := s[:len(s)-len(suffix)]
	for _, known := range allPermissions {
		if known.Resource() == resource {
			return resource, true
		}
	}
	return "", false
}

// Valid reports whether p is a known permission or a recognized wildcard.
func Valid(p Permission) bool {
	if p == Wildcard {
		return true
	}
	if _, ok := resourceWildcard(string(p)); ok {
		return true
	}
	for _, known := range allPermissions {
		if p == known {
			return true
		}
	}
	return false
}

// ValidateStrings returns the subset of perms that are not known permissions
// or wildcards. An empty result means every entry is valid.
func ValidateStrings(perms []string) []string {
	var invalid []string
	for _, p := range perms {
		if !Valid(Permission(p)) {
			invalid = append(invalid, p)
		}
	}
	return invalid
}

// SetHas reports whether a stored permission list (which may contain
// wildcards) covers perm. Use this when evaluating raw stored lists; viewers
// resolve through Viewer.Has instead.
func SetHas(perms []string, perm Permission) bool {
	for _, p := range perms {
		if p == Wildcard || p == string(perm) {
			return true
		}
		if r, ok := resourceWildcard(p); ok && r == perm.Resource() {
			return true
		}
	}
	return false
}

// Expand resolves a stored permission list (with wildcards) into the concrete
// catalog permissions it covers, in catalog order.
func Expand(perms []string) []Permission {
	out := make([]Permission, 0, len(allPermissions))
	for _, known := range allPermissions {
		if SetHas(perms, known) {
			out = append(out, known)
		}
	}
	return out
}
