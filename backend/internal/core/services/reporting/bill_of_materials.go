package reporting

import (
	"github.com/gocarina/gocsv"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// =================================================================================================

type BillOfMaterialsEntry struct {
	PurchaseDate types.Date `csv:"Purchase Date"`
	Name         string     `csv:"Name"`
	Description  string     `csv:"Description"`
	Manufacturer string     `csv:"Manufacturer"`
	SerialNumber string     `csv:"Serial Number"`
	ModelNumber  string     `csv:"Model Number"`
	Quantity     int        `csv:"Quantity"`
	Price        float64    `csv:"Price"`
	TotalPrice   float64    `csv:"Total Price"`
}

// BillOfMaterialsCSV returns a byte slice of the Bill of Materials for a given GID in CSV format
// See BillOfMaterialsEntry for the format of the output
func BillOfMaterialsCSV(entities []repo.ItemOut) ([]byte, error) {
	bomEntries := lo.Map(entities, func(entity repo.ItemOut, _ int) BillOfMaterialsEntry {
		return BillOfMaterialsEntry{
			PurchaseDate: entity.PurchaseTime,
			Name:         entity.Name,
			Description:  entity.Description,
			Manufacturer: entity.Manufacturer,
			SerialNumber: entity.SerialNumber,
			ModelNumber:  entity.ModelNumber,
			Quantity:     entity.Quantity,
			Price:        entity.PurchasePrice,
			TotalPrice:   entity.PurchasePrice * float64(entity.Quantity),
		}
	})

	return gocsv.MarshalBytes(&bomEntries)
}
