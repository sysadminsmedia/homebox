package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleFoundEntityContact godoc
//
//	@Summary	Get found item contact
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path		string	true	"Entity ID"
//	@Success	200	{object}	repo.FoundEntityContact
//	@Router		/v1/found/entities/{id} [GET]
func (ctrl *V1Controller) HandleFoundEntityContact() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.FoundEntityContact, error) {
		return ctrl.repo.Entities.GetFoundEntityContact(r.Context(), ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleFoundAssetContact godoc
//
//	@Summary	Get found asset contact
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path		string	true	"Asset ID"
//	@Success	200	{object}	repo.FoundEntityContact
//	@Router		/v1/found/assets/{id} [GET]
func (ctrl *V1Controller) HandleFoundAssetContact() errchain.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) error {
		assetID, ok := repo.ParseAssetID(chi.URLParam(r, "id"))
		if !ok || assetID.Nil() {
			return validate.NewRouteKeyError("id")
		}

		contact, err := ctrl.repo.Entities.GetFoundEntityContactByAssetID(r.Context(), assetID)
		if err != nil {
			return err
		}

		return server.JSON(rw, http.StatusOK, contact)
	}
}
