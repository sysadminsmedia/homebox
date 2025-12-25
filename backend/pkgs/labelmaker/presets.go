// Package labelmaker provides functionality for generating and printing labels
package labelmaker

// LabelPreset represents a predefined label size configuration
type LabelPreset struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Brand       string  `json:"brand"`
	Width       float64 `json:"width"`      // Width in mm
	Height      float64 `json:"height"`     // Height in mm
	Continuous  bool    `json:"continuous"` // Whether this is continuous tape
	TwoColor    bool    `json:"twoColor"`   // Whether this supports two-color printing (black/red)
	Description string  `json:"description"`

	// Sheet layout information (for Avery-style sheet labels)
	// If SheetLayout is set, labels are arranged on a printable sheet
	SheetLayout *SheetLayout `json:"sheetLayout,omitempty"`
}

// SheetLayout defines how labels are arranged on a printable sheet
type SheetLayout struct {
	PageWidth  float64 `json:"pageWidth"`  // Sheet width in mm (e.g., 215.9 for Letter, 210 for A4)
	PageHeight float64 `json:"pageHeight"` // Sheet height in mm (e.g., 279.4 for Letter, 297 for A4)
	Columns    int     `json:"columns"`    // Number of labels across
	Rows       int     `json:"rows"`       // Number of labels down
	MarginTop  float64 `json:"marginTop"`  // Top margin in mm
	MarginLeft float64 `json:"marginLeft"` // Left margin in mm
	GutterH    float64 `json:"gutterH"`    // Horizontal gap between labels in mm
	GutterV    float64 `json:"gutterV"`    // Vertical gap between labels in mm
}

// LabelPresets contains all supported label size presets
var LabelPresets = []LabelPreset{
	// Brother DK series - Die-cut labels
	{
		Key:         "brother_dk1201",
		Name:        "DK-1201",
		Brand:       "Brother",
		Width:       29,
		Height:      90,
		Continuous:  false,
		Description: "Standard Address Labels 1.1\" x 3.5\"",
	},
	{
		Key:         "brother_dk1202",
		Name:        "DK-1202",
		Brand:       "Brother",
		Width:       62,
		Height:      100,
		Continuous:  false,
		Description: "Shipping Labels 2.4\" x 3.9\"",
	},
	{
		Key:         "brother_dk1204",
		Name:        "DK-1204",
		Brand:       "Brother",
		Width:       17,
		Height:      54,
		Continuous:  false,
		Description: "Multi-Purpose Labels 0.66\" x 2.1\"",
	},
	{
		Key:         "brother_dk1208",
		Name:        "DK-1208",
		Brand:       "Brother",
		Width:       38,
		Height:      90,
		Continuous:  false,
		Description: "Large Address Labels 1.4\" x 3.5\"",
	},
	{
		Key:         "brother_dk1209",
		Name:        "DK-1209",
		Brand:       "Brother",
		Width:       29,
		Height:      62,
		Continuous:  false,
		Description: "Small Address Labels 1.1\" x 2.4\"",
	},
	{
		Key:         "brother_dk1221",
		Name:        "DK-1221",
		Brand:       "Brother",
		Width:       23,
		Height:      23,
		Continuous:  false,
		Description: "Square Labels 0.9\" x 0.9\"",
	},
	{
		Key:         "brother_dk1234",
		Name:        "DK-1234",
		Brand:       "Brother",
		Width:       60,
		Height:      86,
		Continuous:  false,
		Description: "Name Badge Labels 2.3\" x 3.4\"",
	},

	// Brother DK series - Continuous tape
	{
		Key:         "brother_dk2205",
		Name:        "DK-2205",
		Brand:       "Brother",
		Width:       62,
		Height:      30.48, // Default length, can be adjusted
		Continuous:  true,
		Description: "Continuous Paper Tape 2.4\" wide",
	},
	{
		Key:         "brother_dk22251",
		Name:        "DK-22251",
		Brand:       "Brother",
		Width:       62,
		Height:      30.48, // Default length, can be adjusted
		Continuous:  true,
		TwoColor:    true,
		Description: "Continuous Paper Tape 2.4\" wide (Black/Red)",
	},
	{
		Key:         "brother_dk2210",
		Name:        "DK-2210",
		Brand:       "Brother",
		Width:       29,
		Height:      30.48,
		Continuous:  true,
		Description: "Continuous Paper Tape 1.1\" wide",
	},
	{
		Key:         "brother_dk2211",
		Name:        "DK-2211",
		Brand:       "Brother",
		Width:       29,
		Height:      15.24,
		Continuous:  true,
		Description: "Continuous Film Tape 1.1\" wide",
	},
	{
		Key:         "brother_dk2212",
		Name:        "DK-2212",
		Brand:       "Brother",
		Width:       62,
		Height:      15.24,
		Continuous:  true,
		Description: "Continuous Film Tape 2.4\" wide",
	},
	{
		Key:         "brother_dk2243",
		Name:        "DK-2243",
		Brand:       "Brother",
		Width:       102,
		Height:      30.48,
		Continuous:  true,
		Description: "Continuous Paper Tape 4\" wide",
	},
	{
		Key:         "brother_dk2246",
		Name:        "DK-2246",
		Brand:       "Brother",
		Width:       103,
		Height:      30.48,
		Continuous:  true,
		Description: "Continuous Paper Tape 4.07\" wide",
	},

	// Dymo LabelWriter series
	{
		Key:         "dymo_30252",
		Name:        "30252",
		Brand:       "Dymo",
		Width:       28.6,
		Height:      89,
		Continuous:  false,
		Description: "Address Labels 1-1/8\" x 3-1/2\"",
	},
	{
		Key:         "dymo_30256",
		Name:        "30256",
		Brand:       "Dymo",
		Width:       59,
		Height:      102,
		Continuous:  false,
		Description: "Large Shipping Labels 2-5/16\" x 4\"",
	},
	{
		Key:         "dymo_30323",
		Name:        "30323",
		Brand:       "Dymo",
		Width:       54,
		Height:      102,
		Continuous:  false,
		Description: "Shipping Labels 2-1/8\" x 4\"",
	},
	{
		Key:         "dymo_30330",
		Name:        "30330",
		Brand:       "Dymo",
		Width:       19,
		Height:      51,
		Continuous:  false,
		Description: "Return Address Labels 3/4\" x 2\"",
	},
	{
		Key:         "dymo_30332",
		Name:        "30332",
		Brand:       "Dymo",
		Width:       25,
		Height:      25,
		Continuous:  false,
		Description: "Square Labels 1\" x 1\"",
	},
	{
		Key:         "dymo_30334",
		Name:        "30334",
		Brand:       "Dymo",
		Width:       32,
		Height:      57,
		Continuous:  false,
		Description: "Multi-Purpose Labels 1-1/4\" x 2-1/4\"",
	},
	{
		Key:         "dymo_30336",
		Name:        "30336",
		Brand:       "Dymo",
		Width:       25,
		Height:      54,
		Continuous:  false,
		Description: "Small Multipurpose Labels 1\" x 2-1/8\"",
	},
	{
		Key:         "dymo_30346",
		Name:        "30346",
		Brand:       "Dymo",
		Width:       13,
		Height:      51,
		Continuous:  false,
		Description: "Library Labels 1/2\" x 1-7/8\"",
	},
	{
		Key:         "dymo_30364",
		Name:        "30364",
		Brand:       "Dymo",
		Width:       57,
		Height:      32,
		Continuous:  false,
		Description: "Name Badge Labels 2-1/4\" x 1-1/4\"",
	},

	// Avery sheet labels (common sizes) - US Letter (215.9 x 279.4 mm)
	{
		Key:         "avery_5160",
		Name:        "5160/8160",
		Brand:       "Avery",
		Width:       66.7,
		Height:      25.4,
		Continuous:  false,
		Description: "Easy Peel Address Labels 1\" x 2-5/8\" (30/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    3,
			Rows:       10,
			MarginTop:  12.7,
			MarginLeft: 4.8,
			GutterH:    3.2,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5163",
		Name:        "5163/8163",
		Brand:       "Avery",
		Width:       101.6,
		Height:      50.8,
		Continuous:  false,
		Description: "Shipping Labels 2\" x 4\" (10/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    2,
			Rows:       5,
			MarginTop:  12.7,
			MarginLeft: 4.8,
			GutterH:    3.2,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5164",
		Name:        "5164/8164",
		Brand:       "Avery",
		Width:       101.6,
		Height:      85.7,
		Continuous:  false,
		Description: "Shipping Labels 3-1/3\" x 4\" (6/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    2,
			Rows:       3,
			MarginTop:  12.7,
			MarginLeft: 4.8,
			GutterH:    3.2,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5167",
		Name:        "5167/8167",
		Brand:       "Avery",
		Width:       44.5,
		Height:      12.7,
		Continuous:  false,
		Description: "Return Address Labels 1/2\" x 1-3/4\" (80/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    4,
			Rows:       20,
			MarginTop:  12.7,
			MarginLeft: 7.9,
			GutterH:    6.35,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5408",
		Name:        "5408",
		Brand:       "Avery",
		Width:       19.1,
		Height:      12.7,
		Continuous:  false,
		Description: "Removable Color-Coding Labels 1/2\" x 3/4\" (768/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    8,
			Rows:       96,
			MarginTop:  12.7,
			MarginLeft: 12.7,
			GutterH:    4.8,
			GutterV:    0,
		},
	},
	// Additional Avery labels
	{
		Key:         "avery_5260",
		Name:        "5260/8260",
		Brand:       "Avery",
		Width:       66.7,
		Height:      25.4,
		Continuous:  false,
		Description: "Easy Peel Clear Address Labels 1\" x 2-5/8\" (30/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    3,
			Rows:       10,
			MarginTop:  12.7,
			MarginLeft: 4.8,
			GutterH:    3.2,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5168",
		Name:        "5168",
		Brand:       "Avery",
		Width:       88.9,
		Height:      139.7,
		Continuous:  false,
		Description: "Shipping Labels 3-1/2\" x 5\" (4/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    2,
			Rows:       2,
			MarginTop:  0,
			MarginLeft: 19.05,
			GutterH:    0,
			GutterV:    0,
		},
	},
	{
		Key:         "avery_5165",
		Name:        "5165/8165",
		Brand:       "Avery",
		Width:       215.9,
		Height:      279.4,
		Continuous:  false,
		Description: "Full Sheet Label 8-1/2\" x 11\" (1/sheet)",
		SheetLayout: &SheetLayout{
			PageWidth:  215.9,
			PageHeight: 279.4,
			Columns:    1,
			Rows:       1,
			MarginTop:  0,
			MarginLeft: 0,
			GutterH:    0,
			GutterV:    0,
		},
	},

	// Brady label presets
	{
		Key:         "brady_b427",
		Name:        "B-427",
		Brand:       "Brady",
		Width:       25.4,
		Height:      9.5,
		Continuous:  false,
		Description: "Self-Laminating Wire Markers 1\" x 3/8\"",
	},
	{
		Key:         "brady_b428",
		Name:        "B-428",
		Brand:       "Brady",
		Width:       19.1,
		Height:      19.1,
		Continuous:  false,
		Description: "Metallized Polyester Labels 3/4\" x 3/4\"",
	},
	{
		Key:         "brady_b499",
		Name:        "B-499",
		Brand:       "Brady",
		Width:       50.8,
		Height:      25.4,
		Continuous:  false,
		Description: "Nylon Cloth Labels 2\" x 1\"",
	},
	{
		Key:         "brady_ptml",
		Name:        "PTL-M",
		Brand:       "Brady",
		Width:       38.1,
		Height:      12.7,
		Continuous:  false,
		Description: "TLS 2200 Multi-Purpose Labels 1-1/2\" x 1/2\"",
	},

	// Zebra common sizes
	{
		Key:         "zebra_2x1",
		Name:        "2\" x 1\"",
		Brand:       "Zebra",
		Width:       50.8,
		Height:      25.4,
		Continuous:  false,
		Description: "Standard Barcode Labels 2\" x 1\"",
	},
	{
		Key:         "zebra_4x2",
		Name:        "4\" x 2\"",
		Brand:       "Zebra",
		Width:       101.6,
		Height:      50.8,
		Continuous:  false,
		Description: "Shipping Labels 4\" x 2\"",
	},
	{
		Key:         "zebra_4x6",
		Name:        "4\" x 6\"",
		Brand:       "Zebra",
		Width:       101.6,
		Height:      152.4,
		Continuous:  false,
		Description: "Shipping Labels 4\" x 6\"",
	},

	// Custom - allows user-defined dimensions
	{
		Key:         "custom",
		Name:        "Custom",
		Brand:       "",
		Width:       0,
		Height:      0,
		Continuous:  false,
		Description: "Custom label dimensions",
	},
}

// GetPresetByKey returns a preset by its key, or nil if not found
func GetPresetByKey(key string) *LabelPreset {
	for _, preset := range LabelPresets {
		if preset.Key == key {
			return &preset
		}
	}
	return nil
}

// GetPresetsByBrand returns all presets for a specific brand
func GetPresetsByBrand(brand string) []LabelPreset {
	var result []LabelPreset
	for _, preset := range LabelPresets {
		if preset.Brand == brand {
			result = append(result, preset)
		}
	}
	return result
}

// GetBrands returns a list of unique brand names
func GetBrands() []string {
	brands := make(map[string]bool)
	var result []string
	for _, preset := range LabelPresets {
		if preset.Brand != "" && !brands[preset.Brand] {
			brands[preset.Brand] = true
			result = append(result, preset.Brand)
		}
	}
	return result
}
