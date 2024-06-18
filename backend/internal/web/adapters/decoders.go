package adapters

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/hay-kot/httpkit/server"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

var queryDecoder = schema.NewDecoder()

func DecodeQuery[T any](r *http.Request) (T, error) {
	var v T
	err := queryDecoder.Decode(&v, r.URL.Query())
	if err != nil {
		return v, errors.Wrap(err, "decoding error")
	}

	err = validate.Check(v)
	if err != nil {
		return v, errors.Wrap(err, "validation error")
	}

	return v, nil
}

type Validator interface {
	Validate() error
}

func DecodeBody[T any](r *http.Request) (T, error) {
	var val T

	err := server.Decode(r, &val)
	if err != nil {
		return val, errors.Wrap(err, "body decoding error")
	}

	err = validate.Check(val)
	if err != nil {
		return val, err
	}

	if v, ok := any(val).(Validator); ok {
		err = v.Validate()
		if err != nil {
			return val, errors.Wrap(err, "validation error")
		}
	}

	return val, nil
}

func RouteUUID(r *http.Request, key string) (uuid.UUID, error) {
	ID, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		return uuid.Nil, validate.NewRouteKeyError(key)
	}
	return ID, nil
}
