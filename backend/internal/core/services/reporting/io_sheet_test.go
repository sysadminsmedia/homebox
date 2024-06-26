package reporting

import (
	"bytes"
	"reflect"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

var (
	//go:embed .testdata/import/minimal.csv
	minimalImportCSV []byte

	//go:embed .testdata/import/fields.csv
	customFieldImportCSV []byte

	//go:embed .testdata/import/types.csv
	customTypesImportCSV []byte
)

func TestSheet_Read(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []ExportCSVRow
		wantErr bool
	}{
		{
			name: "minimal import",
			data: minimalImportCSV,
			want: []ExportCSVRow{
				{Location: LocationString{"loc"}, Name: "Item 1", Quantity: 1, Description: "Description 1"},
				{Location: LocationString{"loc"}, Name: "Item 2", Quantity: 2, Description: "Description 2"},
				{Location: LocationString{"loc"}, Name: "Item 3", Quantity: 3, Description: "Description 3"},
			},
		},
		{
			name: "custom field import",
			data: customFieldImportCSV,
			want: []ExportCSVRow{
				{
					Location: LocationString{"loc"}, Name: "Item 1", Quantity: 1, Description: "Description 1",
					Fields: []ExportItemFields{
						{Name: "Custom Field 1", Value: "Value 1[1]"},
						{Name: "Custom Field 2", Value: "Value 1[2]"},
						{Name: "Custom Field 3", Value: "Value 1[3]"},
					},
				},
				{
					Location: LocationString{"loc"}, Name: "Item 2", Quantity: 2, Description: "Description 2",
					Fields: []ExportItemFields{
						{Name: "Custom Field 1", Value: "Value 2[1]"},
						{Name: "Custom Field 2", Value: "Value 2[2]"},
						{Name: "Custom Field 3", Value: "Value 2[3]"},
					},
				},
				{
					Location: LocationString{"loc"}, Name: "Item 3", Quantity: 3, Description: "Description 3",
					Fields: []ExportItemFields{
						{Name: "Custom Field 1", Value: "Value 3[1]"},
						{Name: "Custom Field 2", Value: "Value 3[2]"},
						{Name: "Custom Field 3", Value: "Value 3[3]"},
					},
				},
			},
		},
		{
			name: "custom types import",
			data: customTypesImportCSV,
			want: []ExportCSVRow{
				{
					Name:     "Item 1",
					AssetID:  repo.AssetID(1),
					Location: LocationString{"Path", "To", "Location 1"},
					LabelStr: LabelString{"L1", "L2", "L3"},
				},
				{
					Name:     "Item 2",
					AssetID:  repo.AssetID(2),
					Location: LocationString{"Path", "To", "Location 2"},
					LabelStr: LabelString{"L1", "L2", "L3"},
				},
				{
					Name:     "Item 3",
					AssetID:  repo.AssetID(1000003),
					Location: LocationString{"Path", "To", "Location 3"},
					LabelStr: LabelString{"L1", "L2", "L3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)

			sheet := &IOSheet{}
			err := sheet.Read(reader)

			switch {
			case tt.wantErr:
				require.Error(t, err)
			default:
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.want, sheet.Rows)
			}
		})
	}
}

func Test_parseHeaders(t *testing.T) {
	tests := []struct {
		name             string
		rawHeaders       []string
		wantHbHeaders    map[string]int
		wantFieldHeaders []string
		wantErr          bool
	}{
		{
			name:             "no hombox headers",
			rawHeaders:       []string{"Header 1", "Header 2", "Header 3"},
			wantHbHeaders:    nil,
			wantFieldHeaders: nil,
			wantErr:          true,
		},
		{
			name:       "field headers only",
			rawHeaders: []string{"HB.location", "HB.name", "HB.field.1", "HB.field.2", "HB.field.3"},
			wantHbHeaders: map[string]int{
				"HB.location": 0,
				"HB.name":     1,
				"HB.field.1":  2,
				"HB.field.2":  3,
				"HB.field.3":  4,
			},
			wantFieldHeaders: []string{"HB.field.1", "HB.field.2", "HB.field.3"},
			wantErr:          false,
		},
		{
			name:       "mixed headers",
			rawHeaders: []string{"Header 1", "HB.name", "Header 2", "HB.field.2", "Header 3", "HB.field.3", "HB.location"},
			wantHbHeaders: map[string]int{
				"HB.name":     1,
				"HB.field.2":  3,
				"HB.field.3":  5,
				"HB.location": 6,
			},
			wantFieldHeaders: []string{"HB.field.2", "HB.field.3"},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHbHeaders, gotFieldHeaders, err := parseHeaders(tt.rawHeaders)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHbHeaders, tt.wantHbHeaders) {
				t.Errorf("parseHeaders() gotHbHeaders = %v, want %v", gotHbHeaders, tt.wantHbHeaders)
			}
			if !reflect.DeepEqual(gotFieldHeaders, tt.wantFieldHeaders) {
				t.Errorf("parseHeaders() gotFieldHeaders = %v, want %v", gotFieldHeaders, tt.wantFieldHeaders)
			}
		})
	}
}

func Test_determineSeparator(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    rune
		wantErr bool
	}{
		{
			name: "comma",
			args: args{
				data: []byte("a,b,c"),
			},
			want:    ',',
			wantErr: false,
		},
		{
			name: "tab",
			args: args{
				data: []byte("a\tb\tc"),
			},
			want:    '\t',
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				data: []byte("a;b;c"),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineSeparator(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineSeparator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determineSeparator() = %v, want %v", got, tt.want)
			}
		})
	}
}
