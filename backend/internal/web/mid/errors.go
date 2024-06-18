package mid

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func Errors(log zerolog.Logger) errchain.ErrorHandler {
	return func(h errchain.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := h.ServeHTTP(w, r)
			if err != nil {
				var resp ErrorResponse
				var code int

				traceID := r.Context().Value(middleware.RequestIDKey).(string)
				log.Err(err).
					Stack().
					Str("req_id", traceID).
					Msg("ERROR occurred")

				switch {
				case validate.IsUnauthorizedError(err):
					code = http.StatusUnauthorized
					resp = ErrorResponse{
						Error: "unauthorized",
					}
				case validate.IsInvalidRouteKeyError(err):
					code = http.StatusBadRequest
					resp = ErrorResponse{
						Error: err.Error(),
					}
				case validate.IsFieldError(err):
					code = http.StatusUnprocessableEntity

					fieldErrors := err.(validate.FieldErrors) // nolint
					resp.Error = "Validation Error"
					resp.Fields = map[string]string{}

					for _, fieldError := range fieldErrors {
						resp.Fields[fieldError.Field] = fieldError.Error
					}
				case validate.IsRequestError(err):
					requestError := err.(*validate.RequestError) // nolint
					resp.Error = requestError.Error()

					if requestError.Status == 0 {
						code = http.StatusBadRequest
					} else {
						code = requestError.Status
					}
				case ent.IsNotFound(err):
					resp.Error = "Not Found"
					code = http.StatusNotFound
				default:
					resp.Error = "Unknown Error"
					code = http.StatusInternalServerError
				}

				if err := server.JSON(w, code, resp); err != nil {
					log.Err(err).Msg("failed to write response")
				}
			}
		})
	}
}
