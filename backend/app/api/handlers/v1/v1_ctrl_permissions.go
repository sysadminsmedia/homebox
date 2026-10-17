package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

type (
	PermissionGroupSetMembers struct {
		UserIds []uuid.UUID `json:"userIds"`
	}

	AccessGrantUpdate struct {
		Actions []string `json:"actions" validate:"required,min=1"`
	}
)

// HandlePermissionsCatalog godoc
//
//	@Summary	Get Permission Catalog
//	@Tags		Permissions
//	@Produce	json
//	@Success	200	{object}	[]services.PermissionDefinition
//	@Router		/v1/permissions/catalog [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionsCatalog() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]services.PermissionDefinition, error) {
		return ctrl.svc.Permissions.Catalog(), nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupPermissionsSelf godoc
//
//	@Summary	Get Caller's Effective Permissions
//	@Description	Returns the caller's effective permission set for the active tenant.
//	@Tags		Permissions
//	@Produce	json
//	@Success	200	{object}	services.EffectivePermissionsOut
//	@Router		/v1/groups/permissions/self [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupPermissionsSelf() errchain.HandlerFunc {
	fn := func(r *http.Request) (services.EffectivePermissionsOut, error) {
		return ctrl.svc.Permissions.Self(services.NewContext(r.Context()))
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandlePermissionGroupsGetAll godoc
//
//	@Summary	Get All Permission Groups
//	@Tags		Permissions
//	@Produce	json
//	@Success	200	{object}	[]repo.PermissionGroupOut
//	@Router		/v1/groups/permission-groups [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.PermissionGroupOut, error) {
		return ctrl.svc.Permissions.GetPermissionGroups(services.NewContext(r.Context()))
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandlePermissionGroupCreate godoc
//
//	@Summary	Create Permission Group
//	@Tags		Permissions
//	@Produce	json
//	@Param		payload	body		repo.PermissionGroupCreate	true	"Permission Group Data"
//	@Success	201		{object}	repo.PermissionGroupOut
//	@Router		/v1/groups/permission-groups [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.PermissionGroupCreate) (repo.PermissionGroupOut, error) {
		return ctrl.svc.Permissions.CreatePermissionGroup(services.NewContext(r.Context()), body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandlePermissionGroupGet godoc
//
//	@Summary	Get Permission Group
//	@Tags		Permissions
//	@Produce	json
//	@Param		id	path		string	true	"Permission Group ID"
//	@Success	200	{object}	repo.PermissionGroupOut
//	@Router		/v1/groups/permission-groups/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupGet() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.PermissionGroupOut, error) {
		return ctrl.svc.Permissions.GetPermissionGroup(services.NewContext(r.Context()), id)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandlePermissionGroupUpdate godoc
//
//	@Summary	Update Permission Group
//	@Tags		Permissions
//	@Produce	json
//	@Param		id		path		string						true	"Permission Group ID"
//	@Param		payload	body		repo.PermissionGroupUpdate	true	"Permission Group Data"
//	@Success	200		{object}	repo.PermissionGroupOut
//	@Router		/v1/groups/permission-groups/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.PermissionGroupUpdate) (repo.PermissionGroupOut, error) {
		return ctrl.svc.Permissions.UpdatePermissionGroup(services.NewContext(r.Context()), id, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandlePermissionGroupDelete godoc
//
//	@Summary	Delete Permission Group
//	@Tags		Permissions
//	@Produce	json
//	@Param		id	path	string	true	"Permission Group ID"
//	@Success	204
//	@Router		/v1/groups/permission-groups/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (any, error) {
		err := ctrl.svc.Permissions.DeletePermissionGroup(services.NewContext(r.Context()), id)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandlePermissionGroupMembersSet godoc
//
//	@Summary	Set Permission Group Members
//	@Description	Replaces the member list of a permission group. Every member must belong to the tenant.
//	@Tags		Permissions
//	@Produce	json
//	@Param		id		path		string						true	"Permission Group ID"
//	@Param		payload	body		PermissionGroupSetMembers	true	"Member List"
//	@Success	200		{object}	repo.PermissionGroupOut
//	@Router		/v1/groups/permission-groups/{id}/members [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePermissionGroupMembersSet() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body PermissionGroupSetMembers) (repo.PermissionGroupOut, error) {
		return ctrl.svc.Permissions.SetPermissionGroupMembers(services.NewContext(r.Context()), id, body.UserIds)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleGroupMemberPermissionsGet godoc
//
//	@Summary	Get Member Permissions
//	@Description	Returns a member's direct permissions, permission groups, and effective set.
//	@Tags		Permissions
//	@Produce	json
//	@Param		user_id	path		string	true	"User ID"
//	@Success	200		{object}	repo.MemberPermissions
//	@Router		/v1/groups/members/{user_id}/permissions [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupMemberPermissionsGet() errchain.HandlerFunc {
	fn := func(r *http.Request, userID uuid.UUID) (repo.MemberPermissions, error) {
		return ctrl.svc.Permissions.GetMemberPermissions(services.NewContext(r.Context()), userID)
	}

	return adapters.CommandID("user_id", fn, http.StatusOK)
}

// HandleGroupMemberPermissionsSet godoc
//
//	@Summary	Set Member Direct Permissions
//	@Tags		Permissions
//	@Produce	json
//	@Param		user_id	path		string					true	"User ID"
//	@Param		payload	body		MemberPermissionsSet	true	"Permissions"
//	@Success	200		{object}	repo.MemberPermissions
//	@Router		/v1/groups/members/{user_id}/permissions [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupMemberPermissionsSet() errchain.HandlerFunc {
	fn := func(r *http.Request, userID uuid.UUID, body MemberPermissionsSet) (repo.MemberPermissions, error) {
		return ctrl.svc.Permissions.SetMemberPermissions(services.NewContext(r.Context()), userID, body.Permissions)
	}

	return adapters.ActionID("user_id", fn, http.StatusOK)
}

// MemberPermissionsSet is the payload for setting a member's direct permissions.
type MemberPermissionsSet struct {
	Permissions []string `json:"permissions"`
}

// HandleEntityGrantsGetAll godoc
//
//	@Summary	Get Entity Access Grants
//	@Description	Lists the row-level access grants on one entity.
//	@Tags		Permissions
//	@Produce	json
//	@Param		id	path		string	true	"Entity ID"
//	@Success	200	{object}	[]repo.AccessGrantOut
//	@Router		/v1/entities/{id}/permissions [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityGrantsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) ([]repo.AccessGrantOut, error) {
		return ctrl.svc.Permissions.GetEntityGrants(services.NewContext(r.Context()), id)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityGrantCreate godoc
//
//	@Summary	Create Entity Access Grant
//	@Description	Grants a user or permission group row-level access to one entity.
//	@Tags		Permissions
//	@Produce	json
//	@Param		id		path		string					true	"Entity ID"
//	@Param		payload	body		repo.AccessGrantCreate	true	"Grant Data"
//	@Success	201		{object}	repo.AccessGrantOut
//	@Router		/v1/entities/{id}/permissions [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityGrantCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body repo.AccessGrantCreate) (repo.AccessGrantOut, error) {
		return ctrl.svc.Permissions.CreateEntityGrant(services.NewContext(r.Context()), id, body)
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}

// HandleEntityGrantUpdate godoc
//
//	@Summary	Update Entity Access Grant
//	@Tags		Permissions
//	@Produce	json
//	@Param		id			path		string				true	"Entity ID"
//	@Param		grant_id	path		string				true	"Grant ID"
//	@Param		payload		body		AccessGrantUpdate	true	"Grant Actions"
//	@Success	200			{object}	repo.AccessGrantOut
//	@Router		/v1/entities/{id}/permissions/{grant_id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityGrantUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, body AccessGrantUpdate) (repo.AccessGrantOut, error) {
		grantID, err := ctrl.routeUUID(r, "grant_id")
		if err != nil {
			return repo.AccessGrantOut{}, err
		}
		return ctrl.svc.Permissions.UpdateEntityGrant(services.NewContext(r.Context()), id, grantID, body.Actions)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityGrantDelete godoc
//
//	@Summary	Delete Entity Access Grant
//	@Tags		Permissions
//	@Produce	json
//	@Param		id			path	string	true	"Entity ID"
//	@Param		grant_id	path	string	true	"Grant ID"
//	@Success	204
//	@Router		/v1/entities/{id}/permissions/{grant_id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityGrantDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (any, error) {
		grantID, err := ctrl.routeUUID(r, "grant_id")
		if err != nil {
			return nil, err
		}
		err = ctrl.svc.Permissions.DeleteEntityGrant(services.NewContext(r.Context()), id, grantID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
