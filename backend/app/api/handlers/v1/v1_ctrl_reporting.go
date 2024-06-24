package v1

import (
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"net/http"
)

// HandleBillOfMaterialsExport godoc
//
//	@Summary  Export Bill of Materials
//	@Tags     Reporting
//	@Produce  json
//	@Success 200 {string} string "text/csv"
//	@Router   /v1/reporting/bill-of-materials [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleBillOfMaterialsExport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		actor := services.UseUserCtx(r.Context())

		csv, err := ctrl.svc.Items.ExportBillOfMaterialsCSV(r.Context(), actor.GroupID)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=bill-of-materials.csv")
		_, err = w.Write(csv)
		return err
	}
}
