package v1

import (
	"net/http"
	"time"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

type (
	GroupInvitationCreate struct {
		Uses      int       `json:"uses"      validate:"required,min=1,max=100"`
		ExpiresAt time.Time `json:"expiresAt"`
	}

	GroupInvitation struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
		Uses      int       `json:"uses"`
	}
)

// HandleGroupGet godoc
//
//	@Summary	Get Group
//	@Tags		Group
//	@Produce	json
//	@Success	200	{object}	repo.Group
//	@Router		/v1/groups [Get]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupGet() errchain.HandlerFunc {
	fn := func(r *http.Request) (repo.Group, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.GroupByID(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupUpdate godoc
//
//	@Summary	Update Group
//	@Tags		Group
//	@Produce	json
//	@Param		payload	body		repo.GroupUpdate	true	"User Data"
//	@Success	200		{object}	repo.Group
//	@Router		/v1/groups [Put]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.GroupUpdate) (repo.Group, error) {
		auth := services.NewContext(r.Context())

		ok := ctrl.svc.Currencies.IsSupported(body.Currency)
		if !ok {
			return repo.Group{}, validate.NewFieldErrors(
				validate.NewFieldError("currency", "currency '"+body.Currency+"' is not supported"),
			)
		}

		return ctrl.svc.Group.UpdateGroup(auth, body)
	}

	return adapters.Action(fn, http.StatusOK)
}

// HandleGroupInvitationsCreate godoc
//
//	@Summary	Create Group Invitation
//	@Tags		Group
//	@Produce	json
//	@Param		payload	body		GroupInvitationCreate	true	"User Data"
//	@Success	200		{object}	GroupInvitation
//	@Router		/v1/groups/invitations [Post]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupInvitationsCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body GroupInvitationCreate) (GroupInvitation, error) {
		if body.ExpiresAt.IsZero() {
			body.ExpiresAt = time.Now().Add(time.Hour * 24)
		}

		auth := services.NewContext(r.Context())

		token, err := ctrl.svc.Group.NewInvitation(auth, body.Uses, body.ExpiresAt)

		return GroupInvitation{
			Token:     token,
			ExpiresAt: body.ExpiresAt,
			Uses:      body.Uses,
		}, err
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleGroupsGetAll godoc
//
//	@Summary	Get All Groups
//	@Tags		Group
//	@Produce	json
//	@Success	200	{object}	[]repo.Group
//	@Router		/v1/groups [Get]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.Group, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.GetAllGroups(auth)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupCreate godoc
//
//	@Summary	Create Group
//	@Tags		Group
//	@Produce	json
//	@Param		name	body		string	true	"Group Name"
//	@Success	201		{object}	repo.Group
//	@Router		/v1/groups/{id} [Post]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupCreate() errchain.HandlerFunc {
	type CreateRequest struct {
		Name string `json:"name" validate:"required"`
	}

	fn := func(r *http.Request, body CreateRequest) (repo.Group, error) {
		auth := services.NewContext(r.Context())
		return ctrl.svc.Group.CreateGroup(auth, body.Name)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleGroupDelete godoc
//
//	@Summary	Delete Group
//	@Tags		Group
//	@Produce	json
//	@Success	204
//	@Router		/v1/groups/{id} [Delete]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupDelete() errchain.HandlerFunc {
	fn := func(r *http.Request) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.svc.Group.DeleteGroup(auth)
		return nil, err
	}

	return adapters.Command(fn, http.StatusNoContent)
}

// HandleGroupInvitationsGetAll godoc
//
//	@Summary	Get All Group Invitations
//	@Tags		Group
//	@Produce	json
//	@Success	200	{object}	[]repo.GroupInvitation
//	@Router		/v1/groups/invitations [Get]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGroupInvitationsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.GroupInvitation, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.InvitationGetAll(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}
