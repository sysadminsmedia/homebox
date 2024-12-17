package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"

	"github.com/rs/zerolog/log"
)

// HandleAssetGet godocs
//
//	@Summary	Get Item by Asset ID
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Asset ID"
//	@Success	200	{object}	repo.PaginationResult[repo.ItemSummary]{}
//	@Router		/v1/assets/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleAssetGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		assetIDParam := chi.URLParam(r, "id")
		assetIDParam = strings.ReplaceAll(assetIDParam, "-", "") // Remove dashes
		// Convert the asset ID to an int64
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return err
		}
		pageParam := r.URL.Query().Get("page")
		var page int64 = -1
		if pageParam != "" {
			page, err = strconv.ParseInt(pageParam, 10, 32)
			if err != nil {
				return server.JSON(w, http.StatusBadRequest, "Invalid page number")
			}
		}

		pageSizeParam := r.URL.Query().Get("pageSize")
		var pageSize int64 = -1
		if pageSizeParam != "" {
			pageSize, err = strconv.ParseInt(pageSizeParam, 10, 32)
			if err != nil {
				return server.JSON(w, http.StatusBadRequest, "Invalid page size")
			}
		}

		items, err := ctrl.repo.Items.QueryByAssetID(r.Context(), ctx.GID, repo.AssetID(assetID), int(page), int(pageSize))
		if err != nil {
			log.Err(err).Msg("failed to get item")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusOK, items)
	}
}
