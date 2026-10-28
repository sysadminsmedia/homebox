package v1

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
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
