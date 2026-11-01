package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/customsinventory"
	"github.com/sysadminsmedia/homebox/backend/internal/web/mid"
)

// Invalid requests are rejected while decoding, before the service runs, so a
// controller without services is enough here.
func TestCustomsInventoryHandlers_RejectInvalidRequests(t *testing.T) {
	ctrl := &V1Controller{}
	chain := errchain.New(mid.Errors(zerolog.Nop()))
	handlers := map[string]http.HandlerFunc{
		"preview": chain.ToHandlerFunc(ctrl.HandleCustomsInventoryPreview()),
		"export":  chain.ToHandlerFunc(ctrl.HandleCustomsInventoryExport()),
	}
	const id = `"00000000-0000-0000-0000-000000000001"`
	cases := []struct {
		name string
		body string
		want int
	}{
		{"missing reference date", `{"scope":{"itemIds":[` + id + `]},"options":{"groupBy":"location"}}`, http.StatusBadRequest},
		{"empty scope", `{"scope":{"itemIds":[]},"options":{"groupBy":"location","referenceDate":"2026-11-15"}}`, http.StatusBadRequest},
		{"unknown grouping", `{"scope":{"itemIds":[` + id + `]},"options":{"groupBy":"room","referenceDate":"2026-11-15"}}`, http.StatusBadRequest},
		{"title too long", `{"scope":{"itemIds":[` + id + `]},"options":{"groupBy":"location","referenceDate":"2026-11-15"},"header":{"title":"` + strings.Repeat("x", 201) + `"}}`, http.StatusUnprocessableEntity},
	}
	for name, h := range handlers {
		for _, tc := range cases {
			rec := httptest.NewRecorder()
			h(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body)))
			assert.Equal(t, tc.want, rec.Code, "%s: %s: %s", name, tc.name, rec.Body.String())
		}
	}
}

func TestCustomsInventoryErrors_ScopeTooLargeIsBadRequest(t *testing.T) {
	chain := errchain.New(mid.Errors(zerolog.Nop()))
	h := chain.ToHandlerFunc(customsInventoryErrors(func(http.ResponseWriter, *http.Request) error {
		return fmt.Errorf("%w: 5001 entities, at most 5000", customsinventory.ErrScopeTooLarge)
	}))
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "at most 5000")
}

// Once part of a download is sent, an error can no longer become an error
// response, so the connection is aborted instead of ending a cut-off file.
func TestStreamDownload(t *testing.T) {
	boom := errors.New("receipt read failed")
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	t.Run("error before the first byte is returned", func(t *testing.T) {
		err := streamDownload(httptest.NewRecorder(), req, func(io.Writer) error { return boom })
		assert.ErrorIs(t, err, boom)
	})

	t.Run("error after the first byte aborts the response", func(t *testing.T) {
		assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
			_ = streamDownload(httptest.NewRecorder(), req, func(w io.Writer) error {
				_, _ = w.Write([]byte("PK partial zip"))
				return boom
			})
		})
	})

	t.Run("success writes the file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		err := streamDownload(rec, req, func(w io.Writer) error {
			_, err := w.Write([]byte("whole file"))
			return err
		})
		assert.NoError(t, err)
		assert.Equal(t, "whole file", rec.Body.String())
	})
}
