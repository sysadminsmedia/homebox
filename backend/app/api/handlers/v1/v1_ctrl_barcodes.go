package v1

import (
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleGenerateQRCode godoc
//
//	@Summary	Search EAN from Barcode
//	@Tags		Items
//	@Produce	json
//	@Param		data	query		string	false	"barcode to be searched"
//	@Success	200		{object}	[]repo.BarcodeProduct
//	@Router		/v1/products/search-from-barcode [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleProductSearchFromBarcode() errchain.HandlerFunc {
	type query struct {
		// 80 characters is the longest non-2D barcode length (GS1-128)
		EAN string `schema:"productEAN" validate:"required,max=80"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		q, err := adapters.DecodeQuery[query](r)
		if err != nil {
			return err
		}

		products, err := ctrl.repo.Barcode.RetrieveProductsFromBarcode(ctrl.config.Barcode, q.EAN)

		if err != nil {
			return server.JSON(w, http.StatusInternalServerError, nil)
		}

		if len(products) != 0 {
			return server.JSON(w, http.StatusOK, products)
		}

		return server.JSON(w, http.StatusNoContent, nil)
	}
}
