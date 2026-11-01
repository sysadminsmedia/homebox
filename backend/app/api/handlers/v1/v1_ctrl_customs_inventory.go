package v1

import (
	"errors"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
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
		return customsinventory.Preview{}, validate.NewRequestError(errors.New("not implemented"), http.StatusNotImplemented)
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
		if _, err := adapters.DecodeBody[customsinventory.Request](r); err != nil {
			return err
		}
		return validate.NewRequestError(errors.New("not implemented"), http.StatusNotImplemented)
	})
}
