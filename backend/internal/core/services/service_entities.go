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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func entityServiceTracer() trace.Tracer {
	return otel.Tracer("service")
}

func recordServiceSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

var (
	ErrNotFound     = errors.New("not found")
	ErrFileNotFound = errors.New("file not found")
)

type EntityService struct {
	repo *repo.AllRepos

	filepath string

	autoIncrementAssetID bool
}

func (svc *EntityService) Create(ctx Context, entity repo.EntityCreate) (repo.EntityOut, error) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.EntityService.Create",
		trace.WithAttributes(
			attribute.String("group.id", ctx.GID.String()),
			attribute.String("entity.name", entity.Name),
			attribute.Bool("svc.auto_increment_asset_id", svc.autoIncrementAssetID),
		))
	defer span.End()
	ctx.Context = spanCtx

	if svc.autoIncrementAssetID {
		highest, err := svc.repo.Entities.GetHighestAssetID(ctx, ctx.GID)
		if err != nil {
			recordServiceSpanError(span, err)
			return repo.EntityOut{}, err
		}

		entity.AssetID = highest + 1
		span.SetAttributes(attribute.Int64("entity.asset_id", int64(entity.AssetID)))
	}

	out, err := svc.repo.Entities.Create(ctx, ctx.GID, entity)
	if err != nil {
		recordServiceSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.String("entity.id", out.ID.String()))
	return out, nil
}

func (svc *EntityService) Duplicate(ctx Context, gid, id uuid.UUID, options repo.DuplicateOptions) (repo.EntityOut, error) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.EntityService.Duplicate",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.source_id", id.String()),
			attribute.Bool("options.copy_maintenance", options.CopyMaintenance),
			attribute.Bool("options.copy_attachments", options.CopyAttachments),
			attribute.Bool("options.copy_custom_fields", options.CopyCustomFields),
		))
	defer span.End()
	ctx.Context = spanCtx

	out, err := svc.repo.Entities.Duplicate(ctx, gid, id, options)
	if err != nil {
		recordServiceSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.String("entity.id", out.ID.String()))
	return out, nil
}

func (svc *EntityService) EnsureAssetID(ctx context.Context, gid uuid.UUID) (int, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.EnsureAssetID",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	items, err := svc.repo.Entities.GetAllZeroAssetID(ctx, gid)
	if err != nil {
		recordServiceSpanError(span, err)
		return 0, err
	}
	span.SetAttributes(attribute.Int("entities.zero_asset_id.count", len(items)))

	highest, err := svc.repo.Entities.GetHighestAssetID(ctx, gid)
	if err != nil {
		recordServiceSpanError(span, err)
		return 0, err
	}

	_, updateSpan := entityServiceTracer().Start(ctx, "service.EntityService.EnsureAssetID.update",
		trace.WithAttributes(attribute.Int("entities.count", len(items))))
	defer updateSpan.End()

	finished := 0
	for _, item := range items {
		highest++

		err = svc.repo.Entities.SetAssetID(ctx, gid, item.ID, highest)
		if err != nil {
			recordServiceSpanError(updateSpan, err)
			recordServiceSpanError(span, err)
			return 0, err
		}

		finished++
	}

	updateSpan.SetAttributes(attribute.Int("entities.updated.count", finished))
	span.SetAttributes(attribute.Int("entities.updated.count", finished))
	return finished, nil
}

func (svc *EntityService) EnsureImportRef(ctx context.Context, gid uuid.UUID) (int, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.EnsureImportRef",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	ids, err := svc.repo.Entities.GetAllZeroImportRef(ctx, gid)
	if err != nil {
		recordServiceSpanError(span, err)
		return 0, err
	}
	span.SetAttributes(attribute.Int("entities.zero_import_ref.count", len(ids)))

	_, patchSpan := entityServiceTracer().Start(ctx, "service.EntityService.EnsureImportRef.patch",
		trace.WithAttributes(attribute.Int("entities.count", len(ids))))
	defer patchSpan.End()

	finished := 0
	for _, entityID := range ids {
		ref := uuid.New().String()[0:8]
		err = svc.repo.Entities.Patch(ctx, gid, entityID, repo.EntityPatch{ImportRef: &ref})
		if err != nil {
			recordServiceSpanError(patchSpan, err)
			recordServiceSpanError(span, err)
			return 0, err
		}

		finished++
	}

	patchSpan.SetAttributes(attribute.Int("entities.patched.count", finished))
	span.SetAttributes(attribute.Int("entities.patched.count", finished))
	return finished, nil
}

func serializeLocation[T ~[]string](location T) string {
	return strings.Join(location, "/")
}

// CsvImport imports entities from a CSV file using the standard defined format.
//
// CsvImport applies the following rules/operations
//
//  1. If the entity does not exist, it is created.
//  2. If the entity has an ImportRef and it exists it is skipped
//  3. Locations and Tags are created if they do not exist.
func (svc *EntityService) CsvImport(ctx context.Context, gid uuid.UUID, data io.Reader) (int, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.CsvImport",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	_, readSpan := entityServiceTracer().Start(ctx, "service.EntityService.CsvImport.readCsv")
	sheet := reporting.IOSheet{}

	err := sheet.Read(data)
	if err != nil {
		recordServiceSpanError(readSpan, err)
		readSpan.End()
		recordServiceSpanError(span, err)
		return 0, err
	}
	readSpan.SetAttributes(attribute.Int("rows.count", len(sheet.Rows)))
	readSpan.End()
	span.SetAttributes(attribute.Int("rows.count", len(sheet.Rows)))

	// ========================================
	// Tags

	var tagMap map[string]uuid.UUID
	{
		tagsCtx, tagsSpan := entityServiceTracer().Start(ctx, "service.EntityService.CsvImport.loadTags")
		tags, err := svc.repo.Tags.GetAll(tagsCtx, gid)
		if err != nil {
			recordServiceSpanError(tagsSpan, err)
			tagsSpan.End()
			recordServiceSpanError(span, err)
			return 0, err
		}

		tagMap = lo.SliceToMap(tags, func(tag repo.TagSummary) (string, uuid.UUID) {
			return tag.Name, tag.ID
		})
		tagsSpan.SetAttributes(attribute.Int("tags.count", len(tags)))
		tagsSpan.End()
	}

	// ========================================
	// Locations

	locationMap := make(map[string]uuid.UUID)
	{
		locsCtx, locsSpan := entityServiceTracer().Start(ctx, "service.EntityService.CsvImport.loadLocations")
		locations, err := svc.repo.Entities.Tree(locsCtx, gid, repo.TreeQuery{WithItems: false})
		if err != nil {
			recordServiceSpanError(locsSpan, err)
			locsSpan.End()
			recordServiceSpanError(span, err)
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
		locsSpan.SetAttributes(
			attribute.Int("locations.tree.count", len(locations)),
			attribute.Int("locations.flat.count", len(locationMap)),
		)
		locsSpan.End()
	}

	// ========================================
	// Import entities

	// Asset ID Pre-Check
	highestAID := repo.AssetID(-1)
	if svc.autoIncrementAssetID {
		highestAID, err = svc.repo.Entities.GetHighestAssetID(ctx, gid)
		if err != nil {
			recordServiceSpanError(span, err)
			return 0, err
		}
	}

	importCtx, importSpan := entityServiceTracer().Start(ctx, "service.EntityService.CsvImport.importRows",
		trace.WithAttributes(attribute.Int("rows.count", len(sheet.Rows))))
	defer importSpan.End()

	finished := 0

	for i := range sheet.Rows {
		row := sheet.Rows[i]

		rowCtx, rowSpan := entityServiceTracer().Start(importCtx, "service.EntityService.CsvImport.row",
			trace.WithAttributes(
				attribute.Int("row.index", i),
				attribute.String("row.name", row.Name),
				attribute.String("row.import_ref", row.ImportRef),
				attribute.Int("row.tags.count", len(row.TagStr)),
				attribute.Int("row.location.depth", len(row.Location)),
			))
		ctx := rowCtx

		createRequired := true

		if row.ImportRef != "" {
			exists, err := svc.repo.Entities.CheckRef(ctx, gid, row.ImportRef)
			if err != nil {
				wrapped := fmt.Errorf("error checking for existing entity with ref %q: %w", row.ImportRef, err)
				recordServiceSpanError(rowSpan, wrapped)
				rowSpan.End()
				recordServiceSpanError(importSpan, wrapped)
				recordServiceSpanError(span, wrapped)
				return 0, wrapped
			}

			if exists {
				createRequired = false
			}
		}
		rowSpan.SetAttributes(attribute.Bool("row.create_required", createRequired))

		// ========================================
		// Pre-Create tags as necessary
		tagIds := make([]uuid.UUID, len(row.TagStr))

		if len(row.TagStr) > 0 {
			tagsCtx, tagsSpan := entityServiceTracer().Start(rowCtx, "service.EntityService.CsvImport.row.tags",
				trace.WithAttributes(attribute.Int("tags.count", len(row.TagStr))))
			tagsCreated := 0
			for j := range row.TagStr {
				tag := row.TagStr[j]

				id, ok := tagMap[tag]
				if !ok {
					newTag, err := svc.repo.Tags.Create(tagsCtx, gid, repo.TagCreate{Name: tag})
					if err != nil {
						recordServiceSpanError(tagsSpan, err)
						tagsSpan.End()
						recordServiceSpanError(rowSpan, err)
						rowSpan.End()
						recordServiceSpanError(importSpan, err)
						recordServiceSpanError(span, err)
						return 0, err
					}
					id = newTag.ID
					tagsCreated++
				}

				tagIds[j] = id
				tagMap[tag] = id
			}
			tagsSpan.SetAttributes(attribute.Int("tags.created.count", tagsCreated))
			tagsSpan.End()
		}

		// ========================================
		// Pre-Create Locations as necessary
		path := serializeLocation(row.Location)

		locationID, ok := locationMap[path]
		if !ok {
			locsCtx, locsSpan := entityServiceTracer().Start(rowCtx, "service.EntityService.CsvImport.row.locations",
				trace.WithAttributes(attribute.Int("location.depth", len(row.Location))))
			locsCreated := 0
			paths := []string{}
			for i, pathElement := range row.Location {
				paths = append(paths, pathElement)
				path := serializeLocation(paths)

				locationID, ok = locationMap[path]
				if !ok {
					parentID := uuid.Nil

					if i > 0 {
						parentPath := serializeLocation(row.Location[:i])
						parentID = locationMap[parentPath]
					}

					newLocation, err := svc.repo.Entities.CreateContainer(locsCtx, gid, repo.EntityCreate{
						ParentID: parentID,
						Name:     pathElement,
					})
					if err != nil {
						recordServiceSpanError(locsSpan, err)
						locsSpan.End()
						recordServiceSpanError(rowSpan, err)
						rowSpan.End()
						recordServiceSpanError(importSpan, err)
						recordServiceSpanError(span, err)
						return 0, err
					}
					locationID = newLocation.ID
					locsCreated++
				}

				locationMap[path] = locationID
			}
			locsSpan.SetAttributes(attribute.Int("locations.created.count", locsCreated))
			locsSpan.End()

			locationID, ok = locationMap[path]
			if !ok {
				err := errors.New("failed to create location")
				recordServiceSpanError(rowSpan, err)
				rowSpan.End()
				recordServiceSpanError(importSpan, err)
				recordServiceSpanError(span, err)
				return 0, err
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
		// Create Entity
		var entity repo.EntityOut
		switch {
		case createRequired:
			newEntity := repo.EntityCreate{
				ImportRef:   row.ImportRef,
				Name:        row.Name,
				Description: row.Description,
				AssetID:     effAID,
				ParentID:    locationID,
				TagIDs:      tagIds,
			}

			entity, err = svc.repo.Entities.Create(ctx, gid, newEntity)
			if err != nil {
				recordServiceSpanError(rowSpan, err)
				rowSpan.End()
				recordServiceSpanError(importSpan, err)
				recordServiceSpanError(span, err)
				return 0, err
			}
		default:
			entity, err = svc.repo.Entities.GetByRef(ctx, gid, row.ImportRef)
			if err != nil {
				recordServiceSpanError(rowSpan, err)
				rowSpan.End()
				recordServiceSpanError(importSpan, err)
				recordServiceSpanError(span, err)
				return 0, err
			}
		}

		if entity.ID == uuid.Nil {
			wrapped := fmt.Errorf("entity ID is nil for entity with import ref %q", row.ImportRef)
			recordServiceSpanError(rowSpan, wrapped)
			rowSpan.End()
			recordServiceSpanError(importSpan, wrapped)
			recordServiceSpanError(span, wrapped)
			return 0, wrapped
		}
		rowSpan.SetAttributes(attribute.String("entity.id", entity.ID.String()))

		fields := lo.Map(row.Fields, func(f reporting.ExportItemFields, _ int) repo.EntityFieldData {
			return repo.EntityFieldData{
				Name:      f.Name,
				Type:      "text",
				TextValue: f.Value,
			}
		})

		updateEntity := repo.EntityUpdate{
			ID:       entity.ID,
			TagIDs:   tagIds,
			ParentID: locationID,

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

		_, err = svc.repo.Entities.UpdateByGroup(ctx, gid, updateEntity)
		if err != nil {
			recordServiceSpanError(rowSpan, err)
			rowSpan.End()
			recordServiceSpanError(importSpan, err)
			recordServiceSpanError(span, err)
			return 0, err
		}

		finished++
		rowSpan.End()
	}

	importSpan.SetAttributes(attribute.Int("rows.imported.count", finished))
	span.SetAttributes(attribute.Int("rows.imported.count", finished))
	return finished, nil
}

func (svc *EntityService) ExportCSV(ctx context.Context, gid uuid.UUID, hbURL string) ([][]string, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.ExportCSV",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	loadCtx, loadSpan := entityServiceTracer().Start(ctx, "service.EntityService.ExportCSV.load")
	items, err := svc.repo.Entities.GetAll(loadCtx, gid)
	if err != nil {
		recordServiceSpanError(loadSpan, err)
		loadSpan.End()
		recordServiceSpanError(span, err)
		return nil, err
	}
	loadSpan.SetAttributes(attribute.Int("entities.count", len(items)))
	loadSpan.End()
	span.SetAttributes(attribute.Int("entities.count", len(items)))

	readCtx, readSpan := entityServiceTracer().Start(ctx, "service.EntityService.ExportCSV.readItems")
	sheet := reporting.IOSheet{}
	err = sheet.ReadItems(readCtx, items, gid, svc.repo, hbURL)
	if err != nil {
		recordServiceSpanError(readSpan, err)
		readSpan.End()
		recordServiceSpanError(span, err)
		return nil, err
	}
	readSpan.End()

	_, csvSpan := entityServiceTracer().Start(ctx, "service.EntityService.ExportCSV.encode")
	defer csvSpan.End()
	rows, err := sheet.CSV()
	if err != nil {
		recordServiceSpanError(csvSpan, err)
		recordServiceSpanError(span, err)
		return nil, err
	}
	csvSpan.SetAttributes(attribute.Int("rows.count", len(rows)))
	return rows, nil
}

func (svc *EntityService) ExportBillOfMaterialsCSV(ctx context.Context, gid uuid.UUID) ([]byte, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.ExportBillOfMaterialsCSV",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	loadCtx, loadSpan := entityServiceTracer().Start(ctx, "service.EntityService.ExportBillOfMaterialsCSV.load")
	items, err := svc.repo.Entities.GetAll(loadCtx, gid)
	if err != nil {
		recordServiceSpanError(loadSpan, err)
		loadSpan.End()
		recordServiceSpanError(span, err)
		return nil, err
	}
	loadSpan.SetAttributes(attribute.Int("entities.count", len(items)))
	loadSpan.End()
	span.SetAttributes(attribute.Int("entities.count", len(items)))

	_, encodeSpan := entityServiceTracer().Start(ctx, "service.EntityService.ExportBillOfMaterialsCSV.encode")
	defer encodeSpan.End()
	out, err := reporting.BillOfMaterialsCSV(items)
	if err != nil {
		recordServiceSpanError(encodeSpan, err)
		recordServiceSpanError(span, err)
		return nil, err
	}
	encodeSpan.SetAttributes(attribute.Int("bytes.size", len(out)))
	return out, nil
}
