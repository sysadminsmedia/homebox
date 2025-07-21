package repo

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Helper() // mark as helper so failure line points to caller
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBarcode_SearchProducts(t *testing.T) {
	gtins := []string{
		"855800001869",  // Das Keyboard
		"5035048748428", // Dewalt multitool
		"885911209809",  // Dewalt drill
	}

	for _, b := range gtins {
		products, err := tRepos.Barcode.RetrieveProductsFromBarcode(config.BarcodeAPIConf{}, b)

		checkError(t, err)

		for _, p := range products {
			log.Info().Msg("Found products: " + p.Item.Name)
		}

		// Sleep 1 sec to avoid API DoS
		time.Sleep(4 * time.Second)
	}
}
