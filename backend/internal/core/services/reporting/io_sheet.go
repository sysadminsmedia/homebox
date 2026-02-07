package reporting

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// IOSheet is the representation of a CSV/TSV sheet that is used for importing/exporting
// items from homebox. It is used to read/write the data from/to a CSV/TSV file given
// the standard format of the file.
//
// See ExportCSVRow for the format of the data in the sheet.
type IOSheet struct {
	headers []string
	custom  []int
	index   map[string]int
	Rows    []ExportCSVRow
}

func (s *IOSheet) indexHeaders() {
	s.index = make(map[string]int)

	for i, h := range s.headers {
		if strings.HasPrefix(h, "HB.field") {
			s.custom = append(s.custom, i)
		}

		if strings.HasPrefix(h, "HB.") {
			s.index[h] = i
		}
	}
}

// primaryCSVTag returns the primary header name from a csv struct tag which
// may contain alternatives separated by '|'. For example: "HB.tags|HB.labels"
// will return "HB.tags".
func primaryCSVTag(tag string) string {
	if tag == "" {
		return ""
	}
	if i := strings.Index(tag, "|"); i >= 0 {
		return strings.TrimSpace(tag[:i])
	}
	return tag
}

// csvTagAlternatives splits a tag with alternatives separated by '|'.
func csvTagAlternatives(tag string) []string {
	if tag == "" {
		return nil
	}
	parts := strings.Split(tag, "|")
	return lo.Map(parts, func(s string, _ int) string {
		return strings.TrimSpace(s)
	})
}

// findColumnForTag tries to find a column index for a csv tag. The tag may
// contain multiple alternatives separated by '|'. It will return the first
// matching column index it finds.
func (s *IOSheet) findColumnForTag(tag string) (int, bool) {
	if s.index == nil {
		s.indexHeaders()
	}

	// Try primary tag first and then alternatives.
	alts := csvTagAlternatives(tag)
	if len(alts) == 0 {
		return s.GetColumn(tag)
	}

	for _, t := range alts {
		if col, ok := s.GetColumn(t); ok {
			return col, true
		}
	}

	return 0, false
}

func (s *IOSheet) GetColumn(str string) (col int, ok bool) {
	if s.index == nil {
		s.indexHeaders()
	}

	col, ok = s.index[str]
	return
}

// Read reads a CSV/TSV and populates the "Rows" field with the data from the sheet
// Custom Fields are supported via the `HB.field.*` headers. The `HB.field.*` the "Name"
// of the field is the part after the `HB.field.` prefix. Additionally, Custom Fields with
// no value are excluded from the row.Fields slice, this includes empty strings.
//
// Note That
//   - the first row is assumed to be the header
//   - at least 1 row of data is required
//   - rows and columns must be rectangular (i.e. all rows must have the same number of columns)
func (s *IOSheet) Read(data io.Reader) error {
	sheet, err := readRawCsv(data)
	if err != nil {
		return err
	}

	if len(sheet) < 2 {
		return fmt.Errorf("sheet must have at least 1 row of data (header + 1)")
	}

	s.headers = sheet[0]
	s.Rows = make([]ExportCSVRow, len(sheet)-1)

	for i, row := range sheet[1:] {
		if len(row) != len(s.headers) {
			return fmt.Errorf("row has %d columns, expected %d", len(row), len(s.headers))
		}

		rowData := ExportCSVRow{}

		st := reflect.TypeOf(ExportCSVRow{})

		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			tag := field.Tag.Get("csv")
			if tag == "" || tag == "-" {
				continue
			}

			col, ok := s.findColumnForTag(tag)
			if !ok {
				continue
			}

			val := row[col]

			var v interface{}

			switch field.Type {
			case reflect.TypeOf(""):
				v = val
			case reflect.TypeOf(int(0)):
				v = parseInt(val)
			case reflect.TypeOf(bool(false)):
				v = parseBool(val)
			case reflect.TypeOf(float64(0)):
				v = parseFloat(val)

			// Custom Types
			case reflect.TypeOf(types.Date{}):
				v = types.DateFromString(val)
			case reflect.TypeOf(repo.AssetID(0)):
				v, _ = repo.ParseAssetID(val)
			case reflect.TypeOf(LocationString{}):
				v = parseLocationString(val)
			case reflect.TypeOf(TagString{}):
				v = parseTagString(val)
			}

			log.Debug().
				Str("tag", tag).
				Interface("val", v).
				Str("type", fmt.Sprintf("%T", v)).
				Msg("parsed value")

			// Nil values are not allowed at the moment. This may change.
			if v == nil {
				return fmt.Errorf("could not convert %q to %s", val, field.Type)
			}

			ptrField := reflect.ValueOf(&rowData).Elem().Field(i)
			ptrField.Set(reflect.ValueOf(v))
		}

		for _, col := range s.custom {
			colName := strings.TrimPrefix(s.headers[col], "HB.field.")
			customVal := row[col]
			if customVal == "" {
				continue
			}

			rowData.Fields = append(rowData.Fields, ExportItemFields{
				Name:  colName,
				Value: customVal,
			})
		}

		s.Rows[i] = rowData
	}

	return nil
}

// ReadItems writes the sheet to a writer.
func (s *IOSheet) ReadItems(ctx context.Context, items []repo.ItemOut, gid uuid.UUID, repos *repo.AllRepos, hbURL string) error {
	s.Rows = make([]ExportCSVRow, len(items))

	extraHeaders := map[string]struct{}{}

	for i := range items {
		item := items[i]

		// TODO: Support fetching nested locations
		locID := item.Location.ID

		locPaths, err := repos.Locations.PathForLoc(context.Background(), gid, locID)
		if err != nil {
			log.Error().Err(err).Msg("could not get location path")
			return err
		}

		locString := fromPathSlice(locPaths)

		tagString := lo.Map(item.Tags, func(l repo.TagSummary, _ int) string {
			return l.Name
		})

		url := generateItemURL(item, hbURL)

		customFields := lo.Map(item.Fields, func(f repo.ItemField, _ int) ExportItemFields {
			extraHeaders[f.Name] = struct{}{}
			return ExportItemFields{
				Name:  f.Name,
				Value: f.TextValue,
			}
		})

		s.Rows[i] = ExportCSVRow{
			// fill struct
			Location: locString,
			TagStr:   tagString,

			ImportRef:   item.ImportRef,
			AssetID:     item.AssetID,
			Name:        item.Name,
			Quantity:    item.Quantity,
			Description: item.Description,
			Insured:     item.Insured,
			Archived:    item.Archived,
			URL:         url,

			PurchasePrice: item.PurchasePrice,
			PurchaseFrom:  item.PurchaseFrom,
			PurchaseTime:  item.PurchaseTime,

			Manufacturer: item.Manufacturer,
			ModelNumber:  item.ModelNumber,
			SerialNumber: item.SerialNumber,

			LifetimeWarranty: item.LifetimeWarranty,
			WarrantyExpires:  item.WarrantyExpires,
			WarrantyDetails:  item.WarrantyDetails,

			SoldTo:    item.SoldTo,
			SoldTime:  item.SoldTime,
			SoldPrice: item.SoldPrice,
			SoldNotes: item.SoldNotes,

			Notes:  item.Notes,
			Fields: customFields,
		}
	}

	// Extract and sort additional headers for deterministic output
	customHeaders := lo.Keys(extraHeaders)

	sort.Strings(customHeaders)

	st := reflect.TypeOf(ExportCSVRow{})

	// Write headers (use the primary tag when alternatives are provided)
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag := field.Tag.Get("csv")
		if tag == "" || tag == "-" {
			continue
		}

		s.headers = append(s.headers, primaryCSVTag(tag))
	}

	for _, h := range customHeaders {
		s.headers = append(s.headers, "HB.field."+h)
	}

	return nil
}

func generateItemURL(item repo.ItemOut, d string) string {
	url := ""
	if item.ID != uuid.Nil {
		url = fmt.Sprintf("%s/item/%s", d, item.ID.String())
	}
	return url
}

// CSV writes the current sheet to a 2d array, for compatibility with TSV/CSV files.
func (s *IOSheet) CSV() ([][]string, error) {
	memcsv := make([][]string, len(s.Rows)+1)

	memcsv[0] = s.headers

	// use struct tags in rows to dertmine column order
	for i, row := range s.Rows {
		rowIdx := i + 1

		memcsv[rowIdx] = make([]string, len(s.headers))

		st := reflect.TypeOf(row)

		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			tag := field.Tag.Get("csv")
			if tag == "" || tag == "-" {
				continue
			}

			col, ok := s.findColumnForTag(tag)
			if !ok {
				continue
			}

			val := reflect.ValueOf(row).Field(i)

			var v string

			switch field.Type {
			case reflect.TypeOf(""):
				v = val.String()
			case reflect.TypeOf(int(0)):
				v = strconv.Itoa(int(val.Int()))
			case reflect.TypeOf(bool(false)):
				v = strconv.FormatBool(val.Bool())
			case reflect.TypeOf(float64(0)):
				v = strconv.FormatFloat(val.Float(), 'f', -1, 64)

			// Custom Types
			case reflect.TypeOf(types.Date{}):
				v = val.Interface().(types.Date).String()
			case reflect.TypeOf(repo.AssetID(0)):
				v = val.Interface().(repo.AssetID).String()
			case reflect.TypeOf(LocationString{}):
				v = val.Interface().(LocationString).String()
			case reflect.TypeOf(TagString{}):
				v = val.Interface().(TagString).String()
			default:
				log.Debug().Str("type", field.Type.String()).Msg("unknown type")
			}

			memcsv[rowIdx][col] = v
		}

		for _, f := range row.Fields {
			col, ok := s.GetColumn("HB.field." + f.Name)
			if !ok {
				continue
			}

			memcsv[i+1][col] = f.Value
		}
	}

	return memcsv, nil
}
