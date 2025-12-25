package printer

// BrotherMediaType defines the characteristics of a Brother label media
type BrotherMediaType struct {
	Name               string // Human-readable name
	WidthMM            int    // Width in mm
	LengthMM           int    // Length in mm (0 for continuous)
	MediaCode          byte   // Protocol media type code (0x0A=continuous, 0x0B=die-cut)
	TwoColor           bool   // Whether this is a two-color roll
	PrintableDotsWidth int    // Printable width in dots at 300 DPI
}

// Brother QL media definitions
// Reference: Brother QL-800 series command reference
var BrotherMediaTypes = map[string]BrotherMediaType{
	// Continuous rolls (62mm width)
	"DK-22205": {
		Name:               "62mm Continuous (White)",
		WidthMM:            62,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 696,
	},
	"DK-22251": {
		Name:               "62mm Continuous (Black/Red on White)",
		WidthMM:            62,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           true,
		PrintableDotsWidth: 696,
	},
	"DK-22211": {
		Name:               "29mm Continuous Film (White)",
		WidthMM:            29,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 306,
	},
	"DK-22210": {
		Name:               "29mm Continuous Paper (White)",
		WidthMM:            29,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 306,
	},
	"DK-22212": {
		Name:               "62mm Continuous Film (White)",
		WidthMM:            62,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 696,
	},
	"DK-22214": {
		Name:               "12mm Continuous Paper (White)",
		WidthMM:            12,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 106,
	},
	"DK-22223": {
		Name:               "50mm Continuous Paper (White)",
		WidthMM:            50,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 554,
	},
	"DK-22225": {
		Name:               "38mm Continuous Paper (White)",
		WidthMM:            38,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 413,
	},
	"DK-22243": {
		Name:               "102mm Continuous Paper (White)",
		WidthMM:            102,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 1164,
	},
	"DK-22246": {
		Name:               "103.6mm Continuous Paper (White)",
		WidthMM:            103,
		LengthMM:           0, // Continuous
		MediaCode:          0x0A,
		TwoColor:           false,
		PrintableDotsWidth: 1200,
	},

	// Die-cut labels
	"DK-11201": {
		Name:               "29mm x 90mm Address Labels",
		WidthMM:            29,
		LengthMM:           90,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 306,
	},
	"DK-11202": {
		Name:               "62mm x 100mm Shipping Labels",
		WidthMM:            62,
		LengthMM:           100,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 696,
	},
	"DK-11203": {
		Name:               "17mm x 87mm File Folder Labels",
		WidthMM:            17,
		LengthMM:           87,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 165,
	},
	"DK-11204": {
		Name:               "17mm x 54mm Multi-Purpose Labels",
		WidthMM:            17,
		LengthMM:           54,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 165,
	},
	"DK-11207": {
		Name:               "58mm CD/DVD Labels",
		WidthMM:            58,
		LengthMM:           58,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 618,
	},
	"DK-11208": {
		Name:               "38mm x 90mm Address Labels",
		WidthMM:            38,
		LengthMM:           90,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 413,
	},
	"DK-11209": {
		Name:               "29mm x 62mm Small Address Labels",
		WidthMM:            29,
		LengthMM:           62,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 306,
	},
	"DK-11240": {
		Name:               "51mm x 102mm Barcode Labels",
		WidthMM:            51,
		LengthMM:           102,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 554,
	},
	"DK-11241": {
		Name:               "102mm x 152mm Large Shipping Labels",
		WidthMM:            102,
		LengthMM:           152,
		MediaCode:          0x0B,
		TwoColor:           false,
		PrintableDotsWidth: 1164,
	},
}

// GetBrotherMedia returns media type info for a given media identifier
func GetBrotherMedia(mediaType string) (BrotherMediaType, bool) {
	media, ok := BrotherMediaTypes[mediaType]
	return media, ok
}

// GetSupportedBrotherMedia returns a list of all supported media types
func GetSupportedBrotherMedia() []string {
	types := make([]string, 0, len(BrotherMediaTypes))
	for k := range BrotherMediaTypes {
		types = append(types, k)
	}
	return types
}
