package repo

type BarcodeProduct struct {
	SearchEngineName string `json:"search_engine_name"`

	// Identifications
	ModelNumber  string `json:"modelNumber"`
	Manufacturer string `json:"manufacturer"`

	// Extras
	Country string `json:"notes"`
	Barcode string `json:"barcode"`

	// TODO: add image attachement
	// TODO: add asin?
	ImageURL    string `json:"imageURL"`
	ImageBase64 string `json:"imageBase64"`

	Item ItemCreate `json:"item"`
}
