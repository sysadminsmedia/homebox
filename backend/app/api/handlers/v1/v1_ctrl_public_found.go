package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

func publicFoundRequestError(err error) error {
	if ent.IsNotFound(err) {
		return validate.NewRequestError(errors.New("not found"), http.StatusNotFound)
	}
	return validate.NewRequestError(err, http.StatusInternalServerError)
}

// HandlePublicFoundItemGet godoc
//
//	@Summary	Get public found item contact
//	@Tags		Public
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	repo.PublicFoundEntity
//	@Router		/v1/public/found/item/{id} [GET]
func (ctrl *V1Controller) HandlePublicFoundItemGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ID, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		out, err := ctrl.repo.Entities.GetPublicFoundByID(r.Context(), ID)
		if err != nil {
			return publicFoundRequestError(err)
		}

		return server.JSON(w, http.StatusOK, out)
	}
}

// HandlePublicFoundAssetGet godoc
//
//	@Summary	Get public found asset contact
//	@Tags		Public
//	@Produce	json
//	@Param		id	path		string	true	"Asset ID"
//	@Success	200	{object}	repo.PublicFoundEntity
//	@Router		/v1/public/found/asset/{id} [GET]
func (ctrl *V1Controller) HandlePublicFoundAssetGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		assetIDParam := strings.ReplaceAll(chi.URLParam(r, "id"), "-", "")
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return validate.NewRequestError(errors.New("not found"), http.StatusNotFound)
		}
		if repo.AssetID(assetID).Nil() {
			return validate.NewRequestError(errors.New("not found"), http.StatusNotFound)
		}

		out, err := ctrl.repo.Entities.GetPublicFoundByAssetID(r.Context(), repo.AssetID(assetID))
		if err != nil {
			return publicFoundRequestError(err)
		}

		return server.JSON(w, http.StatusOK, out)
	}
}
