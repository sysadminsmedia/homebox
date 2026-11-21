package repo

type BarcodeProduct struct {
	SearchEngineName string `json:"search_engine_name"`

	// Identifications
	ModelNumber  string `json:"modelNumber"`
	Manufacturer string `json:"manufacturer"`

	// Extras
	Country string `json:"notes"`
	Barcode string `json:"barcode"`

	ImageURL    string            `json:"imageURL"`
	ImageBase64 string            `json:"imageBase64"`
	Fields      []EntityFieldData `json:"fields,omitempty"`

	Item EntityCreate `json:"item"`
}
