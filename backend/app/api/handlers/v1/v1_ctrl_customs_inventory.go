package v1

import (
	"errors"
	"io"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/rs/zerolog/log"
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
		return streamDownload(w, r, file.Write)
	})
}

// streamDownload writes a file that is produced while it is sent. An error before
// the first byte still becomes an error response. After that the status and part
// of the file are already out, so the error is logged and the connection aborted:
// the browser then reports a failed download instead of saving a cut-off file.
func streamDownload(w http.ResponseWriter, r *http.Request, write func(io.Writer) error) error {
	cw := &countingWriter{w: w}
	err := write(cw)
	if err == nil || cw.n == 0 {
		return err
	}
	log.Err(err).Ctx(r.Context()).Int64("bytes_sent", cw.n).Msg("download failed partway; aborting the response")
	panic(http.ErrAbortHandler)
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}
