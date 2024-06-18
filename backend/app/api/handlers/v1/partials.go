package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// routeID extracts the ID from the request URL. If the ID is not in a valid
// format, an error is returned. If a error is returned, it can be directly returned
// from the handler. the validate.ErrInvalidID error is known by the error middleware
// and will be handled accordingly.
//
// Example: /api/v1/ac614db5-d8b8-4659-9b14-6e913a6eb18a -> uuid.UUID{ac614db5-d8b8-4659-9b14-6e913a6eb18a}
func (ctrl *V1Controller) routeID(r *http.Request) (uuid.UUID, error) {
	return ctrl.routeUUID(r, "id")
}

func (ctrl *V1Controller) routeUUID(r *http.Request, key string) (uuid.UUID, error) {
	ID, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.Nil, validate.NewRouteKeyError(key)
	}
	return ID, nil
}
