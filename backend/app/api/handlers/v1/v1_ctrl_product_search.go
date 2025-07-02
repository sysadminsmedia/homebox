package v1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
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

// HandleGenerateQRCode godoc
//
//	@Summary	Search EAN from Barcode
//	@Tags		Items
//	@Produce	json
//	@Param		data	query		string	false	"barcode to be searched"
//	@Success	200		{object}	[]repo.BarcodeProduct
//	@Router		/v1/getproductfromean [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleProductSearchFromEAN(conf config.BarcodeAPIConf) errchain.HandlerFunc {
	type query struct {
		// 4,296 characters is the maximum length of a QR code
		EAN string `schema:"productEAN" validate:"required,max=4296"`
	}

	/*fn := func(r *http.Request, ID uuid.UUID) (repo.ItemOut, error) {
		auth := services.NewContext(r.Context())

		return ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)*/

	return func(w http.ResponseWriter, r *http.Request) error {
		q, err := adapters.DecodeQuery[query](r)
		if err != nil {
			return err
		}

		log.Info().Msg("========================" + q.EAN)

		// Search on UPCITEMDB
		var products []repo.BarcodeProduct

		// www.ean-search.org/: not free

		// Example code: dewalt 5035048748428

		upcitemdb := func(iEan string) ([]repo.BarcodeProduct, error) {
			resp, err := http.Get("https://api.upcitemdb.com/prod/trial/lookup?upc=" + iEan)
			if err != nil {
				return nil, err
			}

			// We Read the response body on the line below.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			// Convert the body to type string
			sb := string(body)
			log.Info().Msg("Response: " + sb)

			var result UPCITEMDBResponse
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				log.Error().Msg("Can not unmarshal JSON")
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

		ps, err := upcitemdb(q.EAN)
		if err != nil {
			log.Error().Msg("Can not retrieve product from upcitemdb.com" + err.Error())
		}

		// Barcode spider implementation
		barcodespider := func(tokenAPI string, iEan string) ([]repo.BarcodeProduct, error) {
			if len(tokenAPI) == 0 {
				return nil, errors.New("no api token configured for barcodespider")
			}

			req, err := http.NewRequest(
				"GET", "https://api.barcodespider.com/v1/lookup?upc="+iEan, nil)

			if err != nil {
				return nil, err
			}

			req.Header.Add("token", tokenAPI)

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			// We Read the response body on the line below.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			// Convert the body to type string
			sb := string(body)
			log.Info().Msg("Response: " + sb)

			var result BARCODESPIDER_COMResponse
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				log.Error().Msg("Can not unmarshal JSON")
			}

			// TODO: check 200 code on HTTP response.
			var p repo.BarcodeProduct
			p.Barcode = iEan
			p.SearchEngineName = "barcodespider.com"
			p.Item.Name = result.ItemAttributes.Title
			p.Item.Description = result.ItemAttributes.Description
			p.Manufacturer = result.ItemAttributes.Brand
			p.ModelNumber = result.ItemAttributes.Model
			p.ImageURL = result.ItemAttributes.Image

			var res []repo.BarcodeProduct
			res = append(res, p)

			return res, nil
		}

		ps2, err := barcodespider(conf.TokenBarcodespider, q.EAN)
		if err != nil {
			log.Error().Msg("Can not retrieve product from barcodespider.com: " + err.Error())
		}

		// Merge everything.
		products = append(products, ps...)

		products = append(products, ps2...)

		// Retrieve images if possible
		for i := range products {
			p := &products[i]

			if len(p.ImageURL) == 0 {
				continue
			}

			res, err := http.Get(p.ImageURL)
			if err != nil {
				log.Warn().Msg("Cannot fetch image for URL: " + p.ImageURL + ": " + err.Error())
			}

			defer res.Body.Close()

			// Read data of image
			bytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Warn().Msg(err.Error())
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
			}

			// Append the base64 encoded output
			base64Encoding += base64.StdEncoding.EncodeToString(bytes)

			p.ImageBase64 = base64Encoding
		}

		w.Header().Set("Content-Type", "application/json")

		if len(products) != 0 {
			// Return only the first result for now. Enhance this with a dedicated dialog
			// displaying all the references found?
			return json.NewEncoder(w).Encode(products)
		}

		return nil
	}
}
