package v1

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

const (
	barcodeHTTPTimeoutSec = 10
	schemeHTTPS           = "https"

	// barcodeMaxResponseBytes caps how much of an upstream response is decoded so a
	// misbehaving third-party API cannot drive unbounded allocation.
	barcodeMaxResponseBytes = 4 << 20 // 4 MiB
)

// barcodeHTTPClient is shared by every barcode lookup so outbound requests reuse a
// single connection pool instead of allocating a client per call.
var barcodeHTTPClient = &http.Client{Timeout: barcodeHTTPTimeoutSec * time.Second}

// discogsDetailLookupTimeout bounds enrichment across all releases returned by
// one search. Keep it replaceable so tests can exercise deadline behavior
// without waiting for the production timeout.
var discogsDetailLookupTimeout = barcodeHTTPTimeoutSec * time.Second

// decodeBarcodeResponse streams the response body into v. Decoding directly off the
// body avoids materializing the whole payload before parsing it.
func decodeBarcodeResponse(body io.Reader, v any) error {
	return json.NewDecoder(io.LimitReader(body, barcodeMaxResponseBytes)).Decode(v)
}

// flexibleString is a string that can be unmarshaled from either a JSON
// string or a JSON number. upcitemdb.com sometimes returns price fields
// (e.g. list_price, shipping) as numbers rather than strings, which would
// otherwise cause json.Unmarshal to fail on a plain string field.
type flexibleString string

// UnmarshalJSON accepts a JSON string, number, or null and stores the result
// as a Go string. Composite tokens (objects and arrays) are rejected so that
// unexpected shapes surface as clear errors instead of being silently stored
// as their raw representation.
func (f *flexibleString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		*f = ""
		return nil
	}

	switch c := trimmed[0]; {
	case c == '"':
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		*f = flexibleString(s)
		return nil
	case c == 'n':
		if string(trimmed) != "null" {
			return fmt.Errorf("flexibleString: unexpected token %q", string(trimmed))
		}
		*f = ""
		return nil
	case c == '-' || (c >= '0' && c <= '9'):
		// Validate the token is a well-formed JSON number, then store its literal form.
		var n json.Number
		if err := json.Unmarshal(trimmed, &n); err != nil {
			return err
		}
		*f = flexibleString(n.String())
		return nil
	default:
		return fmt.Errorf("flexibleString: cannot unmarshal %s into string", string(trimmed))
	}
}

type UPCITEMDBResponse struct {
	Code   string `json:"code"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Items  []struct {
		Ean                  string   `json:"ean"`
		Title                string   `json:"title"`
		Description          string   `json:"description"`
		Upc                  string   `json:"upc"`
		Brand                string   `json:"brand"`
		Model                string   `json:"model"`
		Color                string   `json:"color"`
		Size                 string   `json:"size"`
		Dimension            string   `json:"dimension"`
		Weight               string   `json:"weight"`
		Category             string   `json:"category"`
		LowestRecordedPrice  float64  `json:"lowest_recorded_price"`
		HighestRecordedPrice float64  `json:"highest_recorded_price"`
		Images               []string `json:"images"`
		Offers               []struct {
			Merchant     string         `json:"merchant"`
			Domain       string         `json:"domain"`
			Title        string         `json:"title"`
			Currency     string         `json:"currency"`
			ListPrice    flexibleString `json:"list_price"`
			Price        float64        `json:"price"`
			Shipping     flexibleString `json:"shipping"`
			Condition    string         `json:"condition"`
			Availability string         `json:"availability"`
			Link         string         `json:"link"`
			UpdatedT     int            `json:"updated_t"`
		} `json:"offers"`
		Asin string `json:"asin"`
		Elid string `json:"elid"`
	} `json:"items"`
}

type OpenFactsResponse struct {
	Code    string           `json:"code"`
	Status  int              `json:"status"`
	Product openFactsProduct `json:"product"`
}

// Open Food Facts, Open Beauty Facts, and Open Products Facts share the same
// product response shape and API path, so one mapper can safely serve all three.
type openFactsProduct struct {
	ProductName   string `json:"product_name"`
	Brands        string `json:"brands"`
	Categories    string `json:"categories"`
	ImageFrontURL string `json:"image_front_url"`
	ImageURL      string `json:"image_url"`
	Quantity      string `json:"quantity"`
	GenericName   string `json:"generic_name"`
}

type openFactsSource struct {
	Name    string
	BaseURL string
}

var openFactsSources = []openFactsSource{
	{Name: "openfoodfacts.org", BaseURL: "https://world.openfoodfacts.org"},
	{Name: "openbeautyfacts.org", BaseURL: "https://world.openbeautyfacts.org"},
	{Name: "openproductsfacts.org", BaseURL: "https://world.openproductsfacts.org"},
}

type BARCODESPIDER_COMResponse struct {
	ItemResponse struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"item_response"`
	ItemAttributes struct {
		Title          string `json:"title"`
		Upc            string `json:"upc"`
		Ean            string `json:"ean"`
		ParentCategory string `json:"parent_category"`
		Category       string `json:"category"`
		Brand          string `json:"brand"`
		Model          string `json:"model"`
		Mpn            string `json:"mpn"`
		Manufacturer   string `json:"manufacturer"`
		Publisher      string `json:"publisher"`
		Asin           string `json:"asin"`
		Color          string `json:"color"`
		Size           string `json:"size"`
		Weight         string `json:"weight"`
		Image          string `json:"image"`
		IsAdult        string `json:"is_adult"`
		Description    string `json:"description"`
	} `json:"item_attributes"`
	Stores []struct {
		StoreName string `json:"store_name"`
		Title     string `json:"title"`
		Image     string `json:"image"`
		Price     string `json:"price"`
		Currency  string `json:"currency"`
		Link      string `json:"link"`
		Updated   string `json:"updated"`
	} `json:"Stores"`
}

type DiscogsSearchResponse struct {
	Results []struct {
		ID         int      `json:"id"`
		Title      string   `json:"title"`
		Country    string   `json:"country"`
		Year       string   `json:"year"`
		Format     []string `json:"format"`
		Label      []string `json:"label"`
		CatalogNo  string   `json:"catno"`
		Genre      []string `json:"genre"`
		Style      []string `json:"style"`
		CoverImage string   `json:"cover_image"`
		URI        string   `json:"uri"`
	} `json:"results"`
}

type DiscogsReleaseResponse struct {
	ID       int      `json:"id"`
	MasterID int      `json:"master_id"`
	Country  string   `json:"country"`
	Released string   `json:"released"`
	Notes    string   `json:"notes"`
	Genres   []string `json:"genres"`
	Styles   []string `json:"styles"`
	Labels   []struct {
		Name  string `json:"name"`
		CatNo string `json:"catno"`
	} `json:"labels"`
	Formats []struct {
		Name         string   `json:"name"`
		Qty          string   `json:"qty"`
		Text         string   `json:"text"`
		Descriptions []string `json:"descriptions"`
	} `json:"formats"`
	Companies []struct {
		Name           string `json:"name"`
		EntityTypeName string `json:"entity_type_name"`
	} `json:"companies"`
	Identifiers []struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"identifiers"`
	Tracklist []struct {
		Position string `json:"position"`
		Title    string `json:"title"`
		Duration string `json:"duration"`
		Type     string `json:"type_"`
	} `json:"tracklist"`
}

func lookupUPCItemDB(iEan string) ([]repo.BarcodeProduct, error) {
	resp, err := barcodeHTTPClient.Get("https://api.upcitemdb.com/prod/trial/lookup?upc=" + url.QueryEscape(iEan))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var result UPCITEMDBResponse
	if err := decodeBarcodeResponse(resp.Body, &result); err != nil {
		log.Error().Msg("Can not unmarshal JSON from upcitemdb.com")
		return nil, err
	}

	res := make([]repo.BarcodeProduct, 0, len(result.Items))

	for _, it := range result.Items {
		var p repo.BarcodeProduct
		p.SearchEngineName = "upcitemdb.com"
		p.Barcode = iEan

		p.Item.Description = it.Description
		p.Item.Name = it.Title
		p.Manufacturer = it.Brand
		p.ModelNumber = it.Model
		if len(it.Images) != 0 {
			p.ImageURL = it.Images[0]
		}

		res = append(res, p)
	}

	return res, nil
}

func lookupBarcodespider(tokenAPI string, iEan string) ([]repo.BarcodeProduct, error) {
	if len(tokenAPI) == 0 {
		return nil, errors.New("no api token configured for barcodespider. " +
			"Please define the api token in environment variable HBOX_BARCODE_TOKEN_BARCODESPIDER")
	}

	req, err := http.NewRequest(
		"GET", "https://api.barcodespider.com/v1/lookup?upc="+url.QueryEscape(iEan), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("token", tokenAPI)

	resp, err := barcodeHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("barcodespider API returned status code: %d", resp.StatusCode)
	}

	var result BARCODESPIDER_COMResponse
	if err := decodeBarcodeResponse(resp.Body, &result); err != nil {
		log.Error().Msg("Can not unmarshal JSON from barcodespider.com")
		return nil, err
	}

	var p repo.BarcodeProduct
	p.Barcode = iEan
	p.SearchEngineName = "barcodespider.com"
	p.Item.Name = result.ItemAttributes.Title
	p.Item.Description = result.ItemAttributes.Description
	p.Manufacturer = result.ItemAttributes.Brand
	p.ModelNumber = result.ItemAttributes.Model
	p.ImageURL = result.ItemAttributes.Image

	return []repo.BarcodeProduct{p}, nil
}

func lookupDiscogs(ctx context.Context, tokenAPI string, iEan string) ([]repo.BarcodeProduct, error) {
	if strings.TrimSpace(tokenAPI) == "" {
		return nil, errors.New("no Discogs API token configured; set HBOX_BARCODE_TOKEN_DISCOGS")
	}

	endpoint := "https://api.discogs.com/database/search?type=release&format=vinyl&per_page=10&barcode=" + url.QueryEscape(iEan)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Discogs token="+sanitizeHeader(strings.TrimSpace(tokenAPI)))
	req.Header.Set("User-Agent", "Homebox-Discogs-Barcode/1.0 (+https://github.com/sysadminsmedia/homebox)")

	resp, err := barcodeHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discogs API returned status code: %d", resp.StatusCode)
	}

	var result DiscogsSearchResponse
	if err := decodeBarcodeResponse(resp.Body, &result); err != nil {
		return nil, err
	}

	detailCtx, cancelDetails := context.WithTimeout(ctx, discogsDetailLookupTimeout)
	defer cancelDetails()

	products := make([]repo.BarcodeProduct, 0, len(result.Results))
	for _, release := range result.Results {
		if strings.TrimSpace(release.Title) == "" {
			continue
		}

		var p repo.BarcodeProduct
		p.SearchEngineName = "discogs.com"
		p.Barcode = iEan
		p.Item.Name = release.Title
		p.ImageURL = release.CoverImage

		discogsURL := "https://www.discogs.com/release/" + fmt.Sprint(release.ID)
		if strings.HasPrefix(release.URI, "/release/") {
			discogsURL = "https://www.discogs.com" + release.URI
		}

		var details DiscogsReleaseResponse
		var detailErr error
		if detailCtx.Err() == nil {
			details, detailErr = lookupDiscogsRelease(detailCtx, tokenAPI, release.ID)
		} else {
			detailErr = detailCtx.Err()
		}
		if detailErr != nil {
			log.Warn().Err(detailErr).Int("release_id", release.ID).Msg("Discogs release details unavailable; using search result")
			p.Manufacturer = strings.Join(release.Label, ", ")
			p.ModelNumber = release.CatalogNo
			p.Item.Description = "Data provided by Discogs: " + discogsURL
			p.Fields = discogsFallbackFields(release.ID, iEan, release.Year, release.Country, release.Format, release.Genre, release.Style, discogsURL)
		} else {
			p.Manufacturer, p.ModelNumber = discogsLabelAndCatalog(details)
			p.Item.Description = discogsTracklistDescription(details.ID, discogsURL)
			p.Fields = discogsReleaseFields(details, iEan, discogsURL)
		}
		products = append(products, p)
	}

	return products, nil
}

func lookupDiscogsRelease(ctx context.Context, tokenAPI string, releaseID int) (DiscogsReleaseResponse, error) {
	endpoint := "https://api.discogs.com/releases/" + fmt.Sprint(releaseID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DiscogsReleaseResponse{}, err
	}
	req.Header.Set("Authorization", "Discogs token="+sanitizeHeader(strings.TrimSpace(tokenAPI)))
	req.Header.Set("User-Agent", "Homebox-Discogs-Barcode/1.0 (+https://github.com/sysadminsmedia/homebox)")

	resp, err := barcodeHTTPClient.Do(req)
	if err != nil {
		return DiscogsReleaseResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return DiscogsReleaseResponse{}, fmt.Errorf("discogs release API returned status code: %d", resp.StatusCode)
	}

	var result DiscogsReleaseResponse
	if err := decodeBarcodeResponse(resp.Body, &result); err != nil {
		return DiscogsReleaseResponse{}, err
	}
	return result, nil
}

func discogsTextField(name, value string) repo.EntityFieldData {
	value = strings.TrimSpace(value)
	value = truncateUTF8(value, 500)
	return repo.EntityFieldData{Type: "text", Name: name, TextValue: value}
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	const suffix = "..."
	limit := maxBytes - len(suffix)
	for limit > 0 && !utf8.ValidString(value[:limit]) {
		limit--
	}
	return value[:limit] + suffix
}

func appendDiscogsField(fields []repo.EntityFieldData, name, value string) []repo.EntityFieldData {
	if strings.TrimSpace(value) == "" {
		return fields
	}
	return append(fields, discogsTextField(name, value))
}

func discogsFallbackFields(releaseID int, barcode, year, country string, format, genre, style []string, discogsURL string) []repo.EntityFieldData {
	fields := make([]repo.EntityFieldData, 0, 8)
	fields = appendDiscogsField(fields, "Discogs Release ID", fmt.Sprint(releaseID))
	fields = appendDiscogsField(fields, "Discogs URL", discogsURL)
	fields = appendDiscogsField(fields, "Release Year", year)
	fields = appendDiscogsField(fields, "Country", country)
	fields = appendDiscogsField(fields, "Format", strings.Join(format, "; "))
	fields = appendDiscogsField(fields, "Genre", strings.Join(genre, "; "))
	fields = appendDiscogsField(fields, "Style", strings.Join(style, "; "))
	fields = appendDiscogsField(fields, "Barcode", barcode)
	return fields
}

func discogsLabelAndCatalog(release DiscogsReleaseResponse) (string, string) {
	labels := make([]string, 0, len(release.Labels))
	catalog := ""
	for _, label := range release.Labels {
		if strings.TrimSpace(label.Name) != "" && !containsString(labels, label.Name) {
			labels = append(labels, label.Name)
		}
		if catalog == "" {
			catalog = strings.TrimSpace(label.CatNo)
		}
	}
	return strings.Join(labels, ", "), catalog
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func discogsFormat(release DiscogsReleaseResponse) string {
	parts := make([]string, 0, len(release.Formats))
	for _, format := range release.Formats {
		values := make([]string, 0, 3+len(format.Descriptions))
		if strings.TrimSpace(format.Qty) != "" {
			values = append(values, format.Qty+" x")
		}
		values = append(values, format.Name)
		values = append(values, format.Descriptions...)
		if strings.TrimSpace(format.Text) != "" {
			values = append(values, format.Text)
		}
		parts = append(parts, strings.Join(values, ", "))
	}
	return strings.Join(parts, "; ")
}

func discogsReleaseFields(release DiscogsReleaseResponse, barcode, discogsURL string) []repo.EntityFieldData {
	fields := make([]repo.EntityFieldData, 0, 16+len(release.Identifiers))
	fields = appendDiscogsField(fields, "Discogs Release ID", fmt.Sprint(release.ID))
	if release.MasterID != 0 {
		fields = appendDiscogsField(fields, "Discogs Master ID", fmt.Sprint(release.MasterID))
	}
	fields = appendDiscogsField(fields, "Discogs URL", discogsURL)
	fields = appendDiscogsField(fields, "Release Date", release.Released)
	fields = appendDiscogsField(fields, "Country", release.Country)
	fields = appendDiscogsField(fields, "Format", discogsFormat(release))
	fields = appendDiscogsField(fields, "Genre", strings.Join(release.Genres, "; "))
	fields = appendDiscogsField(fields, "Style", strings.Join(release.Styles, "; "))
	fields = appendDiscogsField(fields, "Barcode", barcode)

	for _, company := range release.Companies {
		if company.EntityTypeName == "Pressed By" {
			fields = appendDiscogsField(fields, "Pressed By", company.Name)
		}
	}
	for _, identifier := range release.Identifiers {
		name := identifier.Type
		if identifier.Type == "Barcode" {
			continue
		}
		if strings.TrimSpace(identifier.Description) != "" {
			name += " (" + identifier.Description + ")"
		}
		fields = appendDiscogsField(fields, name, identifier.Value)
	}
	fields = appendDiscogsField(fields, "Tracklist", discogsTracklistText(release.Tracklist))
	fields = appendDiscogsField(fields, "Edition Notes", release.Notes)
	return fields
}

func discogsTracklistText(tracklist []struct {
	Position string `json:"position"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
	Type     string `json:"type_"`
}) string {
	tracks := make([]string, 0, len(tracklist))
	for _, track := range tracklist {
		if track.Type != "track" || strings.TrimSpace(track.Title) == "" {
			continue
		}
		entry := strings.TrimSpace(strings.Join([]string{track.Position, track.Title, track.Duration}, " — "))
		tracks = append(tracks, strings.ReplaceAll(entry, " —  — ", " — "))
	}
	return strings.Join(tracks, "; ")
}

func discogsTracklistDescription(releaseID int, discogsURL string) string {
	return fmt.Sprintf("[Discogs release %d](%s)", releaseID, discogsURL)
}

// sanitizeHeader removes control characters that could cause HTTP header injection.
func sanitizeHeader(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, s)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

var openFactsAllowedImageDomains = []string{
	"openfoodfacts.org",
	"openbeautyfacts.org",
	"openproductsfacts.org",
}

func isAllowedOpenFactsImageHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))

	for _, domain := range openFactsAllowedImageDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

func normalizeOpenFactsImageURL(imageURL string) string {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return ""
	}

	u, err := url.Parse(imageURL)
	if err != nil || u.Hostname() == "" || u.User != nil {
		return ""
	}

	switch u.Scheme {
	case "http":
		u.Scheme = schemeHTTPS
	case schemeHTTPS:
	default:
		return ""
	}

	if !isAllowedOpenFactsImageHost(u.Hostname()) {
		return ""
	}

	return u.String()
}

func buildOpenFactsBarcodeProduct(sourceName string, iEan string, product openFactsProduct) (repo.BarcodeProduct, bool) {
	name := firstNonEmpty(product.ProductName, product.GenericName, product.Brands)
	if name == "" {
		return repo.BarcodeProduct{}, false
	}

	var p repo.BarcodeProduct
	p.Barcode = iEan
	p.SearchEngineName = sourceName
	p.Item.Name = name
	p.Manufacturer = product.Brands

	descriptionSources := [...]string{product.GenericName, product.Categories, product.Quantity}
	descriptionParts := make([]string, 0, len(descriptionSources))
	for _, value := range descriptionSources {
		value = strings.TrimSpace(value)
		if value != "" && value != name {
			descriptionParts = append(descriptionParts, value)
		}
	}
	p.Item.Description = strings.Join(descriptionParts, " | ")

	p.ImageURL = normalizeOpenFactsImageURL(firstNonEmpty(product.ImageFrontURL, product.ImageURL))

	return p, true
}

func lookupOpenFacts(contact string, source openFactsSource, iEan string) ([]repo.BarcodeProduct, error) {
	req, err := http.NewRequest(
		"GET", strings.TrimRight(source.BaseURL, "/")+"/api/v2/product/"+url.PathEscape(iEan)+".json", nil)
	if err != nil {
		return nil, err
	}
	userAgent := "Homebox/1.0 (https://github.com/sysadminsmedia/homebox)"
	safeContact := sanitizeHeader(strings.TrimSpace(contact))
	if len(safeContact) > 0 {
		userAgent = "Homebox/1.0 (contact: " + safeContact + ")"
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := barcodeHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s API returned status code: %d", source.Name, resp.StatusCode)
	}

	var result OpenFactsResponse
	if err := decodeBarcodeResponse(resp.Body, &result); err != nil {
		log.Error().Msg("Can not unmarshal " + source.Name + " JSON")
		return nil, err
	}

	if result.Status == 0 {
		return nil, nil
	}

	p, ok := buildOpenFactsBarcodeProduct(source.Name, iEan, result.Product)
	if !ok {
		return nil, nil
	}

	return []repo.BarcodeProduct{p}, nil
}

// fetchImageBase64 fetches an image from the given HTTPS URL and returns it as a base64-encoded data URI.
func fetchImageBase64(imageURL string) (string, error) {
	res, err := barcodeHTTPClient.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image fetch returned status %d", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("non-image content type: %s", contentType)
	}

	limitedReader := io.LimitReader(res.Body, 8*1024*1024)
	raw, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(raw)
	var prefix string
	switch mimeType {
	case "image/jpeg":
		prefix = "data:image/jpeg;base64,"
	case "image/png":
		prefix = "data:image/png;base64,"
	default:
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	// Encode straight into a single buffer sized for prefix+payload so the data URI
	// is built without an intermediate base64 string and a second concat.
	out := make([]byte, len(prefix)+base64.StdEncoding.EncodedLen(len(raw)))
	copy(out, prefix)
	base64.StdEncoding.Encode(out[len(prefix):], raw)

	return string(out), nil
}

// HandleProductSearchFromBarcode godoc
//
//	@Summary	Search EAN from Barcode
//	@Tags		Items
//	@Produce	json
//	@Param		data	query		string	false	"barcode to be searched"
//	@Success	200		{object}	[]repo.BarcodeProduct
//	@Router		/v1/products/search-from-barcode [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleProductSearchFromBarcode(conf config.BarcodeAPIConf) errchain.HandlerFunc {
	type query struct {
		// 80 characters is the longest non-2D barcode length (GS1-128)
		EAN string `schema:"productEAN" validate:"required,max=80"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		q, err := adapters.DecodeQuery[query](r)
		if err != nil {
			return err
		}

		log.Info().Msg("Processing barcode lookup request on: " + q.EAN)
		// A provider timeout must not consume the server's write deadline and
		// prevent already available results from reaching the client.
		allowSlowResponse(w, r)

		// www.ean-search.org/: not free

		// Example code: dewalt 5035048748428
		type lookup struct {
			name string
			run  func() ([]repo.BarcodeProduct, error)
		}
		lookups := []lookup{{
			name: "upcitemdb.com",
			run:  func() ([]repo.BarcodeProduct, error) { return lookupUPCItemDB(q.EAN) },
		}}

		if conf.TokenBarcodespider != "" {
			lookups = append(lookups, lookup{
				name: "barcodespider.com",
				run: func() ([]repo.BarcodeProduct, error) {
					return lookupBarcodespider(conf.TokenBarcodespider, q.EAN)
				},
			})
		}

		if conf.TokenDiscogs != "" {
			lookups = append(lookups, lookup{
				name: "discogs.com",
				run: func() ([]repo.BarcodeProduct, error) {
					return lookupDiscogs(r.Context(), conf.TokenDiscogs, q.EAN)
				},
			})
		}

		for _, source := range openFactsSources {
			lookups = append(lookups, lookup{
				name: source.Name,
				run: func() ([]repo.BarcodeProduct, error) {
					return lookupOpenFacts(conf.OpenFoodFactsContact, source, q.EAN)
				},
			})
		}

		results := make([][]repo.BarcodeProduct, len(lookups))
		var wg sync.WaitGroup
		for i, lookup := range lookups {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ps, err := lookup.run()
				if err != nil {
					log.Error().Msg("Can not retrieve product from " + lookup.name + ": " + err.Error())
				}
				results[i] = ps
			}()
		}
		wg.Wait()

		products := make([]repo.BarcodeProduct, 0, len(results))
		for _, result := range results {
			products = append(products, result...)
		}

		// Retrieve images if possible
		for i := range products {
			p := &products[i]

			if len(p.ImageURL) == 0 {
				continue
			}

			// Validate URL is HTTPS
			u, err := url.Parse(p.ImageURL)
			if err != nil || u.Scheme != schemeHTTPS {
				log.Warn().Msg("Skipping non-HTTPS image URL: " + p.ImageURL)
				continue
			}

			base64Img, err := fetchImageBase64(p.ImageURL)
			if err != nil {
				log.Warn().Msg("Cannot fetch image for URL: " + p.ImageURL + ": " + err.Error())
				continue
			}
			p.ImageBase64 = base64Img
		}

		if len(products) != 0 {
			return server.JSON(w, http.StatusOK, products)
		}

		return server.JSON(w, http.StatusNoContent, nil)
	}
}
