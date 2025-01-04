package v1

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

type Locales struct {
	Locales []string `json:"locales"`
}

// HandleLocalesGetAll godoc
//
//	@Summary  Get All Locales
//	@Tags     Locales
//	@Produce  json
//	@Success  200 {object} []Locales
//	@Router   /v1/locales [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleLocalesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]Locales, error) {
		// TODO: get a list of locales from files
		return []Locales{}, nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleLocalesGet godoc
//
//	@Summary  Get Locale
//	@Tags     Locales
//	@Produce  json
//	@Param    id  path     string true "Locale ID"
//	@Success  200 {object} interface{}
//	@Router   /v1/locales/{id} [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleLocalesGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (interface{}, error) {
		// TODO: get the current locale
		return interface{}, nil
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}