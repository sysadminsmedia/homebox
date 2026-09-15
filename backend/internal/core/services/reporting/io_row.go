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
	PurchaseDate     types.Date         `csv:"HB.purchase_date|HB.purchase_time"`
	WarrantyExpires  types.Date         `csv:"HB.warranty_expires"`
	SoldDate         types.Date         `csv:"HB.sold_date|HB.sold_time"`
	ImportRef        string             `csv:"HB.import_ref"`
	ParentImportRef  string             `csv:"HB.parent_import_ref"`
	URL              string             `csv:"HB.url"`
	Name             string             `csv:"HB.name"`
	Description      string             `csv:"HB.description"`
	Notes            string             `csv:"HB.notes"`
	PurchaseFrom     string             `csv:"HB.purchase_from"`
	Manufacturer     string             `csv:"HB.manufacturer"`
	ModelNumber      string             `csv:"HB.model_number"`
	SerialNumber     string             `csv:"HB.serial_number"`
	WarrantyDetails  string             `csv:"HB.warranty_details"`
	SoldTo           string             `csv:"HB.sold_to"`
	SoldNotes        string             `csv:"HB.sold_notes"`
	Location         LocationString     `csv:"HB.location"`
	TagStr           TagString          `csv:"HB.tags|HB.labels"`
	Fields           []ExportItemFields `csv:"-"`
	AssetID          repo.AssetID       `csv:"HB.asset_id"`
	Quantity         float64            `csv:"HB.quantity"`
	PurchasePrice    float64            `csv:"HB.purchase_price"`
	SoldPrice        float64            `csv:"HB.sold_price"`
	Archived         bool               `csv:"HB.archived"`
	Insured          bool               `csv:"HB.insured"`
	LifetimeWarranty bool               `csv:"HB.lifetime_warranty"`
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

func fromPathSlice(s []repo.EntityPath) LocationString {
	return lo.Map(s, func(p repo.EntityPath, _ int) string {
		return p.Name
	})
}
