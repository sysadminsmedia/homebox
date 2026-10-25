package authzrules

import (
	"context"

	entgo "entgo.io/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/accessgrant"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/apikey"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authtokens"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entityfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/groupinvitationtoken"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/intercept"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/notifier"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/passwordresettokens"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/permissiongroup"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/templatefield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
)

// The filter interceptors below scope every read to what the viewer may see.
// System contexts skip filtering; a missing viewer fails the query (the
// query policy also denies it). They run on edge traversals too, so e.g.
// group.QueryEntities() is filtered the same way as a direct entity query.

// FilterEntity restricts entities to readable rows (tenant read permission
// or row-level grant).
func FilterEntity() entgo.Interceptor {
	return intercept.TraverseEntity(func(ctx context.Context, q *ent.EntityQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(EntityReadable(v))
		return nil
	})
}

// FilterAttachment: readable via the parent entity.
func FilterAttachment() entgo.Interceptor {
	return intercept.TraverseAttachment(func(ctx context.Context, q *ent.AttachmentQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(attachment.HasEntityWith(EntityReadable(v)))
		return nil
	})
}

// FilterMaintenanceEntry: readable via the parent entity.
func FilterMaintenanceEntry() entgo.Interceptor {
	return intercept.TraverseMaintenanceEntry(func(ctx context.Context, q *ent.MaintenanceEntryQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(maintenanceentry.HasEntityWith(EntityReadable(v)))
		return nil
	})
}

// FilterEntityField: readable via the parent entity.
func FilterEntityField() entgo.Interceptor {
	return intercept.TraverseEntityField(func(ctx context.Context, q *ent.EntityFieldQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(entityfield.HasEntityWith(EntityReadable(v)))
		return nil
	})
}

// FilterTag: tenant reference data, readable by any member.
func FilterTag() entgo.Interceptor {
	return intercept.TraverseTag(func(ctx context.Context, q *ent.TagQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(tag.HasGroupWith(group.ID(v.TenantID)))
		return nil
	})
}

// FilterEntityType: tenant reference data, readable by any member.
func FilterEntityType() entgo.Interceptor {
	return intercept.TraverseEntityType(func(ctx context.Context, q *ent.EntityTypeQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(entitytype.HasGroupWith(group.ID(v.TenantID)))
		return nil
	})
}

// FilterEntityTemplate: tenant reference data, readable by any member.
func FilterEntityTemplate() entgo.Interceptor {
	return intercept.TraverseEntityTemplate(func(ctx context.Context, q *ent.EntityTemplateQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(entitytemplate.HasGroupWith(group.ID(v.TenantID)))
		return nil
	})
}

// FilterTemplateField: readable via the parent template's tenant.
func FilterTemplateField() entgo.Interceptor {
	return intercept.TraverseTemplateField(func(ctx context.Context, q *ent.TemplateFieldQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(templatefield.HasEntityTemplateWith(entitytemplate.HasGroupWith(group.ID(v.TenantID))))
		return nil
	})
}

// FilterExport: tenant rows, readable by any member.
func FilterExport() entgo.Interceptor {
	return intercept.TraverseExport(func(ctx context.Context, q *ent.ExportQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(export.GroupID(v.TenantID))
		return nil
	})
}

// FilterNotifier: notifiers are personal — own rows in the active tenant.
func FilterNotifier() entgo.Interceptor {
	return intercept.TraverseNotifier(func(ctx context.Context, q *ent.NotifierQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(notifier.UserID(v.UserID), notifier.GroupID(v.TenantID))
		return nil
	})
}

// FilterGroup: tenants the viewer is a member of (GET /groups/all lists
// across tenants, so membership — not the active tenant — is the boundary).
func FilterGroup() entgo.Interceptor {
	return intercept.TraverseGroup(func(ctx context.Context, q *ent.GroupQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(group.HasUsersWith(user.ID(v.UserID)))
		return nil
	})
}

// FilterUserGroup: memberships of the active tenant, plus the viewer's own
// memberships in other tenants.
func FilterUserGroup() entgo.Interceptor {
	return intercept.TraverseUserGroup(func(ctx context.Context, q *ent.UserGroupQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(usergroup.Or(
			usergroup.GroupID(v.TenantID),
			usergroup.UserID(v.UserID),
		))
		return nil
	})
}

// FilterPermissionGroup: tenant rows, readable by any member (the UI needs
// them to display effective permissions and membership).
func FilterPermissionGroup() entgo.Interceptor {
	return intercept.TraversePermissionGroup(func(ctx context.Context, q *ent.PermissionGroupQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(permissiongroup.GroupID(v.TenantID))
		return nil
	})
}

// FilterAccessGrant: tenant rows, readable by any member.
func FilterAccessGrant() entgo.Interceptor {
	return intercept.TraverseAccessGrant(func(ctx context.Context, q *ent.AccessGrantQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(accessgrant.GroupID(v.TenantID))
		return nil
	})
}

// FilterGroupInvitationToken: tenant rows.
func FilterGroupInvitationToken() entgo.Interceptor {
	return intercept.TraverseGroupInvitationToken(func(ctx context.Context, q *ent.GroupInvitationTokenQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(groupinvitationtoken.HasGroupWith(group.ID(v.TenantID)))
		return nil
	})
}

// FilterUser: the viewer themselves, plus members of the active tenant
// (needed for member lists and grant target names).
func FilterUser() entgo.Interceptor {
	return intercept.TraverseUser(func(ctx context.Context, q *ent.UserQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(user.Or(
			user.ID(v.UserID),
			user.HasGroupsWith(group.ID(v.TenantID)),
		))
		return nil
	})
}

// FilterAPIKey: own keys only.
func FilterAPIKey() entgo.Interceptor {
	return intercept.TraverseAPIKey(func(ctx context.Context, q *ent.APIKeyQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(apikey.UserID(v.UserID))
		return nil
	})
}

// FilterAuthTokens: own sessions only.
func FilterAuthTokens() entgo.Interceptor {
	return intercept.TraverseAuthTokens(func(ctx context.Context, q *ent.AuthTokensQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(authtokens.HasUserWith(user.ID(v.UserID)))
		return nil
	})
}

// FilterAuthRoles: roles of the viewer's own sessions.
func FilterAuthRoles() entgo.Interceptor {
	return intercept.TraverseAuthRoles(func(ctx context.Context, q *ent.AuthRolesQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(authroles.HasTokenWith(authtokens.HasUserWith(user.ID(v.UserID))))
		return nil
	})
}

// FilterPasswordResetTokens: own rows only (flows are system-context, this
// is a defensive floor).
func FilterPasswordResetTokens() entgo.Interceptor {
	return intercept.TraversePasswordResetTokens(func(ctx context.Context, q *ent.PasswordResetTokensQuery) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		q.Where(passwordresettokens.HasUserWith(user.ID(v.UserID)))
		return nil
	})
}
