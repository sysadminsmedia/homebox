package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"go.opentelemetry.io/otel/attribute"
)

// HandleFoundEntityGet godoc
//
//	@Summary	Get public found entity details
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path		string	true	"Entity ID"
//	@Success	200	{object}	repo.FoundEntityOut
//	@Failure	404	{object}	validate.ErrorResponse
//	@Failure	429	{object}	validate.ErrorResponse
//	@Router		/v1/found/entities/{id} [GET]
func (ctrl *V1Controller) HandleFoundEntityGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.FoundEntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleFoundEntityGet",
			attribute.String("entity.id", ID.String()))
		defer span.End()

		out, err := ctrl.repo.Entities.GetFoundEntity(spanCtx, ID)
		if err != nil {
			recordCtrlSpanError(span, err)
		}
		return out, err
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}
