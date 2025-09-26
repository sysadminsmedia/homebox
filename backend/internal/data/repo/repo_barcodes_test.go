package repo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Helper() // mark as helper so failure line points to caller
		t.Fatalf("unexpected error: %v", err)
	}
}

func readJSONResponseSamplesFromDisk(t *testing.T, subfolder string) map[string]string {
	entries, err := os.ReadDir(subfolder)
	if err != nil {
		t.Fatal(err)
	}

	res := make(map[string]string)

	for _, e := range entries {
		path := filepath.Join(subfolder, e.Name())

		data, err := os.ReadFile(path)

		if err != nil {
			t.Fatal(err)
		}

		res[e.Name()] = string(data[:])
	}

	return res
}

type (
	ServerHandlerFunction func(w http.ResponseWriter, r *http.Request, productMap map[string]string)

	ProductDatabaseTest struct {
		api_name        string
		barcode_handler ProductDatabaseFunc
		products        map[string]string
		mock_server     *httptest.Server
	}
)

func PrepareTest(t *testing.T, apiName string, barcodeHandler ProductDatabaseFunc, folderContainingJsonsSamples string, handler ServerHandlerFunction) ProductDatabaseTest {
	var p ProductDatabaseTest
	p.api_name = apiName
	p.products = readJSONResponseSamplesFromDisk(t, folderContainingJsonsSamples)
	p.barcode_handler = barcodeHandler

	p.mock_server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, p.products)
	}))

	return p
}

func _TestBarcode_UPCDBItemHandler(w http.ResponseWriter, r *http.Request, productMap map[string]string) {

	upc := r.URL.Query().Get("upc")
	log.Info().Msg(upc)

	t := productMap[upc+".json"]

	// Cover the case where no product is found.
	if t == "" {
		t = productMap["empty.json"]
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(t))
}

func _TestBarcode_BarcodeSpiderHandler(w http.ResponseWriter, r *http.Request, productMap map[string]string) {

	upc := r.URL.Query().Get("upc")
	log.Info().Msg(upc)

	t := productMap[upc+".json"]

	// Cover the case where no product is found.
	if t == "" {
		t = productMap["empty.json"]
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(t))
}

func TestBarcode_SearchProducts(t *testing.T) {

	gtins := []string{
		"855800001203",  // Das Keyboard
		"855800001869",  // Das Keyboard
		"5035048748428", // Dewalt multitool
		"885911209809",  // Dewalt drill
	}

	// For each database, we create a "mock" server to simulate the responses from the API.
	productSearchTests := []ProductDatabaseTest{
		PrepareTest(t, "UPCDBItem", tRepos.Barcode.UPCItemDB_Search, "testdata/upcitemdb", _TestBarcode_UPCDBItemHandler),
		PrepareTest(t, "BarcodeSpider", tRepos.Barcode.BarcodeSpider_Search, "testdata/barcodespider", _TestBarcode_BarcodeSpiderHandler),
	}

	for _, api := range productSearchTests {

		log.Info().Msg("Testing " + api.api_name)

		for _, bc := range gtins {
			products, err := api.barcode_handler(config.BarcodeAPIConf{}, bc, api.mock_server.URL)

			checkError(t, err)

			for _, p := range products {
				log.Info().Msg("Found: " + p.Item.Name)
			}
		}
	}
}
