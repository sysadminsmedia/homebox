package v1

import (
	"errors"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/customsinventory"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// customsInventoryErrors turns the package's request errors into 400s. They
// can come from the body decoder (Request.Validate) or from the service.
func customsInventoryErrors(h errchain.HandlerFunc) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := h(w, r)
		if errors.Is(err, customsinventory.ErrInvalidRequest) || errors.Is(err, customsinventory.ErrScopeTooLarge) {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		return err
	}
}

// HandleCustomsInventoryPreview godoc
//
//	@Summary	Preview Customs Inventory List
//	@Tags		Reporting
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		customsinventory.Request	true	"Export request"
//	@Success	200		{object}	customsinventory.Preview
//	@Router		/v1/reporting/customs-inventory/preview [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleCustomsInventoryPreview() errchain.HandlerFunc {
	fn := func(r *http.Request, req customsinventory.Request) (customsinventory.Preview, error) {
		return ctrl.svc.CustomsInventory.Preview(services.NewContext(r.Context()), req)
	}

	return customsInventoryErrors(adapters.Action(fn, http.StatusOK))
}

// HandleCustomsInventoryExport godoc
//
//	@Summary	Export Customs Inventory List
//	@Tags		Reporting
//	@Accept		json
//	@Produce	application/zip,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param		payload	body	customsinventory.Request	true	"Export request"
//	@Success	200		{file}	binary
//	@Router		/v1/reporting/customs-inventory/export [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleCustomsInventoryExport() errchain.HandlerFunc {
	return customsInventoryErrors(func(w http.ResponseWriter, r *http.Request) error {
		req, err := adapters.DecodeBody[customsinventory.Request](r)
		if err != nil {
			return err
		}

		file, err := ctrl.svc.CustomsInventory.Export(services.NewContext(r.Context()), req)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", file.ContentType)
		w.Header().Set("Content-Disposition", `attachment; filename="`+file.Name+`"`)
		return file.Write(w)
	})
}
