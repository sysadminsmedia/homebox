package reporting

import (
	"strings"

	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

type ExportItemFields struct {
	Name  string
	Value string
}

type ExportCSVRow struct {
	ImportRef string         `csv:"HB.import_ref"`
	Location  LocationString `csv:"HB.location"`
	TagStr    TagString      `csv:"HB.tags|HB.labels"`
	AssetID   repo.AssetID   `csv:"HB.asset_id"`
	Archived  bool           `csv:"HB.archived"`
	URL       string         `csv:"HB.url"`

	Name        string `csv:"HB.name"`
	Quantity    int    `csv:"HB.quantity"`
	Description string `csv:"HB.description"`
	Insured     bool   `csv:"HB.insured"`
	Notes       string `csv:"HB.notes"`

	PurchasePrice float64    `csv:"HB.purchase_price"`
	PurchaseFrom  string     `csv:"HB.purchase_from"`
	PurchaseTime  types.Date `csv:"HB.purchase_time"`

	Manufacturer string `csv:"HB.manufacturer"`
	ModelNumber  string `csv:"HB.model_number"`
	SerialNumber string `csv:"HB.serial_number"`

	LifetimeWarranty bool       `csv:"HB.lifetime_warranty"`
	WarrantyExpires  types.Date `csv:"HB.warranty_expires"`
	WarrantyDetails  string     `csv:"HB.warranty_details"`

	SoldTo    string     `csv:"HB.sold_to"`
	SoldPrice float64    `csv:"HB.sold_price"`
	SoldTime  types.Date `csv:"HB.sold_time"`
	SoldNotes string     `csv:"HB.sold_notes"`

	Fields []ExportItemFields `csv:"-"`
}

// ============================================================================

// TagString is a string slice that is used to represent a list of tags.
//
// For example, a list of tags "Important; Work" would be represented as a
// TagString with the following values:
//
//	TagString{"Important", "Work"}
type TagString []string

func parseTagString(s string) TagString {
	v, _ := parseSeparatedString(s, ";")
	return v
}

func (ls TagString) String() string {
	return strings.Join(ls, "; ")
}

// ============================================================================

// LocationString is a string slice that is used to represent a location
// hierarchy.
//
// For example, a location hierarchy of "Home / Bedroom / Desk" would be
// represented as a LocationString with the following values:
//
//	LocationString{"Home", "Bedroom", "Desk"}
type LocationString []string

func parseLocationString(s string) LocationString {
	v, _ := parseSeparatedString(s, "/")
	return v
}

func (csf LocationString) String() string {
	return strings.Join(csf, " / ")
}

func fromPathSlice(s []repo.ItemPath) LocationString {
	return lo.Map(s, func(p repo.ItemPath, _ int) string {
		return p.Name
	})
}
