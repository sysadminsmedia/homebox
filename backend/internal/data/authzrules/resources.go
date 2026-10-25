package authzrules

import (
	"context"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entityfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/notifier"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/templatefield"
)

// --- Entity-owned child resources -----------------------------------------
// Attachments, maintenance entries, and fields inherit the parent entity's
// access: tenant-wide entity:update, or a row grant covering the action.

// AttachmentMutationRule authorizes attachment writes via the parent entity.
func AttachmentMutationRule() privacy.MutationRule {
	return privacy.AttachmentMutationRuleFunc(func(ctx context.Context, m *ent.AttachmentMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		writable := EntityAttachmentsWritable(v)
		if !m.Op().Is(ent.OpCreate) {
			m.Where(attachment.HasEntityWith(writable))
		}
		return checkChildParent(ctx, m, writable)
	})
}

// MaintenanceEntryMutationRule authorizes maintenance writes via the parent.
func MaintenanceEntryMutationRule() privacy.MutationRule {
	return privacy.MaintenanceEntryMutationRuleFunc(func(ctx context.Context, m *ent.MaintenanceEntryMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		writable := EntityChildrenWritable(v)
		if !m.Op().Is(ent.OpCreate) {
			m.Where(maintenanceentry.HasEntityWith(writable))
		}
		return checkChildParent(ctx, m, writable)
	})
}

// EntityFieldMutationRule authorizes custom-field writes via the parent.
func EntityFieldMutationRule() privacy.MutationRule {
	return privacy.EntityFieldMutationRuleFunc(func(ctx context.Context, m *ent.EntityFieldMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		writable := EntityChildrenWritable(v)
		if !m.Op().Is(ent.OpCreate) {
			m.Where(entityfield.HasEntityWith(writable))
		}
		return checkChildParent(ctx, m, writable)
	})
}

// --- Tenant reference data --------------------------------------------------
// Tags, entity types, and templates: any member reads, holders of the
// matching manage permission write. Mutations are pinned to the tenant.

// TagMutationRule requires tag:manage and pins writes to the tenant.
func TagMutationRule() privacy.MutationRule {
	return privacy.TagMutationRuleFunc(func(ctx context.Context, m *ent.TagMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermTagManage) {
			return privacy.Denyf("authz: missing %s", authz.PermTagManage)
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(tag.HasGroupWith(group.ID(v.TenantID)))
		return privacy.Allow
	})
}

// EntityTypeMutationRule requires entitytype:manage, pinned to the tenant.
func EntityTypeMutationRule() privacy.MutationRule {
	return privacy.EntityTypeMutationRuleFunc(func(ctx context.Context, m *ent.EntityTypeMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermEntityTypeManage) {
			return privacy.Denyf("authz: missing %s", authz.PermEntityTypeManage)
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(entitytype.HasGroupWith(group.ID(v.TenantID)))
		return privacy.Allow
	})
}

// EntityTemplateMutationRule requires template:manage, pinned to the tenant.
func EntityTemplateMutationRule() privacy.MutationRule {
	return privacy.EntityTemplateMutationRuleFunc(func(ctx context.Context, m *ent.EntityTemplateMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermTemplateManage) {
			return privacy.Denyf("authz: missing %s", authz.PermTemplateManage)
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(entitytemplate.HasGroupWith(group.ID(v.TenantID)))
		return privacy.Allow
	})
}

// TemplateFieldMutationRule requires template:manage; rows are pinned to
// fields whose template belongs to the tenant.
func TemplateFieldMutationRule() privacy.MutationRule {
	return privacy.TemplateFieldMutationRuleFunc(func(ctx context.Context, m *ent.TemplateFieldMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermTemplateManage) {
			return privacy.Denyf("authz: missing %s", authz.PermTemplateManage)
		}
		if m.Op().Is(ent.OpCreate) {
			// The parent template must be in the viewer's tenant.
			if tid, ok := m.EntityTemplateID(); ok {
				exists, err := m.Client().EntityTemplate.Query().
					Where(entitytemplate.ID(tid), entitytemplate.HasGroupWith(group.ID(v.TenantID))).
					Exist(authz.NewSystemContext(ctx))
				if err != nil {
					return privacy.Denyf("authz: checking template: %v", err)
				}
				if !exists {
					return privacy.Denyf("authz: template not in tenant")
				}
			}
			return privacy.Allow
		}
		m.Where(templatefield.HasEntityTemplateWith(entitytemplate.HasGroupWith(group.ID(v.TenantID))))
		return privacy.Allow
	})
}

// ExportMutationRule requires data:export (or data:import for import-kind
// job rows), pinned to the tenant.
func ExportMutationRule() privacy.MutationRule {
	return privacy.ExportMutationRuleFunc(func(ctx context.Context, m *ent.ExportMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		required := authz.PermDataExport
		if kind, ok := m.Kind(); ok && kind == export.KindImport {
			required = authz.PermDataImport
		}
		if !v.Has(required) {
			return privacy.Denyf("authz: missing %s", required)
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(export.GroupID(v.TenantID))
		return privacy.Allow
	})
}

// NotifierMutationRule requires notifier:manage; notifiers are additionally
// personal, so writes are pinned to the viewer's own rows.
func NotifierMutationRule() privacy.MutationRule {
	return privacy.NotifierMutationRuleFunc(func(ctx context.Context, m *ent.NotifierMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}
		if !v.Has(authz.PermNotifierManage) {
			return privacy.Denyf("authz: missing %s", authz.PermNotifierManage)
		}
		if uid, ok := m.UserID(); ok && uid != v.UserID {
			return privacy.Denyf("authz: notifier belongs to another user")
		}
		if m.Op().Is(ent.OpCreate) {
			return allowIfTenantCreate(m, v)
		}
		m.Where(notifier.UserID(v.UserID), notifier.GroupID(v.TenantID))
		return privacy.Allow
	})
}
