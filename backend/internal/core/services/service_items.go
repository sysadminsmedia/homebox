package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
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

func (svc *ItemService) Duplicate(ctx Context, gid, id uuid.UUID, options repo.DuplicateOptions) (repo.ItemOut, error) {
	return svc.repo.Items.Duplicate(ctx, gid, id, options)
}

func (svc *ItemService) EnsureAssetID(ctx context.Context, gid uuid.UUID) (int, error) {
	items, err := svc.repo.Items.GetAllZeroAssetID(ctx, gid)
	if err != nil {
		return 0, err
	}

	highest, err := svc.repo.Items.GetHighestAssetID(ctx, gid)
	if err != nil {
		return 0, err
	}

	finished := 0
	for _, item := range items {
		highest++

		err = svc.repo.Items.SetAssetID(ctx, gid, item.ID, highest)
		if err != nil {
			return 0, err
		}

		finished++
	}

	return finished, nil
}

func (svc *ItemService) EnsureImportRef(ctx context.Context, gid uuid.UUID) (int, error) {
	ids, err := svc.repo.Items.GetAllZeroImportRef(ctx, gid)
	if err != nil {
		return 0, err
	}

	finished := 0
	for _, itemID := range ids {
		ref := uuid.New().String()[0:8]

		err = svc.repo.Items.Patch(ctx, gid, itemID, repo.ItemPatch{ImportRef: &ref})
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
//  3. Locations and Tags are created if they do not exist.
func (svc *ItemService) CsvImport(ctx context.Context, gid uuid.UUID, data io.Reader) (int, error) {
	sheet := reporting.IOSheet{}

	err := sheet.Read(data)
	if err != nil {
		return 0, err
	}

	// ========================================
	// Tags

	var tagMap map[string]uuid.UUID
	{
		tags, err := svc.repo.Tags.GetAll(ctx, gid)
		if err != nil {
			return 0, err
		}

		tagMap = lo.SliceToMap(tags, func(tag repo.TagSummary) (string, uuid.UUID) {
			return tag.Name, tag.ID
		})
	}

	// ========================================
	// Locations

	locationMap := make(map[string]uuid.UUID)
	{
		locations, err := svc.repo.Locations.Tree(ctx, gid, repo.TreeQuery{WithItems: false})
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
		highestAID, err = svc.repo.Items.GetHighestAssetID(ctx, gid)
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
			exists, err := svc.repo.Items.CheckRef(ctx, gid, row.ImportRef)
			if err != nil {
				return 0, fmt.Errorf("error checking for existing item with ref %q: %w", row.ImportRef, err)
			}

			if exists {
				createRequired = false
			}
		}

		// ========================================
		// Pre-Create tags as necessary
		tagIds := make([]uuid.UUID, len(row.TagStr))

		for j := range row.TagStr {
			tag := row.TagStr[j]

			id, ok := tagMap[tag]
			if !ok {
				newTag, err := svc.repo.Tags.Create(ctx, gid, repo.TagCreate{Name: tag})
				if err != nil {
					return 0, err
				}
				id = newTag.ID
			}

			tagIds[j] = id
			tagMap[tag] = id
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

					newLocation, err := svc.repo.Locations.Create(ctx, gid, repo.LocationCreate{
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
				TagIDs:      tagIds,
			}

			item, err = svc.repo.Items.Create(ctx, gid, newItem)
			if err != nil {
				return 0, err
			}
		default:
			item, err = svc.repo.Items.GetByRef(ctx, gid, row.ImportRef)
			if err != nil {
				return 0, err
			}
		}

		if item.ID == uuid.Nil {
			panic("item ID is nil on import - this should never happen")
		}

		fields := lo.Map(row.Fields, func(f reporting.ExportItemFields, _ int) repo.ItemField {
			return repo.ItemField{
				Name:      f.Name,
				Type:      "text",
				TextValue: f.Value,
			}
		})

		updateItem := repo.ItemUpdate{
			ID:         item.ID,
			TagIDs:     tagIds,
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

		item, err = svc.repo.Items.UpdateByGroup(ctx, gid, updateItem)
		if err != nil {
			return 0, err
		}

		finished++
	}

	return finished, nil
}

func (svc *ItemService) ExportCSV(ctx context.Context, gid uuid.UUID, hbURL string) ([][]string, error) {
	items, err := svc.repo.Items.GetAll(ctx, gid)
	if err != nil {
		return nil, err
	}

	sheet := reporting.IOSheet{}

	err = sheet.ReadItems(ctx, items, gid, svc.repo, hbURL)
	if err != nil {
		return nil, err
	}

	return sheet.CSV()
}

func (svc *ItemService) ExportBillOfMaterialsCSV(ctx context.Context, gid uuid.UUID) ([]byte, error) {
	items, err := svc.repo.Items.GetAll(ctx, gid)
	if err != nil {
		return nil, err
	}

	return reporting.BillOfMaterialsCSV(items)
}
