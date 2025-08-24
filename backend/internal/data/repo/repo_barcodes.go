package repo

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

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

type (
	BarcodeRepository struct {
	}

	BarcodeProduct struct {
		SearchEngineName       string `json:"search_engine_name"`
		SearchEngineURL        string `json:"search_engine_url"`
		SearchEngineProductURL string `json:"search_engine_product_url"`

		// Extras
		Country string `json:"notes"`

		ImageURL    string `json:"imageURL"`
		ImageBase64 string `json:"imageBase64"`

		Item ItemCreate `json:"item"`
	}

	ProductDatabaseFunc func(config.BarcodeAPIConf, string) ([]BarcodeProduct, error)

	ProductDatabaseImpl struct {
		url  string
		name string
		call ProductDatabaseFunc
	}

	UPCITEMDBResponse struct {
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

	BARCODESPIDER_COMResponse struct {
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
)

const FOREIGN_API_CALL_TIMEOUT_SEC = 10

func (r *BarcodeRepository) UPCItemDB_Search(_ config.BarcodeAPIConf, iBarcode string) ([]BarcodeProduct, error) {
	client := &http.Client{Timeout: FOREIGN_API_CALL_TIMEOUT_SEC * time.Second}
	resp, err := client.Get("https://api.upcitemdb.com/prod/trial/lookup?upc=" + iBarcode)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Uncomment the following string for debug
	// sb := string(body)
	// log.Debug().Msg("Response: " + sb)

	var result UPCITEMDBResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Error().Msg("Can not unmarshal JSON")
	}

	var res []BarcodeProduct

	for _, it := range result.Items {
		var p BarcodeProduct
		p.Item.Barcode = iBarcode
		p.SearchEngineProductURL = "https://www.upcitemdb.com/upc/" + iBarcode

		p.Item.Description = it.Description
		p.Item.Name = it.Title
		p.Item.Manufacturer = it.Brand
		p.Item.ModelNumber = it.Model
		if len(it.Images) != 0 {
			p.ImageURL = it.Images[0]
		}

		res = append(res, p)
	}

	return res, nil
}

func (r *BarcodeRepository) BarcodeSpider_Search(conf config.BarcodeAPIConf, iBarcode string) ([]BarcodeProduct, error) {
	if len(conf.TokenBarcodespider) == 0 {
		return nil, errors.New("no api token configured for barcodespider. " +
			"Please define the api token in environment variable HBOX_BARCODE_TOKEN_BARCODESPIDER")
	}

	req, err := http.NewRequest(
		"GET", "https://api.barcodespider.com/v1/lookup?upc="+url.QueryEscape(iBarcode), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("token", conf.TokenBarcodespider)

	client := &http.Client{Timeout: FOREIGN_API_CALL_TIMEOUT_SEC * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// defer the call to Body.Close(). We also check the error code, and merge
	// it with the other error in this code to avoid error overiding.
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("barcodespider API returned status code: %d", resp.StatusCode)
	}

	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Uncomment the following string for debug
	// sb := string(body)
	// log.Debug().Msg("Response: " + sb)

	var result BARCODESPIDER_COMResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Error().Msg("Can not unmarshal JSON")
	}

	// TODO: check 200 code on HTTP response.
	var p BarcodeProduct
	p.Item.Barcode = iBarcode
	p.SearchEngineProductURL = "https://amp.barcodespider.com/" + iBarcode
	p.Item.Name = result.ItemAttributes.Title
	p.Item.Description = result.ItemAttributes.Description
	p.Item.Manufacturer = result.ItemAttributes.Brand
	p.Item.ModelNumber = result.ItemAttributes.Model
	p.ImageURL = result.ItemAttributes.Image

	var res []BarcodeProduct
	res = append(res, p)

	return res, nil
}

func (r *BarcodeRepository) UpdateProductWithImage(iProduct *BarcodeProduct) {
	p := iProduct

	if len(p.ImageURL) == 0 {
		return
	}

	// Validate URL is HTTPS
	u, err := url.Parse(p.ImageURL)
	if err != nil || u.Scheme != "https" {
		log.Warn().Msg("Skipping non-HTTPS image URL: " + p.ImageURL)
		return
	}

	client := &http.Client{Timeout: FOREIGN_API_CALL_TIMEOUT_SEC * time.Second}
	res, err := client.Get(p.ImageURL)
	if err != nil {
		log.Warn().Msg("Cannot fetch image for URL: " + p.ImageURL + ": " + err.Error())
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	// Validate response
	if res.StatusCode != http.StatusOK {
		return
	}

	// Check content type
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return
	}

	// Limit image size to 8MB
	limitedReader := io.LimitReader(res.Body, 8*1024*1024)

	// Read data of image
	bytes, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Warn().Msg(err.Error())
		return
	}

	// Convert to Base64
	var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	default:
		return
	}

	// Append the base64 encoded output
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	p.ImageBase64 = base64Encoding
}

func (r *BarcodeRepository) RetrieveProductsFromBarcode(conf config.BarcodeAPIConf, iBarcode string) ([]BarcodeProduct, error) {
	log.Info().Msg("Processing barcode lookup request on: " + iBarcode)

	// For further implementer: we try to not use non-free databases
	// - www.ean-search.org/: not free
	// - barcodelookup.com/: trial with 50 items search / months. Need phone number for registration.

	remoteAPIs := []ProductDatabaseImpl{
		{
			url:  "https://upcitemdb.com",
			name: "UPCDBItem",
			call: r.UPCItemDB_Search, // Assign function 1
		},
		{
			url:  "https://barcodespider.com",
			name: "BarcodeSpider",
			call: r.BarcodeSpider_Search, // Assign function 2
		},
	}

	var products []BarcodeProduct

	// Call external APIs
	for _, api := range remoteAPIs {
		ps, err := api.call(conf, iBarcode)
		if err != nil {
			log.Error().Msg("Can not retrieve product from " + api.name + err.Error())
		}

		for idx := range ps {
			// Update each product with API information
			p := &ps[idx]
			p.SearchEngineName = api.name
			p.SearchEngineURL = api.url

			// Fetch image of each product on the Internet
			r.UpdateProductWithImage(p)
		}

		// Merge found products.
		products = append(products, ps...)
	}

	return products, nil
}
