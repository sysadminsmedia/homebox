package v1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
)

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
			Merchant     string  `json:"merchant"`
			Domain       string  `json:"domain"`
			Title        string  `json:"title"`
			Currency     string  `json:"currency"`
			ListPrice    string  `json:"list_price"`
			Price        float64 `json:"price"`
			Shipping     string  `json:"shipping"`
			Condition    string  `json:"condition"`
			Availability string  `json:"availability"`
			Link         string  `json:"link"`
			UpdatedT     int     `json:"updated_t"`
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

func lookupUPCItemDB(iEan string) ([]repo.BarcodeProduct, error) {
	client := &http.Client{Timeout: barcodeHTTPTimeoutSec * time.Second}
	resp, err := client.Get("https://api.upcitemdb.com/prod/trial/lookup?upc=" + url.QueryEscape(iEan))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result UPCITEMDBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error().Msg("Can not unmarshal JSON from upcitemdb.com")
		return nil, err
	}

	var res []repo.BarcodeProduct

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

	client := &http.Client{Timeout: barcodeHTTPTimeoutSec * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("barcodespider API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result BARCODESPIDER_COMResponse
	if err := json.Unmarshal(body, &result); err != nil {
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

func isAllowedOpenFactsImageHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	allowedDomains := []string{
		"openfoodfacts.org",
		"openbeautyfacts.org",
		"openproductsfacts.org",
	}

	for _, domain := range allowedDomains {
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

	var descriptionParts []string
	for _, value := range []string{product.GenericName, product.Categories, product.Quantity} {
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
	client := &http.Client{Timeout: barcodeHTTPTimeoutSec * time.Second}
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s API returned status code: %d", source.Name, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result OpenFactsResponse
	if err := json.Unmarshal(body, &result); err != nil {
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
	client := &http.Client{Timeout: barcodeHTTPTimeoutSec * time.Second}
	res, err := client.Get(imageURL)
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
	bytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(bytes)
	var base64Encoding string
	switch mimeType {
	case "image/jpeg":
		base64Encoding = "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding = "data:image/png;base64,"
	default:
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	return base64Encoding + base64.StdEncoding.EncodeToString(bytes), nil
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

		var products []repo.BarcodeProduct

		// www.ean-search.org/: not free

		// Example code: dewalt 5035048748428

		ps, err := lookupUPCItemDB(q.EAN)
		if err != nil {
			log.Error().Msg("Can not retrieve product from upcitemdb.com: " + err.Error())
		}
		products = append(products, ps...)

		if conf.TokenBarcodespider != "" {
			ps2, err := lookupBarcodespider(conf.TokenBarcodespider, q.EAN)
			if err != nil {
				log.Error().Msg("Can not retrieve product from barcodespider.com: " + err.Error())
			}
			products = append(products, ps2...)
		}

		for _, source := range openFactsSources {
			ps3, err := lookupOpenFacts(conf.OpenFoodFactsContact, source, q.EAN)
			if err != nil {
				log.Error().Msg("Can not retrieve product from " + source.Name + ": " + err.Error())
			}
			products = append(products, ps3...)
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
