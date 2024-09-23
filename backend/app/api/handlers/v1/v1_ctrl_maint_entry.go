package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenanceLogGet godoc
//
//	@Summary  Get Maintenance Log
//	@Tags     Item Maintenance
//	@Produce  json
//	@Success  200       {object} repo.MaintenanceLog
//	@Router   /v1/items/{id}/maintenance [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceLogGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, q repo.MaintenanceLogQuery) (repo.MaintenanceLog, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.GetLog(auth, auth.GID, ID, q)
	}

	return adapters.QueryID("id", fn, http.StatusOK)
}

// HandleMaintenanceEntryCreate godoc
//
//	@Summary  Create Maintenance Entry
//	@Tags     Item Maintenance
//	@Produce  json
//	@Param    payload body     repo.MaintenanceEntryCreate true "Entry Data"
//	@Success  201     {object} repo.MaintenanceEntry
//	@Router   /v1/items/{id}/maintenance [POST]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, itemID uuid.UUID, body repo.MaintenanceEntryCreate) (repo.MaintenanceEntry, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.Create(auth, itemID, body)
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}
