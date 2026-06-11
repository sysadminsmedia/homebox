package authzrules

import (
	"context"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/accessgrant"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/apikey"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authtokens"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/groupinvitationtoken"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/permissiongroup"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
)

// GroupMutationRule: any authenticated viewer may create a new tenant;
// updating or deleting one requires settings:manage and is pinned to the
// viewer's active tenant.
func GroupMutationRule() privacy.MutationRule {
	return privacy.GroupMutationRuleFunc(func(ctx context.Context, m *ent.GroupMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if m.Op().Is(ent.OpCreate) {
			return privacy.Allow
		}
		if !v.Has(authz.PermSettingsManage) {
			return privacy.Denyf("authz: missing %s", authz.PermSettingsManage)
		}
		m.Where(group.ID(v.TenantID))
		return privacy.Allow
	})
}

// UserGroupMutationRule authorizes membership changes. Creating or deleting a
// membership requires members:manage; changing a membership's permission list
// requires permissions:manage. All writes are pinned to the viewer's tenant.
// Self-leave and registration/invitation flows run under a system context in
// the service layer after their own validation.
func UserGroupMutationRule() privacy.MutationRule {
	return privacy.UserGroupMutationRuleFunc(func(ctx context.Context, m *ent.UserGroupMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}

		// Permission edits are the more privileged operation.
		_, permsSet := m.Permissions()
		_, permsAppended := m.AppendedPermissions()
		if permsSet || permsAppended {
			if !v.Has(authz.PermPermissionsManage) {
				return privacy.Denyf("authz: missing %s", authz.PermPermissionsManage)
			}
		} else if !v.Has(authz.PermMembersManage) {
			return privacy.Denyf("authz: missing %s", authz.PermMembersManage)
		}

		if m.Op().Is(ent.OpCreate) {
			if gid, ok := m.GroupID(); ok && gid != v.TenantID {
				return privacy.Denyf("authz: cross-tenant membership")
			}
			return privacy.Allow
		}
		m.Where(usergroup.GroupID(v.TenantID))
		return privacy.Allow
	})
}

// PermissionGroupMutationRule requires permissions:manage, pinned to the
// tenant. Member-edge changes ride on the same mutation and are covered.
func PermissionGroupMutationRule() privacy.MutationRule {
	return privacy.PermissionGroupMutationRuleFunc(func(ctx context.Context, m *ent.PermissionGroupMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermPermissionsManage) {
			return privacy.Denyf("authz: missing %s", authz.PermPermissionsManage)
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(permissiongroup.GroupID(v.TenantID))
		return privacy.Allow
	})
}

// AccessGrantMutationRule requires permissions:manage. On create the target
// entity must be readable by the viewer and live in their tenant; other ops
// are pinned to the tenant.
func AccessGrantMutationRule() privacy.MutationRule {
	return privacy.AccessGrantMutationRuleFunc(func(ctx context.Context, m *ent.AccessGrantMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermPermissionsManage) {
			return privacy.Denyf("authz: missing %s", authz.PermPermissionsManage)
		}
		if m.Op().Is(ent.OpCreate) {
			if gid, ok := m.GroupID(); ok && gid != v.TenantID {
				return privacy.Denyf("authz: cross-tenant grant")
			}
			if eid, ok := m.EntityID(); ok {
				exists, err := m.Client().Entity.Query().
					Where(entity.ID(eid), entity.HasGroupWith(group.ID(v.TenantID)), EntityReadable(v)).
					Exist(authz.NewSystemContext(ctx))
				if err != nil {
					return privacy.Denyf("authz: checking grant entity: %v", err)
				}
				if !exists {
					return privacy.Denyf("authz: grant entity not accessible")
				}
			}
			return privacy.Allow
		}
		m.Where(accessgrant.GroupID(v.TenantID))
		return privacy.Allow
	})
}

// GroupInvitationTokenMutationRule requires members:manage. To prevent
// privilege escalation, the permissions carried by an invitation must be a
// subset of the inviter's own effective permissions unless the inviter holds
// permissions:manage.
func GroupInvitationTokenMutationRule() privacy.MutationRule {
	return privacy.GroupInvitationTokenMutationRuleFunc(func(ctx context.Context, m *ent.GroupInvitationTokenMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermMembersManage) {
			return privacy.Denyf("authz: missing %s", authz.PermMembersManage)
		}
		// Inspect both the set and appended permission values: an inviter
		// without permissions:manage may not grant, via either path, any
		// permission they do not themselves hold. (UserGroupMutationRule
		// guards direct membership edits the same way.)
		if !v.Has(authz.PermPermissionsManage) {
			set, setOK := m.Permissions()
			appended, appendedOK := m.AppendedPermissions()
			invited := make([]string, 0, len(set)+len(appended))
			if setOK {
				invited = append(invited, set...)
			}
			if appendedOK {
				invited = append(invited, appended...)
			}
			// Expand wildcards so ["*"] is compared by what it covers, not
			// by string equality.
			for _, p := range authz.Expand(invited) {
				if !v.Has(p) {
					return privacy.Denyf("authz: cannot grant %s via invitation", p)
				}
			}
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(groupinvitationtoken.HasGroupWith(group.ID(v.TenantID)))
		return privacy.Allow
	})
}

// UserMutationRule allows users to mutate only their own row. Creation is
// reserved for the registration flow (system context).
func UserMutationRule() privacy.MutationRule {
	return privacy.UserMutationRuleFunc(func(ctx context.Context, m *ent.UserMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if m.Op().Is(ent.OpCreate) {
			return privacy.Denyf("authz: user creation is registration-only")
		}
		m.Where(user.ID(v.UserID))
		return privacy.Allow
	})
}

// APIKeyMutationRule lets users manage only their own API keys.
func APIKeyMutationRule() privacy.MutationRule {
	return privacy.APIKeyMutationRuleFunc(func(ctx context.Context, m *ent.APIKeyMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if m.Op().Is(ent.OpCreate) {
			if uid, ok := m.UserID(); !ok || uid != v.UserID {
				return privacy.Denyf("authz: api key must belong to the viewer")
			}
			return privacy.Allow
		}
		m.Where(apikey.UserID(v.UserID))
		return privacy.Allow
	})
}

// AuthTokensMutationRule lets users delete (log out) their own sessions.
// Creation happens during login under a system context.
func AuthTokensMutationRule() privacy.MutationRule {
	return privacy.AuthTokensMutationRuleFunc(func(ctx context.Context, m *ent.AuthTokensMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
			return privacy.Skip // -> deny: create/update are system-only
		}
		m.Where(authtokens.HasUserWith(user.ID(v.UserID)))
		return privacy.Allow
	})
}
