package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/itemfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/location"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
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
		TagIDs           []uuid.UUID  `json:"tagIds"`
		NegateTags       bool         `json:"negateTags"`
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

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		TagIDs     []uuid.UUID `json:"tagIds"`
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

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		TagIDs     []uuid.UUID `json:"tagIds"`

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
		TagIDs     []uuid.UUID `json:"tagIds"             extensions:"x-nullable,x-omitempty"`
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

		PurchasePrice float64 `json:"purchasePrice"`

		// Edges
		Location *LocationSummary `json:"location,omitempty" extensions:"x-nullable,x-omitempty"`
		Tags     []TagSummary     `json:"tags"`

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

		Attachments []ItemAttachment `json:"attachments"`
		Fields      []ItemField      `json:"fields"`
	}
)

var mapItemsSummaryErr = mapTEachErrFunc(mapItemSummary)

func mapItemSummary(item *ent.Item) ItemSummary {
	var location *LocationSummary
	if item.Edges.Location != nil {
		loc := mapLocationSummary(item.Edges.Location)
		location = &loc
	}

	tags := lo.Ternary(item.Edges.Tag != nil, mapEach(item.Edges.Tag, mapTagSummary), []TagSummary{})

	var imageID *uuid.UUID
	var thumbnailID *uuid.UUID
	if item.Edges.Attachments != nil {
		if a, ok := lo.Find(item.Edges.Attachments, func(a *ent.Attachment) bool {
			return a.Primary && a.Type == attachment.TypePhoto
		}); ok {
			imageID = &a.ID
			if a.Edges.Thumbnail != nil && a.Edges.Thumbnail.ID != uuid.Nil {
				thumbnailID = &a.Edges.Thumbnail.ID
			}
		}
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

		// Edges
		Location: location,
		Tags:     tags,

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

func mapFields(fields []*ent.ItemField) []ItemField {
	return lo.Map(fields, func(f *ent.ItemField, _ int) ItemField {
		return ItemField{
			ID:           f.ID,
			Type:         f.Type.String(),
			Name:         f.Name,
			TextValue:    f.TextValue,
			NumberValue:  f.NumberValue,
			BooleanValue: f.BooleanValue,
			// TimeValue:    f.TimeValue,
		}
	})
}

func mapItemOut(item *ent.Item) ItemOut {
	var attachments []ItemAttachment
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
		SyncChildItemsLocations: item.SyncChildItemsLocations,

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

func (e *ItemsRepository) getOneTx(ctx context.Context, tx *ent.Tx, where ...predicate.Item) (ItemOut, error) {
	var q *ent.ItemQuery
	if tx != nil {
		q = tx.Item.Query().Where(where...)
	} else {
		q = e.db.Item.Query().Where(where...)
	}

	return mapItemOutErr(q.
		WithFields().
		WithTag().
		WithLocation().
		WithGroup().
		WithParent().
		WithAttachments().
		Only(ctx),
	)
}

func (e *ItemsRepository) getOne(ctx context.Context, where ...predicate.Item) (ItemOut, error) {
	return e.getOneTx(ctx, nil, where...)
}

// GetOne returns a single item by ID. If the item does not exist, an error is returned.
// See also: GetOneByGroup to ensure that the item belongs to a specific group.
func (e *ItemsRepository) GetOne(ctx context.Context, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, item.ID(id))
}

func (e *ItemsRepository) CheckRef(ctx context.Context, gid uuid.UUID, ref string) (bool, error) {
	q := e.db.Item.Query().Where(item.HasGroupWith(group.ID(gid)))
	return q.Where(item.ImportRef(ref)).Exist(ctx)
}

func (e *ItemsRepository) GetByRef(ctx context.Context, gid uuid.UUID, ref string) (ItemOut, error) {
	return e.getOne(ctx, item.ImportRef(ref), item.HasGroupWith(group.ID(gid)))
}

// GetOneByGroup returns a single item by ID. If the item does not exist, an error is returned.
// GetOneByGroup ensures that the item belongs to a specific group.
func (e *ItemsRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, item.ID(id), item.HasGroupWith(group.ID(gid)))
}

// QueryByGroup returns a list of items that belong to a specific group based on the provided query.
func (e *ItemsRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q ItemQuery) (PaginationResult[ItemSummary], error) {
	qb := e.db.Item.Query().Where(
		item.HasGroupWith(group.ID(gid)),
	)

	if q.IncludeArchived {
		qb = qb.Where(
			item.Or(
				item.Archived(true),
				item.Archived(false),
			),
		)
	} else {
		qb = qb.Where(item.Archived(false))
	}

	if q.Search != "" {
		// Use accent-insensitive search predicates that normalize both
		// the search query and database field values during comparison.
		// For queries without accents, the traditional search is more efficient.
		qb.Where(
			item.Or(
				// Regular case-insensitive search (fastest)
				item.NameContainsFold(q.Search),
				item.DescriptionContainsFold(q.Search),
				item.SerialNumberContainsFold(q.Search),
				item.ModelNumberContainsFold(q.Search),
				item.ManufacturerContainsFold(q.Search),
				item.NotesContainsFold(q.Search),
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
		qb = qb.Where(item.AssetID(q.AssetID.Int()))
	}

	// Filters within this block define a AND relationship where each subset
	// of filters is OR'd together.
	//
	// The goal is to allow matches like where the item has
	//  - one of the selected tags AND
	//  - one of the selected locations AND
	//  - one of the selected fields key/value matches
	var andPredicates []predicate.Item
	{
		if len(q.TagIDs) > 0 {
			var tagPredicates []predicate.Item
			if !q.NegateTags {
				tagPredicates = lo.Map(q.TagIDs, func(l uuid.UUID, _ int) predicate.Item {
					return item.HasTagWith(tag.ID(l))
				})
				andPredicates = append(andPredicates, item.Or(tagPredicates...))
			} else {
				tagPredicates = lo.Map(q.TagIDs, func(l uuid.UUID, _ int) predicate.Item {
					return item.Not(item.HasTagWith(tag.ID(l)))
				})
				andPredicates = append(andPredicates, item.And(tagPredicates...))
			}
		}

		if q.OnlyWithoutPhoto {
			andPredicates = append(andPredicates, item.Not(
				item.HasAttachmentsWith(
					attachment.And(
						attachment.Primary(true),
						attachment.TypeEQ(attachment.TypePhoto),
					),
				)),
			)
		}

		if q.OnlyWithPhoto {
			andPredicates = append(andPredicates, item.HasAttachmentsWith(
				attachment.And(
					attachment.Primary(true),
					attachment.TypeEQ(attachment.TypePhoto),
				),
			),
			)
		}

		if len(q.LocationIDs) > 0 {
			locationPredicates := lo.Map(q.LocationIDs, func(l uuid.UUID, _ int) predicate.Item {
				return item.HasLocationWith(location.ID(l))
			})

			andPredicates = append(andPredicates, item.Or(locationPredicates...))
		}

		if len(q.Fields) > 0 {
			fieldPredicates := lo.Map(q.Fields, func(f FieldQuery, _ int) predicate.Item {
				return item.HasFieldsWith(
					itemfield.And(
						itemfield.Name(f.Name),
						itemfield.TextValue(f.Value),
					),
				)
			})

			andPredicates = append(andPredicates, item.Or(fieldPredicates...))
		}

		if len(q.ParentItemIDs) > 0 {
			andPredicates = append(andPredicates, item.HasParentWith(item.IDIn(q.ParentItemIDs...)))
		}
	}

	if len(andPredicates) > 0 {
		qb = qb.Where(item.And(andPredicates...))
	}

	count, err := qb.Count(ctx)
	if err != nil {
		return PaginationResult[ItemSummary]{}, err
	}

	// Order
	switch q.OrderBy {
	case "createdAt":
		qb = qb.Order(ent.Desc(item.FieldCreatedAt))
	case "updatedAt":
		qb = qb.Order(ent.Desc(item.FieldUpdatedAt))
	case "assetId":
		qb = qb.Order(ent.Asc(item.FieldAssetID))
	default: // "name"
		qb = qb.Order(ent.Asc(item.FieldName))
	}

	qb = qb.
		WithTag().
		WithLocation().
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
	qb := e.db.Item.Query().Where(
		item.HasGroupWith(group.ID(gid)),
		item.AssetID(int(assetID)),
	)

	if page != -1 || pageSize != -1 {
		qb.Offset(calculateOffset(page, pageSize)).
			Limit(pageSize)
	} else {
		page = -1
		pageSize = -1
	}

	items, err := mapItemsSummaryErr(
		qb.Order(ent.Asc(item.FieldName)).
			WithTag().
			WithLocation().
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

// GetAll returns all the items in the database with the Tags and Locations eager loaded.
func (e *ItemsRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]ItemOut, error) {
	return mapItemsOutErr(e.db.Item.Query().
		Where(item.HasGroupWith(group.ID(gid))).
		WithTag().
		WithLocation().
		WithFields().
		All(ctx))
}

func (e *ItemsRepository) GetAllZeroAssetID(ctx context.Context, gid uuid.UUID) ([]ItemSummary, error) {
	q := e.db.Item.Query().Where(
		item.HasGroupWith(group.ID(gid)),
		item.AssetID(0),
	).Order(
		ent.Asc(item.FieldCreatedAt),
	)

	return mapItemsSummaryErr(q.All(ctx))
}

func (e *ItemsRepository) GetHighestAssetIDTx(ctx context.Context, tx *ent.Tx, gid uuid.UUID) (AssetID, error) {
	var q *ent.ItemQuery
	if tx != nil {
		q = tx.Item.Query().Where(
			item.HasGroupWith(group.ID(gid)),
		).Order(
			ent.Desc(item.FieldAssetID),
		).Limit(1)
	} else {
		q = e.db.Item.Query().Where(
			item.HasGroupWith(group.ID(gid)),
		).Order(
			ent.Desc(item.FieldAssetID),
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
	q := e.db.Item.Update().Where(
		item.HasGroupWith(group.ID(gid)),
		item.ID(id),
	)

	_, err := q.SetAssetID(int(assetID)).Save(ctx)
	return err
}

func (e *ItemsRepository) Create(ctx context.Context, gid uuid.UUID, data ItemCreate) (ItemOut, error) {
	q := e.db.Item.Create().
		SetImportRef(data.ImportRef).
		SetName(data.Name).
		SetQuantity(data.Quantity).
		SetDescription(data.Description).
		SetGroupID(gid).
		SetLocationID(data.LocationID).
		SetAssetID(int(data.AssetID))

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if len(data.TagIDs) > 0 {
		q.AddTagIDs(data.TagIDs...)
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
	TagIDs           []uuid.UUID
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
	itemBuilder := tx.Item.Create().
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

	if len(data.TagIDs) > 0 {
		itemBuilder.AddTagIDs(data.TagIDs...)
	}

	_, err = itemBuilder.Save(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	// Create custom fields
	for _, field := range data.Fields {
		_, err = tx.ItemField.Create().
			SetItemID(newItemID).
			SetType(itemfield.Type(field.Type)).
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
	itm, err := e.db.Item.Query().
		Where(item.ID(id)).
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
		err := e.attachments.Delete(ctx, gid, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during item deletion")
			// Continue with other attachments even if one fails
		}
	}

	err = e.db.Item.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	e.publishMutationEvent(id)
	return nil
}

func (e *ItemsRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	// Get the item with its attachments before deletion
	itm, err := e.db.Item.Query().
		Where(
			item.ID(id),
			item.HasGroupWith(group.ID(gid)),
		).
		WithAttachments().
		Only(ctx)
	if err != nil {
		return err
	}

	// Delete all attachments (and their files) before deleting the item
	for _, att := range itm.Edges.Attachments {
		err := e.attachments.Delete(ctx, gid, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during item deletion")
			// Continue with other attachments even if one fails
		}
	}

	_, err = e.db.Item.
		Delete().
		Where(
			item.ID(id),
			item.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	e.publishMutationEvent(gid)
	return err
}

func (e *ItemsRepository) WipeInventory(ctx context.Context, gid uuid.UUID, wipeTags bool, wipeLocations bool, wipeMaintenance bool) (int, error) {
	deleted := 0

	// Wipe maintenance records if requested
	// IMPORTANT: Must delete maintenance records BEFORE items since they are linked to items
	if wipeMaintenance {
		maintenanceCount, err := e.db.MaintenanceEntry.Delete().
			Where(maintenanceentry.HasItemWith(item.HasGroupWith(group.ID(gid)))).
			Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete maintenance entries during wipe inventory")
		} else {
			log.Info().Int("count", maintenanceCount).Msg("deleted maintenance entries during wipe inventory")
			deleted += maintenanceCount
		}
	}

	// Get all items for the group
	items, err := e.db.Item.Query().
		Where(item.HasGroupWith(group.ID(gid))).
		WithAttachments().
		All(ctx)
	if err != nil {
		return 0, err
	}

	// Delete each item with its attachments
	// Note: We manually delete attachments and items instead of calling DeleteByGroup
	// to continue processing remaining items even if some deletions fail
	for _, itm := range items {
		// Delete all attachments first
		for _, att := range itm.Edges.Attachments {
			err := e.attachments.Delete(ctx, gid, att.ID)
			if err != nil {
				log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during wipe inventory")
				// Continue with other attachments even if one fails
			}
		}

		// Delete the item
		_, err = e.db.Item.
			Delete().
			Where(
				item.ID(itm.ID),
				item.HasGroupWith(group.ID(gid)),
			).Exec(ctx)
		if err != nil {
			log.Err(err).Str("item_id", itm.ID.String()).Msg("failed to delete item during wipe inventory")
			// Skip to next item without incrementing counter
			continue
		}

		// Only increment counter if deletion succeeded
		deleted++
	}

	// Wipe tags if requested
	if wipeTags {
		tagCount, err := e.db.Tag.Delete().Where(tag.HasGroupWith(group.ID(gid))).Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete tags during wipe inventory")
		} else {
			log.Info().Int("count", tagCount).Msg("deleted tags during wipe inventory")
			deleted += tagCount
		}
	}

	// Wipe locations if requested
	if wipeLocations {
		locationCount, err := e.db.Location.Delete().Where(location.HasGroupWith(group.ID(gid))).Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete locations during wipe inventory")
		} else {
			log.Info().Int("count", locationCount).Msg("deleted locations during wipe inventory")
			deleted += locationCount
		}
	}

	e.publishMutationEvent(gid)
	return deleted, nil
}

func (e *ItemsRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data ItemUpdate) (ItemOut, error) {
	q := e.db.Item.Update().Where(item.ID(data.ID), item.HasGroupWith(group.ID(gid))).
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
		SetSyncChildItemsLocations(data.SyncChildItemsLocations)

	currentTags, err := e.db.Item.Query().Where(item.ID(data.ID)).QueryTag().All(ctx)
	if err != nil {
		return ItemOut{}, err
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
		children, err := e.db.Item.Query().Where(item.ID(data.ID)).QueryChildren().All(ctx)
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

	fields, err := e.db.ItemField.Query().Where(itemfield.HasItemWith(item.ID(data.ID))).All(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	fieldIds := newIDSet(fields)

	// Update Existing Fields
	for _, f := range data.Fields {
		if f.ID == uuid.Nil {
			// Create New Field
			_, err = e.db.ItemField.Create().
				SetItemID(data.ID).
				SetType(itemfield.Type(f.Type)).
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

		opt := e.db.ItemField.Update().
			Where(
				itemfield.ID(f.ID),
				itemfield.HasItemWith(item.ID(data.ID)),
			).
			SetType(itemfield.Type(f.Type)).
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
		_, err = e.db.ItemField.Delete().
			Where(
				itemfield.IDIn(fieldIds.Slice()...),
				itemfield.HasItemWith(item.ID(data.ID)),
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

	err := e.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
			item.Or(
				item.ImportRefEQ(""),
				item.ImportRefIsNil(),
			),
		).
		Select(item.FieldID).
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

	q := tx.Item.Update().
		Where(
			item.ID(id),
			item.HasGroupWith(group.ID(gid)),
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

	if data.TagIDs != nil {
		currentTags, err := tx.Item.Query().Where(item.ID(id), item.HasGroupWith(group.ID(gid))).QueryTag().All(ctx)
		if err != nil {
			return err
		}
		set := newIDSet(currentTags)

		addTags := []uuid.UUID{}
		for _, l := range data.TagIDs {
			if set.Contains(l) {
				set.Remove(l)
			} else {
				addTags = append(addTags, l)
			}
		}

		if len(addTags) > 0 {
			if err := tx.Item.Update().
				Where(item.ID(id), item.HasGroupWith(group.ID(gid))).
				AddTagIDs(addTags...).
				Exec(ctx); err != nil {
				return err
			}
		}
		if set.Len() > 0 {
			if err := tx.Item.Update().
				Where(item.ID(id), item.HasGroupWith(group.ID(gid))).
				RemoveTagIDs(set.Slice()...).
				Exec(ctx); err != nil {
				return err
			}
		}
	}

	if data.LocationID != uuid.Nil {
		itemEnt, err := tx.Item.Query().Where(item.ID(id), item.HasGroupWith(group.ID(gid))).Only(ctx)
		if err != nil {
			return err
		}
		if itemEnt.SyncChildItemsLocations {
			children, err := tx.Item.Query().Where(item.ID(id), item.HasGroupWith(group.ID(gid))).QueryChildren().All(ctx)
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

	err := e.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
		).
		QueryFields().
		Where(
			itemfield.Name(name),
		).
		Unique(true).
		Select(itemfield.FieldTextValue).
		Scan(ctx, &values)
	if err != nil {
		return nil, fmt.Errorf("failed to get field values: %w", err)
	}

	valueStrings := lo.Map(values, func(f st, _ int) string {
		return f.Value
	})

	return valueStrings, nil
}

func (e *ItemsRepository) GetAllCustomFieldNames(ctx context.Context, gid uuid.UUID) ([]string, error) {
	type st struct {
		Name string `json:"name"`
	}

	var fields []st

	err := e.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
		).
		QueryFields().
		Unique(true).
		Select(itemfield.FieldName).
		Scan(ctx, &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom fields: %w", err)
	}

	fieldNames := lo.Map(fields, func(f st, _ int) string {
		return f.Name
	})

	return fieldNames, nil
}

// ZeroOutTimeFields is a helper function that can be invoked via the UI by a group member which will
// set all date fields to the beginning of the day.
//
// This is designed to resolve a long-time bug that has since been fixed with the time selector on the
// frontend. This function is intended to be used as a one-time fix for existing databases and may be
// removed in the future.
func (e *ItemsRepository) ZeroOutTimeFields(ctx context.Context, gid uuid.UUID) (int, error) {
	q := e.db.Item.Query().Where(
		item.HasGroupWith(group.ID(gid)),
		item.Or(
			item.PurchaseTimeNotNil(),
			item.PurchaseFromLT("0002-01-01"),
			item.SoldTimeNotNil(),
			item.SoldToLT("0002-01-01"),
			item.WarrantyExpiresNotNil(),
			item.WarrantyDetailsLT("0002-01-01"),
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
		updateQ := e.db.Item.Update().Where(item.ID(i.ID))

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
	itemIDs, err := e.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
			item.HasAttachmentsWith(
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
				attachment.HasItemWith(item.ID(id)),
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
	originalItem, err := e.getOneTx(ctx, tx, item.ID(id), item.HasGroupWith(group.ID(gid)))
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
	itemBuilder := tx.Item.Create().
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
		SetSyncChildItemsLocations(originalItem.SyncChildItemsLocations)

	if originalItem.Parent != nil {
		itemBuilder.SetParentID(originalItem.Parent.ID)
	}

	// Add tags
	if len(originalItem.Tags) > 0 {
		tagIDs := lo.Map(originalItem.Tags, func(tag TagSummary, _ int) uuid.UUID {
			return tag.ID
		})
		itemBuilder.AddTagIDs(tagIDs...)
	}

	_, err = itemBuilder.Save(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	// Copy custom fields if requested
	if options.CopyCustomFields {
		for _, field := range originalItem.Fields {
			_, err = tx.ItemField.Create().
				SetItemID(newItemID).
				SetType(itemfield.Type(field.Type)).
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
				SetItemID(newItemID).
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
			Where(maintenanceentry.HasItemWith(item.ID(id))).
			All(ctx)
		if err == nil {
			for _, entry := range maintenanceEntries {
				_, err = tx.MaintenanceEntry.Create().
					SetItemID(newItemID).
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
