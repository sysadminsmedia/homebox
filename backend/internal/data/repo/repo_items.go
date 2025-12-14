package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entityfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/label"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

type ItemsRepository struct {
	db          *ent.Client
	bus         *eventbus.EventBus
	attachments *AttachmentRepo
}

type (
	FieldQuery struct {
		Name  string
		Value string
	}

	ItemQuery struct {
		Page             int
		PageSize         int
		Search           string       `json:"search"`
		AssetID          AssetID      `json:"assetId"`
		LocationIDs      []uuid.UUID  `json:"locationIds"`
		LabelIDs         []uuid.UUID  `json:"labelIds"`
		NegateLabels     bool         `json:"negateLabels"`
		OnlyWithoutPhoto bool         `json:"onlyWithoutPhoto"`
		OnlyWithPhoto    bool         `json:"onlyWithPhoto"`
		ParentItemIDs    []uuid.UUID  `json:"parentIds"`
		SortBy           string       `json:"sortBy"`
		IncludeArchived  bool         `json:"includeArchived"`
		Fields           []FieldQuery `json:"fields"`
		OrderBy          string       `json:"orderBy"`
	}

	DuplicateOptions struct {
		CopyMaintenance  bool   `json:"copyMaintenance"`
		CopyAttachments  bool   `json:"copyAttachments"`
		CopyCustomFields bool   `json:"copyCustomFields"`
		CopyPrefix       string `json:"copyPrefix"`
	}

	ItemField struct {
		ID           uuid.UUID `json:"id,omitempty"`
		Type         string    `json:"type"`
		Name         string    `json:"name"`
		TextValue    string    `json:"textValue"`
		NumberValue  int       `json:"numberValue"`
		BooleanValue bool      `json:"booleanValue"`
		// TimeValue    time.Time `json:"timeValue,omitempty"`
	}

	ItemCreate struct {
		ImportRef   string    `json:"-"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Quantity    int       `json:"quantity"`
		Description string    `json:"description" validate:"max=1000"`
		AssetID     AssetID   `json:"-"`
		EntityType  uuid.UUID `json:"entityType"  validate:"required,uuid"`

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		LabelIDs   []uuid.UUID `json:"labelIds"`
	}

	ItemUpdate struct {
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
		LabelIDs   []uuid.UUID `json:"labelIds"`

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
		Notes  string      `json:"notes"`
		Fields []ItemField `json:"fields"`
	}

	ItemPatch struct {
		ID         uuid.UUID   `json:"id"`
		Quantity   *int        `json:"quantity,omitempty" extensions:"x-nullable,x-omitempty"`
		ImportRef  *string     `json:"-,omitempty"        extensions:"x-nullable,x-omitempty"`
		LocationID uuid.UUID   `json:"locationId"         extensions:"x-nullable,x-omitempty"`
		LabelIDs   []uuid.UUID `json:"labelIds"           extensions:"x-nullable,x-omitempty"`
	}

	ItemSummary struct {
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
		Labels   []LabelSummary   `json:"labels"`

		ImageID     *uuid.UUID `json:"imageId,omitempty"     extensions:"x-nullable,x-omitempty"`
		ThumbnailId *uuid.UUID `json:"thumbnailId,omitempty" extensions:"x-nullable,x-omitempty"`

		// Sale details
		SoldTime time.Time `json:"soldTime"`
	}

	ItemOut struct {
		Parent *ItemSummary `json:"parent,omitempty" extensions:"x-nullable,x-omitempty"`
		ItemSummary
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
		Fields      []ItemField        `json:"fields"`
	}
)

var mapItemsSummaryErr = mapTEachErrFunc(mapItemSummary)

func mapItemSummary(item *ent.Entity) ItemSummary {
	var location *LocationSummary
	if item.Edges.Location != nil {
		loc := mapLocationSummary(item.QueryLocation().Where(entity.HasTypeWith(entitytype.IsLocationEQ(true))).FirstX(context.Background()))
		location = &loc
	}

	labels := make([]LabelSummary, len(item.Edges.Label))
	if item.Edges.Label != nil {
		labels = mapEach(item.Edges.Label, mapLabelSummary)
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

	return ItemSummary{
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
		Labels:   labels,

		// Warranty
		Insured:     item.Insured,
		ImageID:     imageID,
		ThumbnailId: thumbnailID,
	}
}

var (
	mapItemOutErr  = mapTErrFunc(mapItemOut)
	mapItemsOutErr = mapTEachErrFunc(mapItemOut)
)

func mapFields(fields []*ent.EntityField) []ItemField {
	result := make([]ItemField, len(fields))
	for i, f := range fields {
		result[i] = ItemField{
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

func mapItemOut(item *ent.Entity) ItemOut {
	var attachments []EntityAttachment
	if item.Edges.Attachments != nil {
		attachments = mapEach(item.Edges.Attachments, ToItemAttachment)
	}

	var fields []ItemField
	if item.Edges.Fields != nil {
		fields = mapFields(item.Edges.Fields)
	}

	var parent *ItemSummary
	if item.Edges.Parent != nil {
		v := mapItemSummary(item.Edges.Parent)
		parent = &v
	}

	return ItemOut{
		Parent:                  parent,
		AssetID:                 AssetID(item.AssetID),
		ItemSummary:             mapItemSummary(item),
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

func (e *ItemsRepository) publishMutationEvent(gid uuid.UUID) {
	if e.bus != nil {
		e.bus.Publish(eventbus.EventItemMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (e *ItemsRepository) getOneTx(ctx context.Context, tx *ent.Tx, where ...predicate.Entity) (ItemOut, error) {
	var q *ent.EntityQuery
	if tx != nil {
		q = tx.Entity.Query().Where(where...)
	} else {
		q = e.db.Entity.Query().Where(where...)
	}

	return mapItemOutErr(q.
		WithFields().
		WithLabel().
		WithLocation().
		WithGroup().
		WithParent().
		WithAttachments().
		WithType().
		Only(ctx),
	)
}

func (e *ItemsRepository) getOne(ctx context.Context, where ...predicate.Entity) (ItemOut, error) {
	return e.getOneTx(ctx, nil, where...)
}

// GetOne returns a single item by ID. If the item does not exist, an error is returned.
// See also: GetOneByGroup to ensure that the item belongs to a specific group.
func (e *ItemsRepository) GetOne(ctx context.Context, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, entity.ID(id))
}

func (e *ItemsRepository) CheckRef(ctx context.Context, gid uuid.UUID, ref string) (bool, error) {
	q := e.db.Entity.Query().Where(entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
	return q.Where(entity.ImportRef(ref)).Exist(ctx)
}

func (e *ItemsRepository) GetByRef(ctx context.Context, gid uuid.UUID, ref string) (ItemOut, error) {
	return e.getOne(ctx, entity.ImportRef(ref), entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
}

// GetOneByGroup returns a single item by ID. If the item does not exist, an error is returned.
// GetOneByGroup ensures that the item belongs to a specific group.
func (e *ItemsRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, entity.ID(id), entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false)))
}

// QueryByGroup returns a list of items that belong to a specific group based on the provided query.
func (e *ItemsRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q ItemQuery) (PaginationResult[ItemSummary], error) {
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
		if len(q.LabelIDs) > 0 {
			labelPredicates := make([]predicate.Entity, 0, len(q.LabelIDs))
			for _, l := range q.LabelIDs {
				if !q.NegateLabels {
					labelPredicates = append(labelPredicates, entity.HasLabelWith(label.ID(l)))
				} else {
					labelPredicates = append(labelPredicates, entity.Not(entity.HasLabelWith(label.ID(l))))
				}
			}
			if !q.NegateLabels {
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
		return PaginationResult[ItemSummary]{}, err
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
		WithLabel().
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

	items, err := mapItemsSummaryErr(qb.All(ctx))
	if err != nil {
		return PaginationResult[ItemSummary]{}, err
	}

	return PaginationResult[ItemSummary]{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    count,
		Items:    items,
	}, nil
}

// QueryByAssetID returns items by asset ID. If the item does not exist, an error is returned.
func (e *ItemsRepository) QueryByAssetID(ctx context.Context, gid uuid.UUID, assetID AssetID, page int, pageSize int) (PaginationResult[ItemSummary], error) {
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

	items, err := mapItemsSummaryErr(
		qb.Order(ent.Asc(entity.FieldName)).
			WithLabel().
			WithLocation().
			WithType().
			All(ctx),
	)
	if err != nil {
		return PaginationResult[ItemSummary]{}, err
	}

	return PaginationResult[ItemSummary]{
		Page:     page,
		PageSize: pageSize,
		Total:    len(items),
		Items:    items,
	}, nil
}

// GetAll returns all the items in the database with the Labels and Locations eager loaded.
func (e *ItemsRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]ItemOut, error) {
	return mapItemsOutErr(e.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid)), entity.HasTypeWith(entitytype.IsLocationEQ(false))).
		WithLabel().
		WithLocation().
		WithFields().
		WithType().
		All(ctx))
}

func (e *ItemsRepository) GetAllZeroAssetID(ctx context.Context, gid uuid.UUID) ([]ItemSummary, error) {
	q := e.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(0),
	).Order(
		ent.Asc(entity.FieldCreatedAt),
	)

	return mapItemsSummaryErr(q.All(ctx))
}

func (e *ItemsRepository) GetHighestAssetIDTx(ctx context.Context, tx *ent.Tx, gid uuid.UUID) (AssetID, error) {
	var q *ent.EntityQuery
	if tx != nil {
		q = tx.Entity.Query().Where(
			entity.HasGroupWith(group.ID(gid)),
		).Order(
			ent.Desc(entity.FieldAssetID),
		).Limit(1)
	} else {
		q = e.db.Entity.Query().Where(
			entity.HasGroupWith(group.ID(gid)),
		).Order(
			ent.Desc(entity.FieldAssetID),
		).Limit(1)
	}

	result, err := q.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}

	return AssetID(result.AssetID), nil
}

func (e *ItemsRepository) GetHighestAssetID(ctx context.Context, gid uuid.UUID) (AssetID, error) {
	return e.GetHighestAssetIDTx(ctx, nil, gid)
}

func (e *ItemsRepository) SetAssetID(ctx context.Context, gid uuid.UUID, id uuid.UUID, assetID AssetID) error {
	q := e.db.Entity.Update().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.ID(id),
	)

	_, err := q.SetAssetID(int(assetID)).Save(ctx)
	return err
}

func (e *ItemsRepository) Create(ctx context.Context, gid uuid.UUID, data ItemCreate) (ItemOut, error) {
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

	if len(data.LabelIDs) > 0 {
		q.AddLabelIDs(data.LabelIDs...)
	}

	result, err := q.Save(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	e.publishMutationEvent(gid)
	return e.GetOne(ctx, result.ID)
}

// ItemCreateFromTemplate contains all data needed to create an item from a template.
type ItemCreateFromTemplate struct {
	Name             string
	Description      string
	Quantity         int
	LocationID       uuid.UUID
	LabelIDs         []uuid.UUID
	Insured          bool
	Manufacturer     string
	ModelNumber      string
	LifetimeWarranty bool
	WarrantyDetails  string
	Fields           []ItemField
}

// CreateFromTemplate creates an item with all template data in a single transaction.
func (e *ItemsRepository) CreateFromTemplate(ctx context.Context, gid uuid.UUID, data ItemCreateFromTemplate) (ItemOut, error) {
	tx, err := e.db.Tx(ctx)
	if err != nil {
		return ItemOut{}, err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during template item creation")
			}
		}
	}()

	// Get next asset ID within transaction
	nextAssetID, err := e.GetHighestAssetIDTx(ctx, tx, gid)
	if err != nil {
		return ItemOut{}, err
	}
	nextAssetID++

	// Create item with all template data
	newItemID := uuid.New()
	itemBuilder := tx.Entity.Create().
		SetID(newItemID).
		SetName(data.Name).
		SetDescription(data.Description).
		SetQuantity(data.Quantity).
		SetLocationID(data.LocationID).
		SetGroupID(gid).
		SetAssetID(int(nextAssetID)).
		SetInsured(data.Insured).
		SetManufacturer(data.Manufacturer).
		SetModelNumber(data.ModelNumber).
		SetLifetimeWarranty(data.LifetimeWarranty).
		SetWarrantyDetails(data.WarrantyDetails)

	if len(data.LabelIDs) > 0 {
		itemBuilder.AddLabelIDs(data.LabelIDs...)
	}

	_, err = itemBuilder.Save(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	// Create custom fields
	for _, field := range data.Fields {
		_, err = tx.EntityField.Create().
			SetEntityID(newItemID).
			SetType(entityfield.Type(field.Type)).
			SetName(field.Name).
			SetTextValue(field.TextValue).
			Save(ctx)
		if err != nil {
			return ItemOut{}, fmt.Errorf("failed to create field %s: %w", field.Name, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return ItemOut{}, err
	}
	committed = true

	e.publishMutationEvent(gid)
	return e.GetOne(ctx, newItemID)
}

func (e *ItemsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get the item with its group and attachments before deletion
	itm, err := e.db.Entity.Query().
		Where(entity.ID(id)).
		WithGroup().
		WithAttachments().
		Only(ctx)
	if err != nil {
		return err
	}

	// Get the group ID for attachment deletion
	var gid uuid.UUID
	if itm.Edges.Group != nil {
		gid = itm.Edges.Group.ID
	}

	// Delete all attachments (and their files) before deleting the item
	for _, att := range itm.Edges.Attachments {
		err := e.attachments.Delete(ctx, gid, id, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during item deletion")
			// Continue with other attachments even if one fails
		}
	}

	err = e.db.Entity.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	e.publishMutationEvent(id)
	return nil
}

func (e *ItemsRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	// Get the item with its attachments before deletion
	itm, err := e.db.Entity.Query().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).
		WithAttachments().
		Only(ctx)
	if err != nil {
		return err
	}

	// Delete all attachments (and their files) before deleting the item
	for _, att := range itm.Edges.Attachments {
		err := e.attachments.Delete(ctx, gid, id, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during item deletion")
			// Continue with other attachments even if one fails
		}
	}

	_, err = e.db.Entity.
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

func (e *ItemsRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data ItemUpdate) (ItemOut, error) {
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

	currentLabels, err := e.db.Entity.Query().Where(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))).QueryLabel().All(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	set := newIDSet(currentLabels)

	for _, l := range data.LabelIDs {
		if set.Contains(l) {
			set.Remove(l)
			continue
		}
		q.AddLabelIDs(l)
	}

	if set.Len() > 0 {
		q.RemoveLabelIDs(set.Slice()...)
	}

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	if data.SyncChildItemsLocations {
		children, err := e.db.Entity.Query().Where(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false))).QueryChildren().All(ctx)
		if err != nil {
			return ItemOut{}, err
		}
		location := data.LocationID

		for _, child := range children {
			childLocation, err := child.QueryLocation().First(ctx)
			if err != nil {
				return ItemOut{}, err
			}

			if location != childLocation.ID {
				err = child.Update().SetLocationID(location).Exec(ctx)
				if err != nil {
					return ItemOut{}, err
				}
			}
		}
	}

	err = q.Exec(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	fields, err := e.db.EntityField.Query().Where(entityfield.HasEntityWith(entity.ID(data.ID), entity.HasTypeWith(entitytype.IsLocationEQ(false)))).All(ctx)
	if err != nil {
		return ItemOut{}, err
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
				return ItemOut{}, err
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
			return ItemOut{}, err
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
			return ItemOut{}, err
		}
	}

	e.publishMutationEvent(gid)
	return e.GetOne(ctx, data.ID)
}

func (e *ItemsRepository) GetAllZeroImportRef(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
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

func (e *ItemsRepository) Patch(ctx context.Context, gid, id uuid.UUID, data ItemPatch) error {
	tx, err := e.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during item patch")
			}
		}
	}()

	q := tx.Entity.Update().
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

	if data.LocationID != uuid.Nil {
		q.SetLocationID(data.LocationID)
	}

	err = q.Exec(ctx)
	if err != nil {
		return err
	}

	if data.LabelIDs != nil {
		currentLabels, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryLabel().All(ctx)
		if err != nil {
			return err
		}
		set := newIDSet(currentLabels)

		addLabels := []uuid.UUID{}
		for _, l := range data.LabelIDs {
			if set.Contains(l) {
				set.Remove(l)
			} else {
				addLabels = append(addLabels, l)
			}
		}

		if len(addLabels) > 0 {
			if err := tx.Entity.Update().
				Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
				AddLabelIDs(addLabels...).
				Exec(ctx); err != nil {
				return err
			}
		}
		if set.Len() > 0 {
			if err := tx.Entity.Update().
				Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
				RemoveLabelIDs(set.Slice()...).
				Exec(ctx); err != nil {
				return err
			}
		}
	}

	if data.LocationID != uuid.Nil {
		itemEnt, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).Only(ctx)
		if err != nil {
			return err
		}
		if itemEnt.SyncChildEntitiesLocations {
			children, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryChildren().All(ctx)
			if err != nil {
				return err
			}
			for _, child := range children {
				childLocation, err := child.QueryLocation().First(ctx)
				if err != nil {
					return err
				}
				if data.LocationID != childLocation.ID {
					err = child.Update().SetLocationID(data.LocationID).Exec(ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	e.publishMutationEvent(gid)
	return nil
}

func (e *ItemsRepository) GetAllCustomFieldValues(ctx context.Context, gid uuid.UUID, name string) ([]string, error) {
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

func (e *ItemsRepository) GetAllCustomFieldNames(ctx context.Context, gid uuid.UUID) ([]string, error) {
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
func (e *ItemsRepository) ZeroOutTimeFields(ctx context.Context, gid uuid.UUID) (int, error) {
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

func (e *ItemsRepository) SetPrimaryPhotos(ctx context.Context, gid uuid.UUID) (int, error) {
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
func (e *ItemsRepository) Duplicate(ctx context.Context, gid, id uuid.UUID, options DuplicateOptions) (ItemOut, error) {
	tx, err := e.db.Tx(ctx)
	if err != nil {
		return ItemOut{}, err
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
	originalItem, err := e.getOneTx(ctx, tx, entity.ID(id), entity.HasGroupWith(group.ID(gid)))
	if err != nil {
		return ItemOut{}, err
	}

	nextAssetID, err := e.GetHighestAssetIDTx(ctx, tx, gid)
	if err != nil {
		return ItemOut{}, err
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
	if len(originalItem.Labels) > 0 {
		labelIDs := make([]uuid.UUID, len(originalItem.Labels))
		for i, label := range originalItem.Labels {
			labelIDs[i] = label.ID
		}
		itemBuilder.AddLabelIDs(labelIDs...)
	}

	_, err = itemBuilder.Save(ctx)
	if err != nil {
		return ItemOut{}, err
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
		return ItemOut{}, err
	}
	committed = true

	e.publishMutationEvent(gid)

	// Get the final item with all copied data
	return e.GetOne(ctx, newItemID)
}
