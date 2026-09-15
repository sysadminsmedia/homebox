package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exports are portable across engines, and sqlite reports BOOLEAN columns as
// int64 0/1 — which lands in the export JSON as 0/1 and decodes back to a
// float64. Postgres rejects that outright ("unable to encode 0 into binary
// format for bool"), so without coercion a sqlite-sourced backup can never be
// restored into postgres. See issue #1641.
func TestValueToBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      any
		want    bool
		wantErr bool
	}{
		{name: "native false", in: false, want: false},
		{name: "native true", in: true, want: true},
		{name: "sqlite zero", in: float64(0), want: false},
		{name: "sqlite one", in: float64(1), want: true},
		{name: "json.Number zero", in: json.Number("0"), want: false},
		{name: "json.Number one", in: json.Number("1"), want: true},
		{name: "string true", in: "true", want: true},
		{name: "string false", in: "false", want: false},
		{name: "string one", in: "1", want: true},
		{name: "string padded", in: " t ", want: true},
		{name: "nonsense string", in: "banana", wantErr: true},
		{name: "unsupported type", in: []string{"nope"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valueToBool(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoerceBoolColumns(t *testing.T) {
	t.Parallel()

	boolCols := map[string]struct{}{"archived": {}, "insured": {}, "missing": {}}
	row := map[string]any{
		"archived": float64(0),
		"insured":  float64(1),
		"name":     "Drill",
		"quantity": float64(2),
		"notes":    nil,
	}

	require.NoError(t, coerceBoolColumns(row, boolCols))

	assert.Equal(t, false, row["archived"], "0 should become a real false")
	assert.Equal(t, true, row["insured"], "1 should become a real true")

	// Non-boolean columns must be left exactly as they were: quantity is a
	// float column and would be corrupted by a blanket numeric conversion.
	quantity, ok := row["quantity"].(float64)
	require.True(t, ok, "quantity must stay a float, not be coerced to a bool")
	assert.InDelta(t, 2, quantity, 0)
	assert.Equal(t, "Drill", row["name"])
	assert.Nil(t, row["notes"])

	// A bool column absent from the row stays absent rather than being
	// inserted as a spurious false.
	_, present := row["missing"]
	assert.False(t, present)
}

func TestCoerceBoolColumnsRejectsGarbage(t *testing.T) {
	t.Parallel()

	err := coerceBoolColumns(
		map[string]any{"archived": "not-a-bool"},
		map[string]struct{}{"archived": {}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "archived")
}

// boolColumns drives a dialect-specific introspection query; this pins the
// sqlite one (pragma_table_info with a bound table name) against a real
// database so a driver change can't silently turn it into a no-op, which
// would quietly restore the broken behaviour.
func TestBoolColumnsSQLite(t *testing.T) {
	ctx := context.Background()
	db := tClient.Sql()

	cols, err := boolColumns(ctx, db, "sqlite3", "entities")
	require.NoError(t, err)

	assert.Contains(t, cols, "archived")
	assert.Contains(t, cols, "insured")
	assert.NotContains(t, cols, "name")
	assert.NotContains(t, cols, "quantity")
}

func TestBoolColumnsRejectsBadIdentifier(t *testing.T) {
	t.Parallel()

	_, err := boolColumns(context.Background(), tClient.Sql(), "sqlite3", "entities; DROP TABLE entities")
	require.Error(t, err)
}
