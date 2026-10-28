package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

const (
	depthQueryParam    = "depth"
	pageSizeQueryParam = "pageSize"
)

func TestLocationDescendantHTTPError(t *testing.T) {
	invalidInputs := []error{
		repo.ErrInvalidLocationQuery,
		fmt.Errorf("%w: location page offset is too large", repo.ErrInvalidLocationQuery),
		repo.ErrLocationRootNotLocation,
	}
	for _, input := range invalidInputs {
		t.Run(input.Error(), func(t *testing.T) {
			err := locationDescendantHTTPError(input)
			var requestError *validate.RequestError
			require.ErrorAs(t, err, &requestError)
			require.Equal(t, http.StatusBadRequest, requestError.Status)
		})
	}

	// Unrelated repository errors must not be presented as client mistakes.
	serverErr := errors.New("database unavailable")
	require.ErrorIs(t, locationDescendantHTTPError(serverErr), serverErr)
	var requestError *validate.RequestError
	require.NotErrorAs(t, locationDescendantHTTPError(serverErr), &requestError)
	require.NoError(t, locationDescendantHTTPError(nil))
}

// All supported depths and parameter boundaries, including invalid input,
// must be covered independently of the HTTP handler's database access.
func TestParseLocationDescendantQuery(t *testing.T) {
	tests := []struct {
		name                  string
		input                 url.Values
		depth, page, pageSize int
		badRequest            bool
	}{
		{name: "defaults", input: url.Values{depthQueryParam: {"1"}}, depth: 1, page: 1, pageSize: 24},
		{name: "depth two", input: url.Values{depthQueryParam: {"2"}}, depth: 2, page: 1, pageSize: 24},
		{name: "depth three", input: url.Values{depthQueryParam: {"3"}}, depth: 3, page: 1, pageSize: 24},
		{name: "all depth", input: url.Values{depthQueryParam: {"all"}}, depth: repo.MaxLocationDepth, page: 1, pageSize: 24},
		{name: "custom page", input: url.Values{depthQueryParam: {"1"}, "page": {"2"}, pageSizeQueryParam: {"96"}}, depth: 1, page: 2, pageSize: 96},
		{name: "min size", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"1"}}, depth: 1, page: 1, pageSize: 1},
		{name: "empty depth", input: url.Values{}, badRequest: true},
		{name: "invalid depth", input: url.Values{depthQueryParam: {"direct"}}, badRequest: true},
		{name: "depth too large", input: url.Values{depthQueryParam: {"64"}}, badRequest: true},
		{name: "non-integer page", input: url.Values{depthQueryParam: {"1"}, "page": {"NaN"}}, badRequest: true},
		{name: "page integer overflow", input: url.Values{depthQueryParam: {"1"}, "page": {"999999999999999999999999999"}}, badRequest: true},
		{name: "zero page", input: url.Values{depthQueryParam: {"1"}, "page": {"0"}}, badRequest: true},
		{name: "negative page", input: url.Values{depthQueryParam: {"1"}, "page": {"-1"}}, badRequest: true},
		{name: "non-integer size", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"no"}}, badRequest: true},
		{name: "size integer overflow", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"999999999999999999999999999"}}, badRequest: true},
		{name: "zero size", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"0"}}, badRequest: true},
		{name: "size above maximum", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"97"}}, badRequest: true},
		{name: "negative size", input: url.Values{depthQueryParam: {"1"}, pageSizeQueryParam: {"-1"}}, badRequest: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, page, size, err := parseLocationDescendantQuery(tt.input)
			if tt.badRequest {
				var requestError *validate.RequestError
				require.ErrorAs(t, err, &requestError)
				require.Equal(t, http.StatusBadRequest, requestError.Status)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.depth, depth)
			require.Equal(t, tt.page, page)
			require.Equal(t, tt.pageSize, size)
		})
	}
}
