package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// foundTestRequest builds a request whose chi route context carries the
// {kind} and {id} params, matching how the router invokes the handlers.
func foundTestRequest(method, kind, id string, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, "/api/v1/found/"+kind+"/"+id, nil)
	} else {
		r = httptest.NewRequest(method, "/api/v1/found/"+kind+"/"+id+"/contact", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("kind", kind)
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// The cases below only exercise paths that fail before any repository or
// service access, so a zero-value controller is sufficient. Cases that reach
// the database are covered by the repository and service layer tests.

func TestFoundLookupRejectsBadRefsBeforeQuery(t *testing.T) {
	ctrl := &V1Controller{}

	cases := []struct {
		name string
		kind string
		id   string
	}{
		{name: "unknown kind", kind: "barcode", id: "000-001"},
		{name: "empty kind", kind: "", id: "000-001"},
		{name: "malformed item uuid", kind: "item", id: "not-a-uuid"},
		{name: "empty item id", kind: "item", id: ""},
		{name: "asset id zero", kind: "asset", id: "000-000"},
		{name: "asset id bare zero", kind: "asset", id: "0"},
		// Note: ParseAssetID strips dashes, so "-5" parses as asset 5 and
		// is a legitimate (queryable) reference; negatives cannot occur.
		{name: "asset id non-numeric", kind: "asset", id: "abc"},
		{name: "empty asset id", kind: "asset", id: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := foundTestRequest(http.MethodGet, tc.kind, tc.id, "")
			_, err := ctrl.foundLookup(r)
			assert.ErrorIs(t, err, errFoundNotFound)
		})
	}
}

func TestHandleFoundGetNotFoundIsOpaque(t *testing.T) {
	ctrl := &V1Controller{}
	handler := ctrl.HandleFoundGet()

	w := httptest.NewRecorder()
	err := handler(w, foundTestRequest(http.MethodGet, "item", "not-a-uuid", ""))

	var reqErr *validate.RequestError
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, http.StatusNotFound, reqErr.Status)
	assert.Equal(t, errFoundNotFound.Error(), reqErr.Error())
}

func TestHandleFoundContactValidation(t *testing.T) {
	ctrl := &V1Controller{}
	handler := ctrl.HandleFoundContact()

	badBodies := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{`},
		{name: "missing message", body: `{}`},
		{name: "empty message", body: `{"message": ""}`},
		{name: "whitespace-only message", body: `{"message": "  \n\t "}`},
		{name: "message too long", body: `{"message": "` + strings.Repeat("a", foundContactMaxMessageLen+1) + `"}`},
		{name: "invalid replyTo", body: `{"message": "hi", "replyTo": "not-an-email"}`},
	}

	for _, tc := range badBodies {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := handler(w, foundTestRequest(http.MethodPost, "item", "not-a-uuid", tc.body))

			var reqErr *validate.RequestError
			require.ErrorAs(t, err, &reqErr)
			assert.Equal(t, http.StatusBadRequest, reqErr.Status)
		})
	}
}

func TestHandleFoundContactBoundaryAndAlways204(t *testing.T) {
	// A zero-value FoundService reports MailerReady false, so valid input
	// hits the mailer-not-configured branch of the always-204 zone.
	ctrl := &V1Controller{svc: &services.AllServices{Found: &services.FoundService{}}}
	handler := ctrl.HandleFoundContact()

	okBodies := []struct {
		name string
		body string
	}{
		// Message exactly at the limit passes validation.
		{name: "message at max length", body: `{"message": "` + strings.Repeat("a", foundContactMaxMessageLen) + `"}`},
		// Empty replyTo is optional.
		{name: "empty replyTo", body: `{"message": "hi", "replyTo": ""}`},
		{name: "valid replyTo", body: `{"message": "hi", "replyTo": "finder@example.com"}`},
	}

	// These exercise the always-204 anti-probing zone: valid input on an
	// unready-mailer instance must return 204 with no error,
	// indistinguishable from a successful send.
	for _, tc := range okBodies {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := handler(w, foundTestRequest(http.MethodPost, "item", "not-a-uuid", tc.body))

			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, w.Code)
		})
	}
}
