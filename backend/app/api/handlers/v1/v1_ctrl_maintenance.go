package v1

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// HandleMaintenancesGetAll godoc
//
//	@Summary  Query All Maintenances
//	@Tags     Maintenances
//	@Produce  json
//  @Param    includeCompleted query    bool     false "include completed"
//	@Success  200       {object} repo.MaintenanceEntry
//	@Router   /v1/maintenances [GET]
//	@Security Bearer

func (ctrl *V1Controller) HandleMaintenancesGetAll() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		params := r.URL.Query()

		maintenances, err := ctrl.repo.MaintEntry.GetAllMaintenances(ctx, queryBool(params.Get("includeCompleted")))

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return server.JSON(w, http.StatusOK, []repo.MaintenanceEntry{})
			}
			log.Err(err).Msg("failed to get maintenances")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusOK, maintenances)
	}
}
