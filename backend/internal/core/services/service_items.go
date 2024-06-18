package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrFileNotFound = errors.New("file not found")
)

type ItemService struct {
	repo *repo.AllRepos

	filepath string

	autoIncrementAssetID bool
}

func (svc *ItemService) Create(ctx Context, item repo.ItemCreate) (repo.ItemOut, error) {
	if svc.autoIncrementAssetID {
		highest, err := svc.repo.Items.GetHighestAssetID(ctx, ctx.GID)
		if err != nil {
			return repo.ItemOut{}, err
		}

		item.AssetID = highest + 1
	}

	return svc.repo.Items.Create(ctx, ctx.GID, item)
}

func (svc *ItemService) EnsureAssetID(ctx context.Context, GID uuid.UUID) (int, error) {
	items, err := svc.repo.Items.GetAllZeroAssetID(ctx, GID)
	if err != nil {
		return 0, err
	}

	highest, err := svc.repo.Items.GetHighestAssetID(ctx, GID)
	if err != nil {
		return 0, err
	}

	finished := 0
	for _, item := range items {
		highest++

		err = svc.repo.Items.SetAssetID(ctx, GID, item.ID, highest)
		if err != nil {
			return 0, err
		}

		finished++
	}

	return finished, nil
}

func (svc *ItemService) EnsureImportRef(ctx context.Context, GID uuid.UUID) (int, error) {
	ids, err := svc.repo.Items.GetAllZeroImportRef(ctx, GID)
	if err != nil {
		return 0, err
	}

	finished := 0
	for _, itemID := range ids {
		ref := uuid.New().String()[0:8]

		err = svc.repo.Items.Patch(ctx, GID, itemID, repo.ItemPatch{ImportRef: &ref})
		if err != nil {
			return 0, err
		}

		finished++
	}

	return finished, nil
}

func serializeLocation[T ~[]string](location T) string {
	return strings.Join(location, "/")
}

// CsvImport imports items from a CSV file. using the standard defined format.
//
// CsvImport applies the following rules/operations
//
//  1. If the item does not exist, it is created.
//  2. If the item has a ImportRef and it exists it is skipped
//  3. Locations and Labels are created if they do not exist.
func (svc *ItemService) CsvImport(ctx context.Context, GID uuid.UUID, data io.Reader) (int, error) {
	sheet := reporting.IOSheet{}

	err := sheet.Read(data)
	if err != nil {
		return 0, err
	}

	// ========================================
	// Labels

	labelMap := make(map[string]uuid.UUID)
	{
		labels, err := svc.repo.Labels.GetAll(ctx, GID)
		if err != nil {
			return 0, err
		}

		for _, label := range labels {
			labelMap[label.Name] = label.ID
		}
	}

	// ========================================
	// Locations

	locationMap := make(map[string]uuid.UUID)
	{
		locations, err := svc.repo.Locations.Tree(ctx, GID, repo.TreeQuery{WithItems: false})
		if err != nil {
			return 0, err
		}

		// Traverse the tree and build a map of location full paths to IDs
		// where the full path is the location name joined by slashes.
		var traverse func(location *repo.TreeItem, path []string)
		traverse = func(location *repo.TreeItem, path []string) {
			path = append(path, location.Name)

			locationMap[serializeLocation(path)] = location.ID

			for _, child := range location.Children {
				traverse(child, path)
			}
		}

		for _, location := range locations {
			traverse(&location, []string{})
		}
	}

	// ========================================
	// Import items

	// Asset ID Pre-Check
	highestAID := repo.AssetID(-1)
	if svc.autoIncrementAssetID {
		highestAID, err = svc.repo.Items.GetHighestAssetID(ctx, GID)
		if err != nil {
			return 0, err
		}
	}

	finished := 0

	for i := range sheet.Rows {
		row := sheet.Rows[i]

		createRequired := true

		// ========================================
		// Preflight check for existing item
		if row.ImportRef != "" {
			exists, err := svc.repo.Items.CheckRef(ctx, GID, row.ImportRef)
			if err != nil {
				return 0, fmt.Errorf("error checking for existing item with ref %q: %w", row.ImportRef, err)
			}

			if exists {
				createRequired = false
			}
		}

		// ========================================
		// Pre-Create Labels as necessary
		labelIds := make([]uuid.UUID, len(row.LabelStr))

		for j := range row.LabelStr {
			label := row.LabelStr[j]

			id, ok := labelMap[label]
			if !ok {
				newLabel, err := svc.repo.Labels.Create(ctx, GID, repo.LabelCreate{Name: label})
				if err != nil {
					return 0, err
				}
				id = newLabel.ID
			}

			labelIds[j] = id
			labelMap[label] = id
		}

		// ========================================
		// Pre-Create Locations as necessary
		path := serializeLocation(row.Location)

		locationID, ok := locationMap[path]
		if !ok { // Traverse the path of LocationStr and check each path element to see if it exists already, if not create it.
			paths := []string{}
			for i, pathElement := range row.Location {
				paths = append(paths, pathElement)
				path := serializeLocation(paths)

				locationID, ok = locationMap[path]
				if !ok {
					parentID := uuid.Nil

					// Get the parent ID
					if i > 0 {
						parentPath := serializeLocation(row.Location[:i])
						parentID = locationMap[parentPath]
					}

					newLocation, err := svc.repo.Locations.Create(ctx, GID, repo.LocationCreate{
						ParentID: parentID,
						Name:     pathElement,
					})
					if err != nil {
						return 0, err
					}
					locationID = newLocation.ID
				}

				locationMap[path] = locationID
			}

			locationID, ok = locationMap[path]
			if !ok {
				return 0, errors.New("failed to create location")
			}
		}

		var effAID repo.AssetID
		if svc.autoIncrementAssetID && row.AssetID.Nil() {
			effAID = highestAID + 1
			highestAID++
		} else {
			effAID = row.AssetID
		}

		// ========================================
		// Create Item
		var item repo.ItemOut
		switch {
		case createRequired:
			newItem := repo.ItemCreate{
				ImportRef:   row.ImportRef,
				Name:        row.Name,
				Description: row.Description,
				AssetID:     effAID,
				LocationID:  locationID,
				LabelIDs:    labelIds,
			}

			item, err = svc.repo.Items.Create(ctx, GID, newItem)
			if err != nil {
				return 0, err
			}
		default:
			item, err = svc.repo.Items.GetByRef(ctx, GID, row.ImportRef)
			if err != nil {
				return 0, err
			}
		}

		if item.ID == uuid.Nil {
			panic("item ID is nil on import - this should never happen")
		}

		fields := make([]repo.ItemField, len(row.Fields))
		for i := range row.Fields {
			fields[i] = repo.ItemField{
				Name:      row.Fields[i].Name,
				Type:      "text",
				TextValue: row.Fields[i].Value,
			}
		}

		updateItem := repo.ItemUpdate{
			ID:         item.ID,
			LabelIDs:   labelIds,
			LocationID: locationID,

			Name:        row.Name,
			Description: row.Description,
			AssetID:     effAID,
			Insured:     row.Insured,
			Quantity:    row.Quantity,
			Archived:    row.Archived,

			PurchasePrice: row.PurchasePrice,
			PurchaseFrom:  row.PurchaseFrom,
			PurchaseTime:  row.PurchaseTime,

			Manufacturer: row.Manufacturer,
			ModelNumber:  row.ModelNumber,
			SerialNumber: row.SerialNumber,

			LifetimeWarranty: row.LifetimeWarranty,
			WarrantyExpires:  row.WarrantyExpires,
			WarrantyDetails:  row.WarrantyDetails,

			SoldTo:    row.SoldTo,
			SoldTime:  row.SoldTime,
			SoldPrice: row.SoldPrice,
			SoldNotes: row.SoldNotes,

			Notes:  row.Notes,
			Fields: fields,
		}

		item, err = svc.repo.Items.UpdateByGroup(ctx, GID, updateItem)
		if err != nil {
			return 0, err
		}

		finished++
	}

	return finished, nil
}

func (svc *ItemService) ExportTSV(ctx context.Context, GID uuid.UUID) ([][]string, error) {
	items, err := svc.repo.Items.GetAll(ctx, GID)
	if err != nil {
		return nil, err
	}

	sheet := reporting.IOSheet{}

	err = sheet.ReadItems(ctx, items, GID, svc.repo)
	if err != nil {
		return nil, err
	}

	return sheet.TSV()
}

func (svc *ItemService) ExportBillOfMaterialsTSV(ctx context.Context, GID uuid.UUID) ([]byte, error) {
	items, err := svc.repo.Items.GetAll(ctx, GID)
	if err != nil {
		return nil, err
	}

	return reporting.BillOfMaterialsTSV(items)
}
