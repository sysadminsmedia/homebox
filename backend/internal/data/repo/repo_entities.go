package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entityfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type EntitiesRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	EntityFieldQuery struct {
		Name  string
		Value string
	}

	EntityQuery struct {
		Page             int
		PageSize         int
		Search           string       `json:"search"`
		AssetID          AssetID      `json:"assetId"`
		LocationIDs      []uuid.UUID  `json:"locationIds"`
		TagIDs           []uuid.UUID  `json:"labelIds"`
		NegateTags       bool         `json:"negateTags"`
		OnlyWithoutPhoto bool         `json:"onlyWithoutPhoto"`
		OnlyWithPhoto    bool         `json:"onlyWithPhoto"`
		ParentItemIDs    []uuid.UUID  `json:"parentIds"`
		SortBy           string       `json:"sortBy"`
		IncludeArchived  bool         `json:"includeArchived"`
		Fields           []FieldQuery `json:"fields"`
		OrderBy          string       `json:"orderBy"`
	}

	EntityDuplicateOptions struct {
		CopyMaintenance  bool   `json:"copyMaintenance"`
		CopyAttachments  bool   `json:"copyAttachments"`
		CopyCustomFields bool   `json:"copyCustomFields"`
		CopyPrefix       string `json:"copyPrefix"`
	}

	EntityField struct {
		ID           uuid.UUID `json:"id,omitempty"`
		Type         string    `json:"type"`
		Name         string    `json:"name"`
		TextValue    string    `json:"textValue"`
		NumberValue  int       `json:"numberValue"`
		BooleanValue bool      `json:"booleanValue"`
		// TimeValue    time.Time `json:"timeValue,omitempty"`
	}

	EntityCreate struct {
		ImportRef   string    `json:"-"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Quantity    int       `json:"quantity"`
		Description string    `json:"description" validate:"max=1000"`
		AssetID     AssetID   `json:"-"`
		EntityType  uuid.UUID `json:"entityType"  validate:"required,uuid"`

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		TagIDs     []uuid.UUID `json:"labelIds"`
	}

	EntityUpdate struct {
		ParentID                uuid.UUID `json:"parentId"                extensions:"x-nullable,x-omitempty"`
		ID                      uuid.UUID `json:"id"`
		AssetID                 AssetID   `json:"assetId"                 swaggertype:"string"`
		Name                    string    `json:"name"                    validate:"required,min=1,max=255"`
		Description             string    `json:"description"             validate:"max=1000"`
		Quantity                int       `json:"quantity"`
		Insured                 bool      `json:"insured"`
		Archived                bool      `json:"archived"`
		SyncChildItemsLocations bool      `json:"syncChildItemsLocations"`
		EntityType              uuid.UUID `json:"entityType"              validate:"required,uuid"`

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		TagIDs     []uuid.UUID `json:"labelIds"`

		// Identifications
		SerialNumber string `json:"serialNumber"`
		ModelNumber  string `json:"modelNumber"`
		Manufacturer string `json:"manufacturer"`

		// Warranty
		LifetimeWarranty bool       `json:"lifetimeWarranty"`
		WarrantyExpires  types.Date `json:"warrantyExpires"`
		WarrantyDetails  string     `json:"warrantyDetails"`

		// Purchase
		PurchaseTime  types.Date `json:"purchaseTime"`
		PurchaseFrom  string     `json:"purchaseFrom"  validate:"max=255"`
		PurchasePrice float64    `json:"purchasePrice" extensions:"x-nullable,x-omitempty"`

		// Sold
		SoldTime  types.Date `json:"soldTime"`
		SoldTo    string     `json:"soldTo"    validate:"max=255"`
		SoldPrice float64    `json:"soldPrice" extensions:"x-nullable,x-omitempty"`
		SoldNotes string     `json:"soldNotes"`

		// Extras
		Notes  string        `json:"notes"`
		Fields []EntityField `json:"fields"`
	}

	EntityPatch struct {
		ID        uuid.UUID `json:"id"`
		Quantity  *int      `json:"quantity,omitempty" extensions:"x-nullable,x-omitempty"`
		ImportRef *string   `json:"-,omitempty"        extensions:"x-nullable,x-omitempty"`
	}

	EntitySummary struct {
		ImportRef   string    `json:"-"`
		ID          uuid.UUID `json:"id"`
		AssetID     AssetID   `json:"assetId,string"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Quantity    int       `json:"quantity"`
		Insured     bool      `json:"insured"`
		Archived    bool      `json:"archived"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		EntityType  uuid.UUID `json:"entityType"`

		PurchasePrice float64 `json:"purchasePrice"`

		// Edges
		Location *LocationSummary `json:"location,omitempty" extensions:"x-nullable,x-omitempty"`
		Tags     []TagSummary     `json:"labels"`

		ImageID     *uuid.UUID `json:"imageId,omitempty"     extensions:"x-nullable,x-omitempty"`
		ThumbnailId *uuid.UUID `json:"thumbnailId,omitempty" extensions:"x-nullable,x-omitempty"`

		// Sale details
		SoldTime time.Time `json:"soldTime"`
	}

	EntityOut struct {
		Parent *EntitySummary `json:"parent,omitempty" extensions:"x-nullable,x-omitempty"`
		EntitySummary
		AssetID AssetID `json:"assetId,string"`

		SyncChildItemsLocations bool `json:"syncChildItemsLocations"`

		SerialNumber string `json:"serialNumber"`
		ModelNumber  string `json:"modelNumber"`
		Manufacturer string `json:"manufacturer"`

		// Warranty
		LifetimeWarranty bool       `json:"lifetimeWarranty"`
		WarrantyExpires  types.Date `json:"warrantyExpires"`
		WarrantyDetails  string     `json:"warrantyDetails"`

		// Purchase
		PurchaseTime types.Date `json:"purchaseTime"`
		PurchaseFrom string     `json:"purchaseFrom"`

		// Sold
		SoldTime  types.Date `json:"soldTime"`
		SoldTo    string     `json:"soldTo"`
		SoldPrice float64    `json:"soldPrice"`
		SoldNotes string     `json:"soldNotes"`

		// Extras
		Notes string `json:"notes"`

		Attachments []EntityAttachment `json:"attachments"`
		Fields      []EntityField      `json:"fields"`
	}
	TreeEntity struct {
		ID       uuid.UUID     `json:"id"`
		Name     string        `json:"name"`
		Type     string        `json:"type"`
		Children []*TreeEntity `json:"children"`
	}
	FlatTreeEntity struct {
		ID       uuid.UUID
		Name     string
		Type     string
		ParentID uuid.UUID
		Level    int
	}
	EntityTreeQuery struct {
		WithItems bool `json:"withItems" schema:"withItems"`
	}
	EntityEntityType string
	EntityPath       struct {
		Type ItemType  `json:"type"`
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
)

const (
	EntityTypeLocation ItemType = "location"
	EntityTypeItem     ItemType = "item"
)

var mapEntitiesSummaryErr = mapTEachErrFunc(mapEntitySummary)

func mapEntitySummary(item *ent.Entity) EntitySummary {
	var location *LocationSummary
	if item.Edges.Location != nil {
		loc := mapLocationSummary(item.QueryLocation().Where(entity.HasTypeWith(entitytype.IsLocationEQ(true))).FirstX(context.Background()))
		location = &loc
	}

	labels := make([]TagSummary, len(item.Edges.Tag))
	if item.Edges.Tag != nil {
		labels = mapEach(item.Edges.Tag, mapTagSummary)
	}

	var imageID *uuid.UUID
	var thumbnailID *uuid.UUID
	if item.Edges.Attachments != nil {
		for _, a := range item.Edges.Attachments {
			if a.Primary && a.Type == attachment.TypePhoto {
				imageID = &a.ID
				if a.Edges.Thumbnail != nil {
					if a.Edges.Thumbnail.ID != uuid.Nil {
						thumbnailID = &a.Edges.Thumbnail.ID
					} else {
						thumbnailID = nil
					}
				} else {
					thumbnailID = nil
				}
				break
			}
		}
	}

	var typeID uuid.UUID
	if item.Edges.Type != nil {
		typeID = item.Edges.Type.ID
	}

	return EntitySummary{
		ID:            item.ID,
		AssetID:       AssetID(item.AssetID),
		Name:          item.Name,
		Description:   item.Description,
		ImportRef:     item.ImportRef,
		Quantity:      item.Quantity,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		Archived:      item.Archived,
		PurchasePrice: item.PurchasePrice,
		EntityType:    typeID,

		// Edges
		Location: location,
		Tags:     labels,

		// Warranty
		Insured:     item.Insured,
		ImageID:     imageID,
		ThumbnailId: thumbnailID,
	}
}

var (
	mapEntityOutErr   = mapTErrFunc(mapEntityOut)
	mapEntitiesOutErr = mapTEachErrFunc(mapEntityOut)
)

func mapEntityFields(fields []*ent.EntityField) []EntityField {
	result := make([]EntityField, len(fields))
	for i, f := range fields {
		result[i] = EntityField{
			ID:           f.ID,
			Type:         f.Type.String(),
			Name:         f.Name,
			TextValue:    f.TextValue,
			NumberValue:  f.NumberValue,
			BooleanValue: f.BooleanValue,
			// TimeValue:    f.TimeValue,
		}
	}
	return result
}

func mapEntityOut(item *ent.Entity) EntityOut {
	var attachments []EntityAttachment
	if item.Edges.Attachments != nil {
		attachments = mapEach(item.Edges.Attachments, ToItemAttachment)
	}

	var fields []EntityField
	if item.Edges.Fields != nil {
		fields = mapEntityFields(item.Edges.Fields)
	}

	var parent *EntitySummary
	if item.Edges.Parent != nil {
		v := mapEntitySummary(item.Edges.Parent)
		parent = &v
	}

	return EntityOut{
		Parent:                  parent,
		AssetID:                 AssetID(item.AssetID),
		EntitySummary:           mapEntitySummary(item),
		LifetimeWarranty:        item.LifetimeWarranty,
		WarrantyExpires:         types.DateFromTime(item.WarrantyExpires),
		WarrantyDetails:         item.WarrantyDetails,
		SyncChildItemsLocations: item.SyncChildEntitiesLocations,

		// Identification
		SerialNumber: item.SerialNumber,
		ModelNumber:  item.ModelNumber,
		Manufacturer: item.Manufacturer,

		// Purchase
		PurchaseTime: types.DateFromTime(item.PurchaseTime),
		PurchaseFrom: item.PurchaseFrom,

		// Sold
		SoldTime:  types.DateFromTime(item.SoldTime),
		SoldTo:    item.SoldTo,
		SoldPrice: item.SoldPrice,
		SoldNotes: item.SoldNotes,

		// Extras
		Notes:       item.Notes,
		Attachments: attachments,
		Fields:      fields,
	}
}

func (e *EntitiesRepository) publishMutationEvent(gid uuid.UUID) {
	if e.bus != nil {
		e.bus.Publish(eventbus.EventItemMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (e *EntitiesRepository) getOne(ctx context.Context, where ...predicate.Entity) (EntityOut, error) {
	q := e.db.Entity.Query().Where(where...).Where(entity.HasTypeWith(entitytype.IsLocationEQ(false)))

	return mapEntityOutErr(q.
		WithFields().
		WithTag().
		WithLocation().
		WithGroup().
		WithParent().
		WithAttachments().
		WithType().
		Only(ctx),
	)
}

// GetOne returns a single item by ID. If the item does not exist, an error is returned.
// See also: GetOneByGroup to ensure that the item belongs to a specific group.
func (e *EntitiesRepository) GetOne(ctx context.Context, id uuid.UUID) (EntityOut, error) {
	return e.getOne(ctx, entity.ID(id))
}

func (e *EntitiesRepository) CheckRef(ctx context.Context, gid uuid.UUID, ref string) (bool, error) {
	q := e.db.Entity.Query().Where(entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
	return q.Where(entity.ImportRef(ref)).Exist(ctx)
}

func (e *EntitiesRepository) GetByRef(ctx context.Context, gid uuid.UUID, ref string) (EntityOut, error) {
	return e.getOne(ctx, entity.ImportRef(ref), entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
}

// GetOneByGroup returns a single item by ID. If the item does not exist, an error is returned.
// GetOneByGroup ensures that the item belongs to a specific group.
func (e *EntitiesRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (EntityOut, error) {
	return e.getOne(ctx, entity.ID(id), entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
}

// QueryByGroup returns a list of items that belong to a specific group based on the provided query.
func (e *EntitiesRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q ItemQuery) (PaginationResult[EntitySummary], error) {
	qb := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.HasTypeWith(entitytype.IsLocationEQ(false)),
	)

	if q.IncludeArchived {
		qb = qb.Where(
			entity.Or(
				entity.Archived(true),
				entity.Archived(false),
			),
		)
	} else {
		qb = qb.Where(entity.Archived(false))
	}

	if q.Search != "" {
		// Use accent-insensitive search predicates that normalize both
		// the search query and database field values during comparison.
		// For queries without accents, the traditional search is more efficient.
		qb.Where(
			entity.Or(
				// Regular case-insensitive search (fastest)
				entity.NameContainsFold(q.Search),
				entity.DescriptionContainsFold(q.Search),
				entity.SerialNumberContainsFold(q.Search),
				entity.ModelNumberContainsFold(q.Search),
				entity.ManufacturerContainsFold(q.Search),
				entity.NotesContainsFold(q.Search),
				// Accent-insensitive search using custom predicates
				ent.ItemNameAccentInsensitiveContains(q.Search),
				ent.ItemDescriptionAccentInsensitiveContains(q.Search),
				ent.ItemSerialNumberAccentInsensitiveContains(q.Search),
				ent.ItemModelNumberAccentInsensitiveContains(q.Search),
				ent.ItemManufacturerAccentInsensitiveContains(q.Search),
				ent.ItemNotesAccentInsensitiveContains(q.Search),
			),
		)
	}

	if !q.AssetID.Nil() {
		qb = qb.Where(entity.AssetID(q.AssetID.Int()))
	}

	// Filters within this block define a AND relationship where each subset
	// of filters is OR'd together.
	//
	// The goal is to allow matches like where the item has
	//  - one of the selected labels AND
	//  - one of the selected locations AND
	//  - one of the selected fields key/value matches
	var andPredicates []predicate.Entity
	{
		if len(q.TagIDs) > 0 {
			labelPredicates := make([]predicate.Entity, 0, len(q.TagIDs))
			for _, l := range q.TagIDs {
				if !q.NegateTags {
					labelPredicates = append(labelPredicates, entity.HasTagWith(tag.ID(l)))
				} else {
					labelPredicates = append(labelPredicates, entity.Not(entity.HasTagWith(tag.ID(l))))
				}
			}
			if !q.NegateTags {
				andPredicates = append(andPredicates, entity.Or(labelPredicates...))
			} else {
				andPredicates = append(andPredicates, entity.And(labelPredicates...))
			}
		}

		if q.OnlyWithoutPhoto {
			andPredicates = append(andPredicates, entity.Not(
				entity.HasAttachmentsWith(
					attachment.And(
						attachment.Primary(true),
						attachment.TypeEQ(attachment.TypePhoto),
					),
				)),
			)
		}

		if q.OnlyWithPhoto {
			andPredicates = append(andPredicates, entity.HasAttachmentsWith(
				attachment.And(
					attachment.Primary(true),
					attachment.TypeEQ(attachment.TypePhoto),
				),
			),
			)
		}

		if len(q.LocationIDs) > 0 {
			locationPredicates := make([]predicate.Entity, 0, len(q.LocationIDs))
			for _, l := range q.LocationIDs {
				locationPredicates = append(locationPredicates, entity.HasLocationWith(entity.ID(l), entity.HasTypeWith(entitytype.IsLocationEQ(false))))
			}

			andPredicates = append(andPredicates, entity.Or(locationPredicates...))
		}

		if len(q.Fields) > 0 {
			fieldPredicates := make([]predicate.Entity, 0, len(q.Fields))
			for _, f := range q.Fields {
				fieldPredicates = append(fieldPredicates, entity.HasFieldsWith(
					entityfield.And(
						entityfield.Name(f.Name),
						entityfield.TextValue(f.Value),
					),
				))
			}

			andPredicates = append(andPredicates, entity.Or(fieldPredicates...))
		}

		if len(q.ParentItemIDs) > 0 {
			andPredicates = append(andPredicates, entity.HasParentWith(entity.IDIn(q.ParentItemIDs...)))
		}
	}

	if len(andPredicates) > 0 {
		qb = qb.Where(entity.And(andPredicates...))
	}

	count, err := qb.Count(ctx)
	if err != nil {
		return PaginationResult[EntitySummary]{}, err
	}

	// Order
	switch q.OrderBy {
	case "createdAt":
		qb = qb.Order(ent.Desc(entity.FieldCreatedAt))
	case "updatedAt":
		qb = qb.Order(ent.Desc(entity.FieldUpdatedAt))
	case "assetId":
		qb = qb.Order(ent.Asc(entity.FieldAssetID))
	default: // "name"
		qb = qb.Order(ent.Asc(entity.FieldName))
	}

	qb = qb.
		WithTag().
		WithLocation().
		WithType().
		WithAttachments(func(aq *ent.AttachmentQuery) {
			aq.Where(
				attachment.Primary(true),
			)
			aq.WithThumbnail()
		})

	if q.Page != -1 || q.PageSize != -1 {
		qb = qb.
			Offset(calculateOffset(q.Page, q.PageSize)).
			Limit(q.PageSize)
	}

	items, err := mapEntitiesSummaryErr(qb.All(ctx))
	if err != nil {
		return PaginationResult[EntitySummary]{}, err
	}

	return PaginationResult[EntitySummary]{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    count,
		Items:    items,
	}, nil
}

// QueryByAssetID returns items by asset ID. If the item does not exist, an error is returned.
func (e *EntitiesRepository) QueryByAssetID(ctx context.Context, gid uuid.UUID, assetID AssetID, page int, pageSize int) (PaginationResult[EntitySummary], error) {
	qb := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(int(assetID)),
		entity.HasTypeWith(entitytype.IsLocationEQ(false)),
	)

	if page != -1 || pageSize != -1 {
		qb.Offset(calculateOffset(page, pageSize)).
			Limit(pageSize)
	} else {
		page = -1
		pageSize = -1
	}

	items, err := mapEntitiesSummaryErr(
		qb.Order(ent.Asc(entity.FieldName)).
			WithTag().
			WithLocation().
			WithType().
			All(ctx),
	)
	if err != nil {
		return PaginationResult[EntitySummary]{}, err
	}

	return PaginationResult[EntitySummary]{
		Page:     page,
		PageSize: pageSize,
		Total:    len(items),
		Items:    items,
	}, nil
}

// GetAll returns all the items in the database with the Tags and Locations eager loaded.
func (e *EntitiesRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]EntityOut, error) {
	return mapEntitiesOutErr(e.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false))).
		WithTag().
		WithLocation().
		WithFields().
		WithType().
		All(ctx))
}

func (e *EntitiesRepository) GetAllZeroAssetID(ctx context.Context, gid uuid.UUID) ([]EntitySummary, error) {
	q := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(0),
	).Order(
		ent.Asc(entity.FieldCreatedAt),
	)

	return mapEntitiesSummaryErr(q.All(ctx))
}

func (e *EntitiesRepository) GetHighestAssetID(ctx context.Context, gid uuid.UUID) (AssetID, error) {
	q := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
	).Order(
		ent.Desc(entity.FieldAssetID),
	).Limit(1)

	result, err := q.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}

	return AssetID(result.AssetID), nil
}

func (e *EntitiesRepository) SetAssetID(ctx context.Context, gid uuid.UUID, id uuid.UUID, assetID AssetID) error {
	q := e.db.Entity.Update().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.ID(id),
	)

	_, err := q.SetAssetID(int(assetID)).Save(ctx)
	return err
}

func (e *EntitiesRepository) Create(ctx context.Context, gid uuid.UUID, data EntityCreate) (EntityOut, error) {
	q := e.db.Entity.Create().
		SetImportRef(data.ImportRef).
		SetName(data.Name).
		SetQuantity(data.Quantity).
		SetDescription(data.Description).
		SetGroupID(gid).
		SetAssetID(int(data.AssetID)).
		SetTypeID(data.EntityType).
		SetLocationID(data.LocationID)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if len(data.TagIDs) > 0 {
		q.AddTagIDs(data.TagIDs...)
	}

	result, err := q.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	e.publishMutationEvent(gid)
	return e.GetOne(ctx, result.ID)
}

func (e *EntitiesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := e.db.Entity.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	e.publishMutationEvent(id)
	return nil
}

func (e *EntitiesRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := e.db.Entity.
		Delete().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	e.publishMutationEvent(gid)
	return err
}

func (e *EntitiesRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data EntityUpdate) (EntityOut, error) {
	q := e.db.Entity.Update().Where(entity.ID(data.ID), entity.HasGroupWith(group.ID(gid))).
		SetName(data.Name).
		SetDescription(data.Description).
		SetLocationID(data.LocationID).
		SetSerialNumber(data.SerialNumber).
		SetModelNumber(data.ModelNumber).
		SetManufacturer(data.Manufacturer).
		SetArchived(data.Archived).
		SetPurchaseTime(data.PurchaseTime.Time()).
		SetPurchaseFrom(data.PurchaseFrom).
		SetPurchasePrice(data.PurchasePrice).
		SetSoldTime(data.SoldTime.Time()).
		SetSoldTo(data.SoldTo).
		SetSoldPrice(data.SoldPrice).
		SetSoldNotes(data.SoldNotes).
		SetNotes(data.Notes).
		SetLifetimeWarranty(data.LifetimeWarranty).
		SetInsured(data.Insured).
		SetWarrantyExpires(data.WarrantyExpires.Time()).
		SetWarrantyDetails(data.WarrantyDetails).
		SetQuantity(data.Quantity).
		SetAssetID(int(data.AssetID)).
		SetTypeID(data.EntityType).
		SetSyncChildEntitiesLocations(data.SyncChildItemsLocations)

	currentTags, err := e.db.Entity.Query().Where(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))).QueryTag().All(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	set := newIDSet(currentTags)

	for _, l := range data.TagIDs {
		if set.Contains(l) {
			set.Remove(l)
			continue
		}
		q.AddTagIDs(l)
	}

	if set.Len() > 0 {
		q.RemoveTagIDs(set.Slice()...)
	}

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	if data.SyncChildItemsLocations {
		children, err := e.db.Entity.Query().Where(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))).QueryChildren().All(ctx)
		if err != nil {
			return EntityOut{}, err
		}
		location := data.LocationID

		for _, child := range children {
			childLocation, err := child.QueryLocation().First(ctx)
			if err != nil {
				return EntityOut{}, err
			}

			if location != childLocation.ID {
				err = child.Update().SetLocationID(location).Exec(ctx)
				if err != nil {
					return EntityOut{}, err
				}
			}
		}
	}

	err = q.Exec(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	fields, err := e.db.EntityField.Query().Where(entityfield.HasEntityWith(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false)))).All(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	fieldIds := newIDSet(fields)

	// Update Existing Fields
	for _, f := range data.Fields {
		if f.ID == uuid.Nil {
			// Create New Field
			_, err = e.db.EntityField.Create().
				SetEntityID(data.ID).
				SetType(entityfield.Type(f.Type)).
				SetName(f.Name).
				SetTextValue(f.TextValue).
				SetNumberValue(f.NumberValue).
				SetBooleanValue(f.BooleanValue).
				// SetTimeValue(f.TimeValue).
				Save(ctx)
			if err != nil {
				return EntityOut{}, err
			}
		}

		opt := e.db.EntityField.Update().
			Where(
				entityfield.ID(f.ID),
				entityfield.HasEntityWith(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))),
			).
			SetType(entityfield.Type(f.Type)).
			SetName(f.Name).
			SetTextValue(f.TextValue).
			SetNumberValue(f.NumberValue).
			SetBooleanValue(f.BooleanValue)
		// SetTimeValue(f.TimeValue)

		_, err = opt.Save(ctx)
		if err != nil {
			return EntityOut{}, err
		}

		fieldIds.Remove(f.ID)
		continue
	}

	// Delete Fields that are no longer present
	if fieldIds.Len() > 0 {
		_, err = e.db.EntityField.Delete().
			Where(
				entityfield.IDIn(fieldIds.Slice()...),
				entityfield.HasEntityWith(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))),
			).Exec(ctx)
		if err != nil {
			return EntityOut{}, err
		}
	}

	e.publishMutationEvent(gid)
	return e.GetOne(ctx, data.ID)
}

func (e *EntitiesRepository) GetAllZeroImportRef(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID

	err := e.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
			entity.HasTypeWith(entitytype.IsLocationEQ(false)),
			entity.Or(
				entity.ImportRefEQ(""),
				entity.ImportRefIsNil(),
			),
		).
		Select(entity.FieldID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (e *EntitiesRepository) Patch(ctx context.Context, gid, id uuid.UUID, data EntityPatch) error {
	q := e.db.Entity.Update().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
			entity.HasTypeWith(entitytype.IsLocationEQ(false)),
		)

	if data.ImportRef != nil {
		q.SetImportRef(*data.ImportRef)
	}

	if data.Quantity != nil {
		q.SetQuantity(*data.Quantity)
	}

	e.publishMutationEvent(gid)
	return q.Exec(ctx)
}

func (e *EntitiesRepository) GetAllCustomFieldValues(ctx context.Context, gid uuid.UUID, name string) ([]string, error) {
	type st struct {
		Value string `json:"text_value"`
	}

	var values []st

	err := e.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
		).
		QueryFields().
		Where(
			entityfield.Name(name),
		).
		Unique(true).
		Select(entityfield.FieldTextValue).
		Scan(ctx, &values)
	if err != nil {
		return nil, fmt.Errorf("failed to get field values: %w", err)
	}

	valueStrings := make([]string, len(values))
	for i, f := range values {
		valueStrings[i] = f.Value
	}

	return valueStrings, nil
}

func (e *EntitiesRepository) GetAllCustomFieldNames(ctx context.Context, gid uuid.UUID) ([]string, error) {
	type st struct {
		Name string `json:"name"`
	}

	var fields []st

	err := e.db.Entity.Query().
		Where(
			entity.HasTypeWith(entitytype.IsLocationEQ(false)),
			entity.HasGroupWith(group.ID(gid)),
		).
		QueryFields().
		Unique(true).
		Select(entityfield.FieldName).
		Scan(ctx, &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom fields: %w", err)
	}

	fieldNames := make([]string, len(fields))
	for i, f := range fields {
		fieldNames[i] = f.Name
	}

	return fieldNames, nil
}

// ZeroOutTimeFields is a helper function that can be invoked via the UI by a group member which will
// set all date fields to the beginning of the day.
//
// This is designed to resolve a long-time bug that has since been fixed with the time selector on the
// frontend. This function is intended to be used as a one-time fix for existing databases and may be
// removed in the future.
func (e *EntitiesRepository) ZeroOutTimeFields(ctx context.Context, gid uuid.UUID) (int, error) {
	q := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.HasTypeWith(entitytype.IsLocationEQ(false)),
		entity.Or(
			entity.PurchaseTimeNotNil(),
			entity.PurchaseFromLT("0002-01-01"),
			entity.SoldTimeNotNil(),
			entity.SoldToLT("0002-01-01"),
			entity.WarrantyExpiresNotNil(),
			entity.WarrantyDetailsLT("0002-01-01"),
		),
	)

	items, err := q.All(ctx)
	if err != nil {
		return -1, fmt.Errorf("ZeroOutTimeFields() -> failed to get items: %w", err)
	}

	toDateOnly := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	updated := 0

	for _, i := range items {
		updateQ := e.db.Entity.Update().Where(entity.ID(i.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false)))

		if !i.PurchaseTime.IsZero() {
			switch {
			case i.PurchaseTime.Year() < 100:
				updateQ.ClearPurchaseTime()
			default:
				updateQ.SetPurchaseTime(toDateOnly(i.PurchaseTime))
			}
		} else {
			updateQ.ClearPurchaseTime()
		}

		if !i.SoldTime.IsZero() {
			switch {
			case i.SoldTime.Year() < 100:
				updateQ.ClearSoldTime()
			default:
				updateQ.SetSoldTime(toDateOnly(i.SoldTime))
			}
		} else {
			updateQ.ClearSoldTime()
		}

		if !i.WarrantyExpires.IsZero() {
			switch {
			case i.WarrantyExpires.Year() < 100:
				updateQ.ClearWarrantyExpires()
			default:
				updateQ.SetWarrantyExpires(toDateOnly(i.WarrantyExpires))
			}
		} else {
			updateQ.ClearWarrantyExpires()
		}

		_, err = updateQ.Save(ctx)
		if err != nil {
			return updated, fmt.Errorf("ZeroOutTimeFields() -> failed to update item: %w", err)
		}

		updated++
	}

	return updated, nil
}

func (e *EntitiesRepository) SetPrimaryPhotos(ctx context.Context, gid uuid.UUID) (int, error) {
	// All items where there is no primary photo
	itemIDs, err := e.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
			entity.HasTypeWith(entitytype.IsLocationEQ(false)),
			entity.HasAttachmentsWith(
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Not(
					attachment.And(
						attachment.Primary(true),
						attachment.TypeEQ(attachment.TypePhoto),
					),
				),
			),
		).
		IDs(ctx)
	if err != nil {
		return -1, err
	}

	updated := 0
	for _, id := range itemIDs {
		// Find the first photo attachment
		a, err := e.db.Attachment.Query().
			Where(
				attachment.HasEntityWith(entity.ID(id), entity.HasTypeWith(entitytype.IsLocationEQ(false))),
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Primary(false),
			).
			First(ctx)
		if err != nil {
			return updated, err
		}

		// Set it as primary
		_, err = e.db.Attachment.UpdateOne(a).
			SetPrimary(true).
			Save(ctx)
		if err != nil {
			return updated, err
		}

		updated++
	}

	return updated, nil
}

// Duplicate creates a copy of an item with configurable options for what data to copy.
// The new item will have the next available asset ID and a customizable prefix in the name.
func (e *EntitiesRepository) Duplicate(ctx context.Context, gid, id uuid.UUID, options EntityDuplicateOptions) (EntityOut, error) {
	tx, err := e.db.Tx(ctx)
	if err != nil {
		return EntityOut{}, err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during item duplication")
			}
		}
	}()

	// Get the original item with all its data
	originalItem, err := e.getOne(ctx, entity.ID(id), entity.HasGroupWith(group.ID(gid)))
	if err != nil {
		return EntityOut{}, err
	}

	nextAssetID, err := e.GetHighestAssetID(ctx, gid)
	if err != nil {
		return EntityOut{}, err
	}
	nextAssetID++

	// Set default copy prefix if not provided
	if options.CopyPrefix == "" {
		options.CopyPrefix = "Copy of "
	}

	// Create the new item directly in the transaction
	newItemID := uuid.New()
	itemBuilder := tx.Entity.Create().
		SetID(newItemID).
		SetName(options.CopyPrefix + originalItem.Name).
		SetDescription(originalItem.Description).
		SetQuantity(originalItem.Quantity).
		SetLocationID(originalItem.Location.ID).
		SetGroupID(gid).
		SetAssetID(int(nextAssetID)).
		SetSerialNumber(originalItem.SerialNumber).
		SetModelNumber(originalItem.ModelNumber).
		SetManufacturer(originalItem.Manufacturer).
		SetLifetimeWarranty(originalItem.LifetimeWarranty).
		SetWarrantyExpires(originalItem.WarrantyExpires.Time()).
		SetWarrantyDetails(originalItem.WarrantyDetails).
		SetPurchaseTime(originalItem.PurchaseTime.Time()).
		SetPurchaseFrom(originalItem.PurchaseFrom).
		SetPurchasePrice(originalItem.PurchasePrice).
		SetSoldTime(originalItem.SoldTime.Time()).
		SetSoldTo(originalItem.SoldTo).
		SetSoldPrice(originalItem.SoldPrice).
		SetSoldNotes(originalItem.SoldNotes).
		SetNotes(originalItem.Notes).
		SetInsured(originalItem.Insured).
		SetArchived(originalItem.Archived).
		SetTypeID(originalItem.EntityType).
		SetSyncChildEntitiesLocations(originalItem.SyncChildItemsLocations)

	if originalItem.Parent != nil {
		itemBuilder.SetParentID(originalItem.Parent.ID)
	}

	// Add labels
	if len(originalItem.Tags) > 0 {
		labelIDs := make([]uuid.UUID, len(originalItem.Tags))
		for i, label := range originalItem.Tags {
			labelIDs[i] = label.ID
		}
		itemBuilder.AddTagIDs(labelIDs...)
	}

	_, err = itemBuilder.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	// Copy custom fields if requested
	if options.CopyCustomFields {
		for _, field := range originalItem.Fields {
			_, err = tx.EntityField.Create().
				SetEntityID(newItemID).
				SetType(entityfield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				SetNumberValue(field.NumberValue).
				SetBooleanValue(field.BooleanValue).
				Save(ctx)
			if err != nil {
				log.Warn().Err(err).Str("field_name", field.Name).Msg("failed to copy custom field during duplication")
				continue
			}
		}
	}

	// Copy attachments if requested
	if options.CopyAttachments {
		for _, att := range originalItem.Attachments {
			// Get the original attachment file
			originalAttachment, err := tx.Attachment.Query().
				Where(attachment.ID(att.ID)).
				Only(ctx)
			if err != nil {
				// Log error but continue to copy other attachments
				log.Warn().Err(err).Str("attachment_id", att.ID.String()).Msg("failed to find attachment during duplication")
				continue
			}

			// Create a copy of the attachment with the same file path
			// Since files are stored with hash-based paths, this is safe
			_, err = tx.Attachment.Create().
				SetEntityID(newItemID).
				SetType(originalAttachment.Type).
				SetTitle(originalAttachment.Title).
				SetPath(originalAttachment.Path).
				SetMimeType(originalAttachment.MimeType).
				SetPrimary(originalAttachment.Primary).
				Save(ctx)
			if err != nil {
				log.Warn().Err(err).Str("original_attachment_id", att.ID.String()).Msg("failed to copy attachment during duplication")
				continue
			}
		}
	}

	// Copy maintenance entries if requested
	if options.CopyMaintenance {
		maintenanceEntries, err := tx.MaintenanceEntry.Query().
			Where(maintenanceentry.HasEntityWith(entity.ID(id))).
			All(ctx)
		if err == nil {
			for _, entry := range maintenanceEntries {
				_, err = tx.MaintenanceEntry.Create().
					SetEntityID(newItemID).
					SetDate(entry.Date).
					SetScheduledDate(entry.ScheduledDate).
					SetName(entry.Name).
					SetDescription(entry.Description).
					SetCost(entry.Cost).
					Save(ctx)
				if err != nil {
					log.Warn().Err(err).Str("maintenance_entry_id", entry.ID.String()).Msg("failed to copy maintenance entry during duplication")
					continue
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return EntityOut{}, err
	}
	committed = true

	e.publishMutationEvent(gid)

	// Get the final item with all copied data
	return e.GetOne(ctx, newItemID)
}

func (r *EntitiesRepository) PathForLoc(ctx context.Context, gid, locID uuid.UUID) ([]EntityPath, error) {
	query := `WITH RECURSIVE location_path AS (
		SELECT e.id, e.name, e.entity_parent
		FROM entities e
		JOIN entity_types et ON e.entity_type_entities = et.id
		WHERE e.id = $1
		AND e.group_entities = $2
		AND et.is_location = true

		UNION ALL

		SELECT e.id, e.name, e.entity_parent
		FROM entities e
		JOIN entity_types et ON e.entity_type_entities = et.id
		JOIN location_path lp ON e.id = lp.entity_parent
		WHERE et.is_location = true
	  )

	  SELECT id, name
	  FROM location_path`

	rows, err := r.db.Sql().QueryContext(ctx, query, locID, gid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var locations []EntityPath

	for rows.Next() {
		var location EntityPath
		location.Type = ItemTypeLocation
		if err := rows.Scan(&location.ID, &location.Name); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse the order of the locations so that the root is last
	for i := len(locations)/2 - 1; i >= 0; i-- {
		opp := len(locations) - 1 - i
		locations[i], locations[opp] = locations[opp], locations[i]
	}

	return locations, nil
}

func (r *EntitiesRepository) Tree(ctx context.Context, gid uuid.UUID, tq EntityTreeQuery) ([]TreeEntity, error) {
	query := `
		WITH recursive
    location_tree(id, NAME, parent_id, level, node_type) AS
        (SELECT e.id,
                e.NAME,
                e.entity_parent AS parent_id,
                0               AS level,
                'location'      AS node_type
         FROM entities e
                  JOIN entity_types et ON e.entity_type_entities = et.id
         WHERE e.entity_parent IS NULL
           AND et.is_location = true
           AND e.group_entities = ?
         UNION ALL
         SELECT c.id,
                c.NAME,
                c.entity_parent AS parent_id,
                level + 1,
                'location'      AS node_type
         FROM entities c
                  JOIN entity_types et ON c.entity_type_entities = et.id
                  JOIN location_tree p
                       ON c.entity_parent = p.id
         WHERE et.is_location = true
           AND level < 10){{ WITH_ITEMS }}

		SELECT id,
       NAME,
       level,
       parent_id,
       node_type
FROM (SELECT *
      FROM location_tree

					{{ WITH_ITEMS_FROM }}

				) tree
ORDER BY node_type DESC,
         level,
         lower(NAME)`

	if tq.WithItems {
		itemQuery := `,
    item_tree(id, NAME, parent_id, level, node_type) AS
        (SELECT e.id,
                e.NAME,
                -- 1. Set parent_id to the location's ID
                e.entity_location AS parent_id,
                -- 2. Set level to be the location's level + 1
                lt.level + 1      AS level,
                'item'            AS node_type
         FROM entities e
                  JOIN entity_types et ON e.entity_type_entities = et.id
             -- Join location_tree to get the parent location's level
                  JOIN location_tree lt ON e.entity_location = lt.id
         WHERE et.is_location = false

         UNION ALL

         SELECT c.id,
                c.NAME,
                c.entity_parent AS parent_id,
                level + 1,
                'item'          AS node_type
         FROM entities c
                  JOIN entity_types et ON c.entity_type_entities = et.id
                  JOIN item_tree p
                       ON c.entity_parent = p.id
         WHERE c.entity_parent IS NOT NULL
           AND et.is_location = false
           AND level < 10
        )`

		// Conditional table joined to main query
		itemsFrom := `
		UNION ALL
		SELECT  *
		FROM    item_tree`

		query = strings.ReplaceAll(query, "{{ WITH_ITEMS }}", itemQuery)
		query = strings.ReplaceAll(query, "{{ WITH_ITEMS_FROM }}", itemsFrom)
	} else {
		query = strings.ReplaceAll(query, "{{ WITH_ITEMS }}", "")
		query = strings.ReplaceAll(query, "{{ WITH_ITEMS_FROM }}", "")
	}

	rows, err := r.db.Sql().QueryContext(ctx, query, gid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var locations []FlatTreeEntity
	for rows.Next() {
		var location FlatTreeEntity
		if err := rows.Scan(&location.ID, &location.Name, &location.Level, &location.ParentID, &location.Type); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ConvertEntitiesToTree(locations), nil
}

func ConvertEntitiesToTree(locations []FlatTreeEntity) []TreeEntity {
	locationMap := make(map[uuid.UUID]*TreeEntity, len(locations))

	var rootIds []uuid.UUID

	for _, location := range locations {
		loc := &TreeEntity{
			ID:       location.ID,
			Name:     location.Name,
			Type:     location.Type,
			Children: []*TreeEntity{},
		}

		locationMap[location.ID] = loc
		if location.ParentID != uuid.Nil {
			parent, ok := locationMap[location.ParentID]
			if ok {
				parent.Children = append(parent.Children, loc)
			}
		} else {
			rootIds = append(rootIds, location.ID)
		}
	}

	roots := make([]TreeEntity, 0, len(rootIds))
	for _, id := range rootIds {
		roots = append(roots, *locationMap[id])
	}

	return roots
}
