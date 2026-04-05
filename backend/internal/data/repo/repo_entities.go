package repo

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
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
)

type EntityRepository struct {
	db          *ent.Client
	bus         *eventbus.EventBus
	attachments *AttachmentRepo
}

type (
	FieldQuery struct {
		Name  string
		Value string
	}

	EntityQuery struct {
		Page             int
		PageSize         int
		Search           string       `json:"search"`
		AssetID          AssetID      `json:"assetId"`
		ParentIDs        []uuid.UUID  `json:"parentIds"`
		TagIDs           []uuid.UUID  `json:"tagIds"`
		NegateTags       bool         `json:"negateTags"`
		OnlyWithoutPhoto bool         `json:"onlyWithoutPhoto"`
		OnlyWithPhoto    bool         `json:"onlyWithPhoto"`
		ParentItemIDs    []uuid.UUID  `json:"parentItemIds"`
		SortBy           string       `json:"sortBy"`
		IncludeArchived  bool         `json:"includeArchived"`
		IsLocation       *bool        `json:"isLocation"` // nil=all, true=locations only, false=items only
		Fields           []FieldQuery `json:"fields"`
		OrderBy          string       `json:"orderBy"`
	}

	DuplicateOptions struct {
		CopyMaintenance  bool   `json:"copyMaintenance"`
		CopyAttachments  bool   `json:"copyAttachments"`
		CopyCustomFields bool   `json:"copyCustomFields"`
		CopyPrefix       string `json:"copyPrefix"`
	}

	EntityFieldData struct {
		ID           uuid.UUID `json:"id,omitempty"`
		Type         string    `json:"type"`
		Name         string    `json:"name"`
		TextValue    string    `json:"textValue"`
		NumberValue  int       `json:"numberValue"`
		BooleanValue bool      `json:"booleanValue"`
	}

	EntityCreate struct {
		ImportRef    string    `json:"-"`
		ParentID     uuid.UUID `json:"parentId"       extensions:"x-nullable"`
		Name         string    `json:"name"           validate:"required,min=1,max=255"`
		Quantity     float64   `json:"quantity"`
		Description  string    `json:"description"    validate:"max=1000"`
		AssetID      AssetID   `json:"-"`
		EntityTypeID uuid.UUID `json:"entityTypeId"`

		// Edges
		TagIDs []uuid.UUID `json:"tagIds"`
	}

	EntityUpdate struct {
		ParentID                 uuid.UUID `json:"parentId"                    extensions:"x-nullable,x-omitempty"`
		ID                       uuid.UUID `json:"id"`
		AssetID                  AssetID   `json:"assetId"                     swaggertype:"string"`
		Name                     string    `json:"name"                        validate:"required,min=1,max=255"`
		Description              string    `json:"description"                 validate:"max=1000"`
		Quantity                 float64   `json:"quantity"`
		Insured                  bool      `json:"insured"`
		Archived                 bool      `json:"archived"`
		SyncChildEntityLocations bool      `json:"syncChildEntityLocations"`
		EntityTypeID             uuid.UUID `json:"entityTypeId"`

		// Edges
		TagIDs []uuid.UUID `json:"tagIds"`

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
		Notes  string            `json:"notes"`
		Fields []EntityFieldData `json:"fields"`
	}

	EntityPatch struct {
		ID           uuid.UUID   `json:"id"`
		Quantity     *float64    `json:"quantity,omitempty"     extensions:"x-nullable,x-omitempty"`
		ImportRef    *string     `json:"-"                      extensions:"x-nullable,x-omitempty"`
		ParentID     uuid.UUID   `json:"parentId"               extensions:"x-nullable,x-omitempty"`
		EntityTypeID uuid.UUID   `json:"entityTypeId"           extensions:"x-nullable,x-omitempty"`
		TagIDs       []uuid.UUID `json:"tagIds"                 extensions:"x-nullable,x-omitempty"`
	}

	EntitySummary struct {
		ImportRef   string    `json:"-"`
		ID          uuid.UUID `json:"id"`
		AssetID     AssetID   `json:"assetId,string"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Quantity    float64   `json:"quantity"`
		Insured     bool      `json:"insured"`
		Archived    bool      `json:"archived"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`

		PurchasePrice float64 `json:"purchasePrice"`

		// Edges
		Parent     *EntitySummary     `json:"parent,omitempty"     extensions:"x-nullable,x-omitempty"`
		EntityType *EntityTypeSummary `json:"entityType,omitempty" extensions:"x-nullable,x-omitempty"`
		Tags       []TagSummary       `json:"tags"`

		ImageID     *uuid.UUID `json:"imageId,omitempty"     extensions:"x-nullable,x-omitempty"`
		ThumbnailId *uuid.UUID `json:"thumbnailId,omitempty" extensions:"x-nullable,x-omitempty"`

		// Sale details
		SoldTime time.Time `json:"soldTime"`
	}

	EntityOut struct {
		Parent *EntitySummary `json:"parent,omitempty" extensions:"x-nullable,x-omitempty"`
		EntitySummary
		AssetID AssetID `json:"assetId,string"`

		SyncChildEntityLocations bool `json:"syncChildEntityLocations"`

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

		Attachments []ItemAttachment  `json:"attachments"`
		Fields      []EntityFieldData `json:"fields"`

		// Container-specific fields (for entities whose entity_type.is_location = true)
		Children   []EntitySummary `json:"children,omitempty"`
		TotalPrice float64         `json:"totalPrice,omitempty"`
	}

	// EntityOutCount is used for container listing with child count.
	EntityOutCount struct {
		EntitySummary
		ItemCount float64 `json:"itemCount"`
	}
)

var mapEntitiesSummaryErr = mapTEachErrFunc(mapEntitySummary)

func mapEntitySummary(e *ent.Entity) EntitySummary {
	var parent *EntitySummary
	if e.Edges.Parent != nil {
		p := mapEntitySummary(e.Edges.Parent)
		parent = &p
	}

	var et *EntityTypeSummary
	if e.Edges.EntityType != nil {
		s := mapEntityTypeSummary(e.Edges.EntityType)
		et = &s
	}

	tags := lo.Ternary(e.Edges.Tag != nil, mapEach(e.Edges.Tag, mapTagSummary), []TagSummary{})

	var imageID *uuid.UUID
	var thumbnailID *uuid.UUID
	if e.Edges.Attachments != nil {
		if a, ok := lo.Find(e.Edges.Attachments, func(a *ent.Attachment) bool {
			return a.Primary && a.Type == attachment.TypePhoto
		}); ok {
			imageID = &a.ID
			if a.Edges.Thumbnail != nil && a.Edges.Thumbnail.ID != uuid.Nil {
				thumbnailID = &a.Edges.Thumbnail.ID
			}
		}
	}

	return EntitySummary{
		ID:            e.ID,
		AssetID:       AssetID(e.AssetID),
		Name:          e.Name,
		Description:   e.Description,
		ImportRef:     e.ImportRef,
		Quantity:      e.Quantity,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		Archived:      e.Archived,
		PurchasePrice: e.PurchasePrice,

		// Edges
		Parent:     parent,
		EntityType: et,
		Tags:       tags,

		// Warranty
		Insured:     e.Insured,
		ImageID:     imageID,
		ThumbnailId: thumbnailID,
	}
}

var (
	mapEntityOutErr   = mapTErrFunc(mapEntityOut)
	mapEntitiesOutErr = mapTEachErrFunc(mapEntityOut)
)

func mapEntityFields(fields []*ent.EntityField) []EntityFieldData {
	return lo.Map(fields, func(f *ent.EntityField, _ int) EntityFieldData {
		return EntityFieldData{
			ID:           f.ID,
			Type:         f.Type.String(),
			Name:         f.Name,
			TextValue:    f.TextValue,
			NumberValue:  f.NumberValue,
			BooleanValue: f.BooleanValue,
		}
	})
}

func mapEntityOut(e *ent.Entity) EntityOut {
	var attachments []ItemAttachment
	if e.Edges.Attachments != nil {
		attachments = mapEach(e.Edges.Attachments, ToItemAttachment)
	}

	var fields []EntityFieldData
	if e.Edges.Fields != nil {
		fields = mapEntityFields(e.Edges.Fields)
	}

	var parent *EntitySummary
	if e.Edges.Parent != nil {
		p := mapEntitySummary(e.Edges.Parent)
		parent = &p
	}

	var children []EntitySummary
	if e.Edges.Children != nil {
		// Only include location-type children (sub-containers), not items
		children = lo.FilterMap(e.Edges.Children, func(c *ent.Entity, _ int) (EntitySummary, bool) {
			if c.Edges.EntityType != nil && c.Edges.EntityType.IsLocation {
				return mapEntitySummary(c), true
			}
			return EntitySummary{}, false
		})
	}

	return EntityOut{
		Parent:                   parent,
		AssetID:                  AssetID(e.AssetID),
		EntitySummary:            mapEntitySummary(e),
		LifetimeWarranty:         e.LifetimeWarranty,
		WarrantyExpires:          types.DateFromTime(e.WarrantyExpires),
		WarrantyDetails:          e.WarrantyDetails,
		SyncChildEntityLocations: e.SyncChildEntityLocations,

		// Identification
		SerialNumber: e.SerialNumber,
		ModelNumber:  e.ModelNumber,
		Manufacturer: e.Manufacturer,

		// Purchase
		PurchaseTime: types.DateFromTime(e.PurchaseTime),
		PurchaseFrom: e.PurchaseFrom,

		// Sold
		SoldTime:  types.DateFromTime(e.SoldTime),
		SoldTo:    e.SoldTo,
		SoldPrice: e.SoldPrice,
		SoldNotes: e.SoldNotes,

		// Extras
		Notes:       e.Notes,
		Attachments: attachments,
		Fields:      fields,
		Children:    children,
	}
}

// resolveDefaultEntityType finds or creates the default entity type for a group.
func (r *EntityRepository) resolveDefaultEntityType(ctx context.Context, gid uuid.UUID, isLocation bool) (uuid.UUID, error) {
	et, err := r.db.EntityType.Query().
		Where(
			entitytype.HasGroupWith(group.ID(gid)),
			entitytype.IsLocation(isLocation),
		).
		Order(entitytype.ByCreatedAt()).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			name := "Item"
			if isLocation {
				name = "Location"
			}
			created, err := r.db.EntityType.Create().
				SetName(name).
				SetDescription("").
				SetIsLocation(isLocation).
				SetGroupID(gid).
				Save(ctx)
			if err != nil {
				return uuid.Nil, err
			}
			return created.ID, nil
		}
		return uuid.Nil, err
	}
	return et.ID, nil
}

func (r *EntityRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventEntityMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (r *EntityRepository) getOneTx(ctx context.Context, tx *ent.Tx, where ...predicate.Entity) (EntityOut, error) {
	var q *ent.EntityQuery
	if tx != nil {
		q = tx.Entity.Query().Where(where...)
	} else {
		q = r.db.Entity.Query().Where(where...)
	}

	return mapEntityOutErr(q.
		WithFields().
		WithTag().
		WithParent().
		WithEntityType().
		WithGroup().
		WithChildren(func(eq *ent.EntityQuery) {
			eq.WithEntityType()
		}).
		WithAttachments().
		Only(ctx),
	)
}

func (r *EntityRepository) getOne(ctx context.Context, where ...predicate.Entity) (EntityOut, error) {
	return r.getOneTx(ctx, nil, where...)
}

// GetOne returns a single entity by ID. If the entity does not exist, an error is returned.
func (r *EntityRepository) GetOne(ctx context.Context, id uuid.UUID) (EntityOut, error) {
	return r.getOne(ctx, entity.ID(id))
}

func (r *EntityRepository) CheckRef(ctx context.Context, gid uuid.UUID, ref string) (bool, error) {
	q := r.db.Entity.Query().Where(entity.HasGroupWith(group.ID(gid)))
	return q.Where(entity.ImportRef(ref)).Exist(ctx)
}

func (r *EntityRepository) GetByRef(ctx context.Context, gid uuid.UUID, ref string) (EntityOut, error) {
	return r.getOne(ctx, entity.ImportRef(ref), entity.HasGroupWith(group.ID(gid)))
}

// GetOneByGroup returns a single entity by ID, verified to belong to a specific group.
func (r *EntityRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (EntityOut, error) {
	return r.getOne(ctx, entity.ID(id), entity.HasGroupWith(group.ID(gid)))
}

// QueryByGroup returns a list of entities that belong to a specific group based on the provided query.
func (r *EntityRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q EntityQuery) (PaginationResult[EntitySummary], error) {
	qb := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
	)

	// Filter by entity type (location vs item) when specified.
	// Default (nil) = items only (excludes locations for backward compat)
	switch {
	case q.IsLocation != nil && *q.IsLocation:
		qb = qb.Where(entity.HasEntityTypeWith(entitytype.IsLocation(true)))
	default:
		// nil or false: exclude locations
		qb = qb.Where(
			entity.Or(
				entity.Not(entity.HasEntityType()),
				entity.HasEntityTypeWith(entitytype.IsLocation(false)),
			),
		)
	}

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
		qb.Where(
			entity.Or(
				entity.NameContainsFold(q.Search),
				entity.DescriptionContainsFold(q.Search),
				entity.SerialNumberContainsFold(q.Search),
				entity.ModelNumberContainsFold(q.Search),
				entity.ManufacturerContainsFold(q.Search),
				entity.NotesContainsFold(q.Search),
			),
		)
	}

	if !q.AssetID.Nil() {
		qb = qb.Where(entity.AssetID(q.AssetID.Int()))
	}

	var andPredicates []predicate.Entity
	{
		if len(q.TagIDs) > 0 {
			tagRepo := &TagRepository{r.db, r.bus}
			descendants, err := tagRepo.GetDescendantTagIDs(ctx, q.TagIDs)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get descendant tags, using only direct tags")
				descendants = q.TagIDs
			} else if len(descendants) == 0 {
				descendants = q.TagIDs
			}

			var tagPredicates []predicate.Entity
			if !q.NegateTags {
				tagPredicates = lo.Map(descendants, func(l uuid.UUID, _ int) predicate.Entity {
					return entity.HasTagWith(tag.ID(l))
				})
				andPredicates = append(andPredicates, entity.Or(tagPredicates...))
			} else {
				tagPredicates = lo.Map(descendants, func(l uuid.UUID, _ int) predicate.Entity {
					return entity.Not(entity.HasTagWith(tag.ID(l)))
				})
				andPredicates = append(andPredicates, entity.And(tagPredicates...))
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

		if len(q.ParentIDs) > 0 {
			parentPredicates := lo.Map(q.ParentIDs, func(l uuid.UUID, _ int) predicate.Entity {
				return entity.HasParentWith(entity.ID(l))
			})
			andPredicates = append(andPredicates, entity.Or(parentPredicates...))
		}

		if len(q.Fields) > 0 {
			fieldPredicates := lo.Map(q.Fields, func(f FieldQuery, _ int) predicate.Entity {
				return entity.HasFieldsWith(
					entityfield.And(
						entityfield.Name(f.Name),
						entityfield.TextValue(f.Value),
					),
				)
			})
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
		WithParent().
		WithEntityType().
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

	entities, err := mapEntitiesSummaryErr(qb.All(ctx))
	if err != nil {
		return PaginationResult[EntitySummary]{}, err
	}

	return PaginationResult[EntitySummary]{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    count,
		Items:    entities,
	}, nil
}

// QueryByAssetID returns entities by asset ID.
func (r *EntityRepository) QueryByAssetID(ctx context.Context, gid uuid.UUID, assetID AssetID, page int, pageSize int) (PaginationResult[EntitySummary], error) {
	qb := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(int(assetID)),
	)

	if page != -1 || pageSize != -1 {
		qb.Offset(calculateOffset(page, pageSize)).
			Limit(pageSize)
	} else {
		page = -1
		pageSize = -1
	}

	entities, err := mapEntitiesSummaryErr(
		qb.Order(ent.Asc(entity.FieldName)).
			WithTag().
			WithParent().
			WithEntityType().
			All(ctx),
	)
	if err != nil {
		return PaginationResult[EntitySummary]{}, err
	}

	return PaginationResult[EntitySummary]{
		Page:     page,
		PageSize: pageSize,
		Total:    len(entities),
		Items:    entities,
	}, nil
}

// GetAll returns all the entities in the database with the Tags, Parent, and EntityType eager loaded.
func (r *EntityRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]EntityOut, error) {
	return mapEntitiesOutErr(r.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid))).
		WithTag().
		WithParent().
		WithEntityType().
		WithFields().
		All(ctx))
}

func (r *EntityRepository) GetAllZeroAssetID(ctx context.Context, gid uuid.UUID) ([]EntitySummary, error) {
	q := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(0),
	).Order(
		ent.Asc(entity.FieldCreatedAt),
	)

	return mapEntitiesSummaryErr(q.All(ctx))
}

func (r *EntityRepository) GetHighestAssetIDTx(ctx context.Context, tx *ent.Tx, gid uuid.UUID) (AssetID, error) {
	var q *ent.EntityQuery
	if tx != nil {
		q = tx.Entity.Query().Where(
			entity.HasGroupWith(group.ID(gid)),
		).Order(
			ent.Desc(entity.FieldAssetID),
		).Limit(1)
	} else {
		q = r.db.Entity.Query().Where(
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

func (r *EntityRepository) GetHighestAssetID(ctx context.Context, gid uuid.UUID) (AssetID, error) {
	return r.GetHighestAssetIDTx(ctx, nil, gid)
}

func (r *EntityRepository) SetAssetID(ctx context.Context, gid uuid.UUID, id uuid.UUID, assetID AssetID) error {
	q := r.db.Entity.Update().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.ID(id),
	)

	_, err := q.SetAssetID(int(assetID)).Save(ctx)
	return err
}

func validateQuantity(op string, quantity float64) error {
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return fmt.Errorf("%s: invalid quantity: must be a finite number", op)
	}

	return nil
}

func (r *EntityRepository) Create(ctx context.Context, gid uuid.UUID, data EntityCreate) (EntityOut, error) {
	if err := validateQuantity("create entity", data.Quantity); err != nil {
		return EntityOut{}, err
	}

	q := r.db.Entity.Create().
		SetImportRef(data.ImportRef).
		SetName(data.Name).
		SetQuantity(data.Quantity).
		SetDescription(data.Description).
		SetGroupID(gid).
		SetAssetID(int(data.AssetID))

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	} else {
		// Auto-resolve default "Item" entity type for the group
		etID, err := r.resolveDefaultEntityType(ctx, gid, false)
		if err != nil {
			return EntityOut{}, err
		}
		q.SetEntityTypeID(etID)
	}

	if len(data.TagIDs) > 0 {
		q.AddTagIDs(data.TagIDs...)
	}

	result, err := q.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, result.ID)
}

// EntityCreateFromTemplate contains all data needed to create an entity from a template.
type EntityCreateFromTemplate struct {
	Name             string
	Description      string
	Quantity         float64
	ParentID         uuid.UUID
	EntityTypeID     uuid.UUID
	TagIDs           []uuid.UUID
	Insured          bool
	Manufacturer     string
	ModelNumber      string
	LifetimeWarranty bool
	WarrantyDetails  string
	Fields           []EntityFieldData
}

// CreateFromTemplate creates an entity with all template data in a single transaction.
func (r *EntityRepository) CreateFromTemplate(ctx context.Context, gid uuid.UUID, data EntityCreateFromTemplate) (EntityOut, error) {
	if err := validateQuantity("create entity from template", data.Quantity); err != nil {
		return EntityOut{}, err
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return EntityOut{}, err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during template entity creation")
			}
		}
	}()

	// Get next asset ID within transaction
	nextAssetID, err := r.GetHighestAssetIDTx(ctx, tx, gid)
	if err != nil {
		return EntityOut{}, err
	}
	nextAssetID++

	// Create entity with all template data
	newEntityID := uuid.New()
	entityBuilder := tx.Entity.Create().
		SetID(newEntityID).
		SetName(data.Name).
		SetDescription(data.Description).
		SetQuantity(data.Quantity).
		SetGroupID(gid).
		SetAssetID(int(nextAssetID)).
		SetInsured(data.Insured).
		SetManufacturer(data.Manufacturer).
		SetModelNumber(data.ModelNumber).
		SetLifetimeWarranty(data.LifetimeWarranty).
		SetWarrantyDetails(data.WarrantyDetails)

	if data.ParentID != uuid.Nil {
		entityBuilder.SetParentID(data.ParentID)
	}

	if data.EntityTypeID != uuid.Nil {
		entityBuilder.SetEntityTypeID(data.EntityTypeID)
	}

	if len(data.TagIDs) > 0 {
		entityBuilder.AddTagIDs(data.TagIDs...)
	}

	_, err = entityBuilder.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	// Create custom fields
	for _, field := range data.Fields {
		_, err = tx.EntityField.Create().
			SetEntityID(newEntityID).
			SetType(entityfield.Type(field.Type)).
			SetName(field.Name).
			SetTextValue(field.TextValue).
			Save(ctx)
		if err != nil {
			return EntityOut{}, fmt.Errorf("failed to create field %s: %w", field.Name, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return EntityOut{}, err
	}
	committed = true

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, newEntityID)
}

func (r *EntityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get the entity with its group and attachments before deletion
	e, err := r.db.Entity.Query().
		Where(entity.ID(id)).
		WithGroup().
		WithAttachments().
		Only(ctx)
	if err != nil {
		return err
	}

	// Get the group ID for attachment deletion
	var gid uuid.UUID
	if e.Edges.Group != nil {
		gid = e.Edges.Group.ID
	}

	// Delete all attachments (and their files) before deleting the entity
	for _, att := range e.Edges.Attachments {
		err := r.attachments.Delete(ctx, gid, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during entity deletion")
		}
	}

	err = r.db.Entity.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(id)
	return nil
}

func (r *EntityRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	// Get the entity with its attachments before deletion
	e, err := r.db.Entity.Query().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).
		WithAttachments().
		Only(ctx)
	if err != nil {
		return err
	}

	// Delete all attachments (and their files) before deleting the entity
	for _, att := range e.Edges.Attachments {
		err := r.attachments.Delete(ctx, gid, att.ID)
		if err != nil {
			log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during entity deletion")
		}
	}

	_, err = r.db.Entity.
		Delete().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return err
}

func (r *EntityRepository) WipeInventory(ctx context.Context, gid uuid.UUID, wipeTags bool, wipeContainers bool, wipeMaintenance bool) (int, error) {
	deleted := 0

	// Wipe maintenance records if requested
	// IMPORTANT: Must delete maintenance records BEFORE entities since they are linked to entities
	if wipeMaintenance {
		maintenanceCount, err := r.db.MaintenanceEntry.Delete().
			Where(maintenanceentry.HasEntityWith(entity.HasGroupWith(group.ID(gid)))).
			Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete maintenance entries during wipe inventory")
		} else {
			log.Info().Int("count", maintenanceCount).Msg("deleted maintenance entries during wipe inventory")
			deleted += maintenanceCount
		}
	}

	// Get all entities for the group
	entities, err := r.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid))).
		WithAttachments().
		All(ctx)
	if err != nil {
		return 0, err
	}

	// Delete each entity with its attachments
	for _, e := range entities {
		for _, att := range e.Edges.Attachments {
			err := r.attachments.Delete(ctx, gid, att.ID)
			if err != nil {
				log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during wipe inventory")
			}
		}

		_, err = r.db.Entity.
			Delete().
			Where(
				entity.ID(e.ID),
				entity.HasGroupWith(group.ID(gid)),
			).Exec(ctx)
		if err != nil {
			log.Err(err).Str("entity_id", e.ID.String()).Msg("failed to delete entity during wipe inventory")
			continue
		}

		deleted++
	}

	// Wipe tags if requested
	if wipeTags {
		tagCount, err := r.db.Tag.Delete().Where(tag.HasGroupWith(group.ID(gid))).Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete tags during wipe inventory")
		} else {
			log.Info().Int("count", tagCount).Msg("deleted tags during wipe inventory")
			deleted += tagCount
		}
	}

	// Wipe containers (location-type entities) if requested
	if wipeContainers {
		containerCount, err := r.db.Entity.Delete().
			Where(
				entity.HasGroupWith(group.ID(gid)),
				entity.HasEntityTypeWith(entitytype.IsLocation(true)),
			).Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to delete containers during wipe inventory")
		} else {
			log.Info().Int("count", containerCount).Msg("deleted containers during wipe inventory")
			deleted += containerCount
		}
	}

	r.publishMutationEvent(gid)
	return deleted, nil
}

func (r *EntityRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data EntityUpdate) (EntityOut, error) {
	if err := validateQuantity("update entity", data.Quantity); err != nil {
		return EntityOut{}, err
	}

	q := r.db.Entity.Update().Where(entity.ID(data.ID), entity.HasGroupWith(group.ID(gid))).
		SetName(data.Name).
		SetDescription(data.Description).
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
		SetSyncChildEntityLocations(data.SyncChildEntityLocations)

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	}

	currentTags, err := r.db.Entity.Query().Where(entity.ID(data.ID)).QueryTag().All(ctx)
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

	if data.SyncChildEntityLocations {
		children, err := r.db.Entity.Query().Where(entity.ID(data.ID)).QueryChildren().All(ctx)
		if err != nil {
			return EntityOut{}, err
		}

		for _, child := range children {
			if data.ParentID != uuid.Nil {
				childParent, err := child.QueryParent().First(ctx)
				if err != nil || childParent.ID != data.ParentID {
					err = child.Update().SetParentID(data.ParentID).Exec(ctx)
					if err != nil {
						return EntityOut{}, err
					}
				}
			}
		}
	}

	err = q.Exec(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	fields, err := r.db.EntityField.Query().Where(entityfield.HasEntityWith(entity.ID(data.ID))).All(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	fieldIds := newIDSet(fields)

	// Update Existing Fields
	for _, f := range data.Fields {
		if f.ID == uuid.Nil {
			// Create New Field
			_, err = r.db.EntityField.Create().
				SetEntityID(data.ID).
				SetType(entityfield.Type(f.Type)).
				SetName(f.Name).
				SetTextValue(f.TextValue).
				SetNumberValue(f.NumberValue).
				SetBooleanValue(f.BooleanValue).
				Save(ctx)
			if err != nil {
				return EntityOut{}, err
			}
		}

		opt := r.db.EntityField.Update().
			Where(
				entityfield.ID(f.ID),
				entityfield.HasEntityWith(entity.ID(data.ID)),
			).
			SetType(entityfield.Type(f.Type)).
			SetName(f.Name).
			SetTextValue(f.TextValue).
			SetNumberValue(f.NumberValue).
			SetBooleanValue(f.BooleanValue)

		_, err = opt.Save(ctx)
		if err != nil {
			return EntityOut{}, err
		}

		fieldIds.Remove(f.ID)
		continue
	}

	// Delete Fields that are no longer present
	if fieldIds.Len() > 0 {
		_, err = r.db.EntityField.Delete().
			Where(
				entityfield.IDIn(fieldIds.Slice()...),
				entityfield.HasEntityWith(entity.ID(data.ID)),
			).Exec(ctx)
		if err != nil {
			return EntityOut{}, err
		}
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, data.ID)
}

func (r *EntityRepository) GetAllZeroImportRef(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID

	err := r.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
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

func (r *EntityRepository) Patch(ctx context.Context, gid, id uuid.UUID, data EntityPatch) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during entity patch")
			}
		}
	}()

	q := tx.Entity.Update().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		)

	if data.ImportRef != nil {
		q.SetImportRef(*data.ImportRef)
	}

	if data.Quantity != nil {
		if err := validateQuantity("patch entity", *data.Quantity); err != nil {
			return err
		}

		q.SetQuantity(*data.Quantity)
	}

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	}

	err = q.Exec(ctx)
	if err != nil {
		return err
	}

	if data.TagIDs != nil {
		currentTags, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryTag().All(ctx)
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
			if err := tx.Entity.Update().
				Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
				AddTagIDs(addTags...).
				Exec(ctx); err != nil {
				return err
			}
		}
		if set.Len() > 0 {
			if err := tx.Entity.Update().
				Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
				RemoveTagIDs(set.Slice()...).
				Exec(ctx); err != nil {
				return err
			}
		}
	}

	if data.ParentID != uuid.Nil {
		entityEnt, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).Only(ctx)
		if err != nil {
			return err
		}
		if entityEnt.SyncChildEntityLocations {
			children, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryChildren().All(ctx)
			if err != nil {
				return err
			}
			for _, child := range children {
				childParent, err := child.QueryParent().First(ctx)
				if err != nil || childParent.ID != data.ParentID {
					err = child.Update().SetParentID(data.ParentID).Exec(ctx)
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

	r.publishMutationEvent(gid)
	return nil
}

func (r *EntityRepository) GetAllCustomFieldValues(ctx context.Context, gid uuid.UUID, name string) ([]string, error) {
	type st struct {
		Value string `json:"text_value"`
	}

	var values []st

	err := r.db.Entity.Query().
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

	valueStrings := lo.Map(values, func(f st, _ int) string {
		return f.Value
	})

	return valueStrings, nil
}

func (r *EntityRepository) GetAllCustomFieldNames(ctx context.Context, gid uuid.UUID) ([]string, error) {
	type st struct {
		Name string `json:"name"`
	}

	var fields []st

	err := r.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
		).
		QueryFields().
		Unique(true).
		Select(entityfield.FieldName).
		Scan(ctx, &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom fields: %w", err)
	}

	fieldNames := lo.Map(fields, func(f st, _ int) string {
		return f.Name
	})

	return fieldNames, nil
}

// ZeroOutTimeFields sets all date fields to the beginning of the day.
func (r *EntityRepository) ZeroOutTimeFields(ctx context.Context, gid uuid.UUID) (int, error) {
	q := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.Or(
			entity.PurchaseTimeNotNil(),
			entity.PurchaseFromLT("0002-01-01"),
			entity.SoldTimeNotNil(),
			entity.SoldToLT("0002-01-01"),
			entity.WarrantyExpiresNotNil(),
			entity.WarrantyDetailsLT("0002-01-01"),
		),
	)

	entities, err := q.All(ctx)
	if err != nil {
		return -1, fmt.Errorf("ZeroOutTimeFields() -> failed to get entities: %w", err)
	}

	toDateOnly := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	updated := 0

	for _, e := range entities {
		updateQ := r.db.Entity.Update().Where(entity.ID(e.ID))

		if !e.PurchaseTime.IsZero() {
			switch {
			case e.PurchaseTime.Year() < 100:
				updateQ.ClearPurchaseTime()
			default:
				updateQ.SetPurchaseTime(toDateOnly(e.PurchaseTime))
			}
		} else {
			updateQ.ClearPurchaseTime()
		}

		if !e.SoldTime.IsZero() {
			switch {
			case e.SoldTime.Year() < 100:
				updateQ.ClearSoldTime()
			default:
				updateQ.SetSoldTime(toDateOnly(e.SoldTime))
			}
		} else {
			updateQ.ClearSoldTime()
		}

		if !e.WarrantyExpires.IsZero() {
			switch {
			case e.WarrantyExpires.Year() < 100:
				updateQ.ClearWarrantyExpires()
			default:
				updateQ.SetWarrantyExpires(toDateOnly(e.WarrantyExpires))
			}
		} else {
			updateQ.ClearWarrantyExpires()
		}

		_, err = updateQ.Save(ctx)
		if err != nil {
			return updated, fmt.Errorf("ZeroOutTimeFields() -> failed to update entity: %w", err)
		}

		updated++
	}

	return updated, nil
}

func (r *EntityRepository) SetPrimaryPhotos(ctx context.Context, gid uuid.UUID) (int, error) {
	// All entities where there is no primary photo
	entityIDs, err := r.db.Entity.Query().
		Where(
			entity.HasGroupWith(group.ID(gid)),
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
	for _, id := range entityIDs {
		a, err := r.db.Attachment.Query().
			Where(
				attachment.HasEntityWith(entity.ID(id)),
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Primary(false),
			).
			First(ctx)
		if err != nil {
			return updated, err
		}

		_, err = r.db.Attachment.UpdateOne(a).
			SetPrimary(true).
			Save(ctx)
		if err != nil {
			return updated, err
		}

		updated++
	}

	return updated, nil
}

// Duplicate creates a copy of an entity with configurable options for what data to copy.
func (r *EntityRepository) Duplicate(ctx context.Context, gid, id uuid.UUID, options DuplicateOptions) (EntityOut, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return EntityOut{}, err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Warn().Err(err).Msg("failed to rollback transaction during entity duplication")
			}
		}
	}()

	// Get the original entity with all its data
	originalEntity, err := r.getOneTx(ctx, tx, entity.ID(id), entity.HasGroupWith(group.ID(gid)))
	if err != nil {
		return EntityOut{}, err
	}

	nextAssetID, err := r.GetHighestAssetIDTx(ctx, tx, gid)
	if err != nil {
		return EntityOut{}, err
	}
	nextAssetID++

	if options.CopyPrefix == "" {
		options.CopyPrefix = "Copy of "
	}

	newEntityID := uuid.New()
	entityBuilder := tx.Entity.Create().
		SetID(newEntityID).
		SetName(options.CopyPrefix + originalEntity.Name).
		SetDescription(originalEntity.Description).
		SetQuantity(originalEntity.Quantity).
		SetGroupID(gid).
		SetAssetID(int(nextAssetID)).
		SetSerialNumber(originalEntity.SerialNumber).
		SetModelNumber(originalEntity.ModelNumber).
		SetManufacturer(originalEntity.Manufacturer).
		SetLifetimeWarranty(originalEntity.LifetimeWarranty).
		SetWarrantyExpires(originalEntity.WarrantyExpires.Time()).
		SetWarrantyDetails(originalEntity.WarrantyDetails).
		SetPurchaseTime(originalEntity.PurchaseTime.Time()).
		SetPurchaseFrom(originalEntity.PurchaseFrom).
		SetPurchasePrice(originalEntity.PurchasePrice).
		SetSoldTime(originalEntity.SoldTime.Time()).
		SetSoldTo(originalEntity.SoldTo).
		SetSoldPrice(originalEntity.SoldPrice).
		SetSoldNotes(originalEntity.SoldNotes).
		SetNotes(originalEntity.Notes).
		SetInsured(originalEntity.Insured).
		SetArchived(originalEntity.Archived).
		SetSyncChildEntityLocations(originalEntity.SyncChildEntityLocations)

	if originalEntity.Parent != nil {
		entityBuilder.SetParentID(originalEntity.Parent.ID)
	}

	if originalEntity.EntityType != nil {
		entityBuilder.SetEntityTypeID(originalEntity.EntityType.ID)
	}

	// Add tags
	if len(originalEntity.Tags) > 0 {
		tagIDs := lo.Map(originalEntity.Tags, func(tag TagSummary, _ int) uuid.UUID {
			return tag.ID
		})
		entityBuilder.AddTagIDs(tagIDs...)
	}

	_, err = entityBuilder.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	// Copy custom fields if requested
	if options.CopyCustomFields {
		for _, field := range originalEntity.Fields {
			_, err = tx.EntityField.Create().
				SetEntityID(newEntityID).
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
		for _, att := range originalEntity.Attachments {
			originalAttachment, err := tx.Attachment.Query().
				Where(attachment.ID(att.ID)).
				Only(ctx)
			if err != nil {
				log.Warn().Err(err).Str("attachment_id", att.ID.String()).Msg("failed to find attachment during duplication")
				continue
			}

			_, err = tx.Attachment.Create().
				SetEntityID(newEntityID).
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
					SetEntityID(newEntityID).
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

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, newEntityID)
}

// ============================================================================
// Container / Location methods (absorbed from LocationRepository)
// ============================================================================

type ContainerQuery struct {
	FilterChildren bool `json:"filterChildren" schema:"filterChildren"`
}

// GetAllContainers returns all container entities (entity_type.is_location = true) with child entity counts.
func (r *EntityRepository) GetAllContainers(ctx context.Context, gid uuid.UUID, filter ContainerQuery) ([]EntityOutCount, error) {
	query := `--sql
		SELECT
			e.id,
			e.name,
			e.description,
			e.created_at,
			e.updated_at,
			(
				SELECT
					SUM(child.quantity)
				FROM
					entities child
				JOIN
					entity_types ct ON ct.id = child.entity_type_entities
				WHERE
					child.entity_children = e.id
					AND child.archived = false
					AND ct.is_location = false
			) as item_count
		FROM
			entities e
		JOIN
			entity_types et ON et.id = e.entity_type_entities
		WHERE
			e.group_entities = $1
			AND et.is_location = true
			{{ FILTER_CHILDREN }}
		ORDER BY
			e.name ASC
`

	if filter.FilterChildren {
		query = strings.Replace(query, "{{ FILTER_CHILDREN }}", "AND e.entity_children IS NULL", 1)
	} else {
		query = strings.Replace(query, "{{ FILTER_CHILDREN }}", "", 1)
	}

	rows, err := r.db.Sql().QueryContext(ctx, query, gid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	list := []EntityOutCount{}
	for rows.Next() {
		var ct EntityOutCount

		var maybeCount *float64

		err := rows.Scan(&ct.ID, &ct.Name, &ct.Description, &ct.CreatedAt, &ct.UpdatedAt, &maybeCount)
		if err != nil {
			return nil, err
		}

		if maybeCount != nil {
			ct.ItemCount = *maybeCount
		}

		list = append(list, ct)
	}

	return list, err
}

// GetContainerByGroup returns a single container entity by ID, verified to belong to a specific group.
func (r *EntityRepository) GetContainerByGroup(ctx context.Context, gid, id uuid.UUID) (EntityOut, error) {
	return mapEntityOutErr(r.db.Entity.Query().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).
		WithGroup().
		WithParent().
		WithChildren(func(eq *ent.EntityQuery) {
			eq.Order(ent.Asc(entity.FieldName))
		}).
		WithEntityType().
		Only(ctx))
}

// CreateContainer creates a container entity (with a location-type entity_type).
func (r *EntityRepository) CreateContainer(ctx context.Context, gid uuid.UUID, data EntityCreate) (EntityOut, error) {
	// Validate: if a parent is specified, it must also be a location-type entity
	if data.ParentID != uuid.Nil {
		parentEntity, err := r.db.Entity.Query().
			Where(entity.ID(data.ParentID)).
			WithEntityType().
			Only(ctx)
		if err != nil {
			return EntityOut{}, fmt.Errorf("parent entity not found: %w", err)
		}
		if parentEntity.Edges.EntityType == nil || !parentEntity.Edges.EntityType.IsLocation {
			return EntityOut{}, fmt.Errorf("locations can only have other locations as parents, not items")
		}
	}

	q := r.db.Entity.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetGroupID(gid)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	} else {
		// Auto-resolve default "Location" entity type for the group
		etID, err := r.resolveDefaultEntityType(ctx, gid, true)
		if err != nil {
			return EntityOut{}, err
		}
		q.SetEntityTypeID(etID)
	}

	result, err := q.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	result.Edges.Group = &ent.Group{ID: gid}
	r.publishMutationEvent(gid)
	return mapEntityOut(result), nil
}

// UpdateContainer updates a container entity.
func (r *EntityRepository) UpdateContainer(ctx context.Context, gid, id uuid.UUID, data EntityUpdate) (EntityOut, error) {
	// Validate: if a parent is specified, it must also be a location-type entity
	if data.ParentID != uuid.Nil {
		parentEntity, err := r.db.Entity.Query().
			Where(entity.ID(data.ParentID)).
			WithEntityType().
			Only(ctx)
		if err != nil {
			return EntityOut{}, fmt.Errorf("parent entity not found: %w", err)
		}
		if parentEntity.Edges.EntityType == nil || !parentEntity.Edges.EntityType.IsLocation {
			return EntityOut{}, fmt.Errorf("locations can only have other locations as parents, not items")
		}
	}

	q := r.db.Entity.Update().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).
		SetName(data.Name).
		SetDescription(data.Description)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	_, err := q.Save(ctx)
	if err != nil {
		return EntityOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, id)
}

// DeleteContainerByGroup deletes a container entity by group.
func (r *EntityRepository) DeleteContainerByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := r.db.Entity.Delete().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).Exec(ctx)
	if err != nil {
		return err
	}
	r.publishMutationEvent(gid)
	return nil
}

// ============================================================================
// Tree and Path methods (absorbed from LocationRepository)
// ============================================================================

type TreeItem struct {
	ID       uuid.UUID   `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Children []*TreeItem `json:"children"`
}

type FlatTreeItem struct {
	ID       uuid.UUID
	Name     string
	Type     string
	ParentID uuid.UUID
	Level    int
}

type TreeQuery struct {
	WithItems bool `json:"withItems" schema:"withItems"`
}

type EntityPathType string

const (
	EntityPathTypeLocation EntityPathType = "location"
	EntityPathTypeItem     EntityPathType = "item"
)

type EntityPath struct {
	Type EntityPathType `json:"type"`
	ID   uuid.UUID      `json:"id"`
	Name string         `json:"name"`
}

func (r *EntityRepository) PathForEntity(ctx context.Context, gid, entityID uuid.UUID) ([]EntityPath, error) {
	query := `WITH RECURSIVE entity_path AS (
		SELECT id, name, entity_children
		FROM entities
		WHERE id = $1
		AND group_entities = $2

		UNION ALL

		SELECT e.id, e.name, e.entity_children
		FROM entities e
		JOIN entity_path ep ON e.id = ep.entity_children
	  )

	  SELECT id, name
	  FROM entity_path`

	rows, err := r.db.Sql().QueryContext(ctx, query, entityID, gid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var path []EntityPath

	for rows.Next() {
		var entry EntityPath
		entry.Type = EntityPathTypeLocation
		if err := rows.Scan(&entry.ID, &entry.Name); err != nil {
			return nil, err
		}
		path = append(path, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse the order so that the root is first
	mutable.Reverse(path)

	return path, nil
}

func (r *EntityRepository) Tree(ctx context.Context, gid uuid.UUID, tq TreeQuery) ([]TreeItem, error) {
	query := `
		WITH recursive entity_tree(id, NAME, parent_id, level, node_type) AS
		(
			SELECT  e.id,
					e.NAME,
					e.entity_children AS parent_id,
					0 AS level,
					CASE WHEN et.is_location THEN 'location' ELSE 'item' END AS node_type
			FROM    entities e
			JOIN    entity_types et ON et.id = e.entity_type_entities
			WHERE   e.entity_children IS NULL
			AND     e.group_entities = $1
			AND     et.is_location = true

			UNION ALL
			SELECT  c.id,
					c.NAME,
					c.entity_children AS parent_id,
					level + 1,
					CASE WHEN ct.is_location THEN 'location' ELSE 'item' END AS node_type
			FROM   entities c
			JOIN   entity_types ct ON ct.id = c.entity_type_entities
			JOIN   entity_tree p
			ON     c.entity_children = p.id
			WHERE  level < 10 -- prevent infinite loop & excessive recursion
			AND    ct.is_location = true
		){{ WITH_ITEMS }}

		SELECT   id,
				 NAME,
				 level,
				 parent_id,
				 node_type
		FROM    (
					SELECT  *
					FROM    entity_tree

					{{ WITH_ITEMS_FROM }}

				) tree
		ORDER BY node_type DESC, -- sort locations before items
				 level,
				 lower(NAME)`

	if tq.WithItems {
		itemQuery := `, item_tree(id, NAME, parent_id, level, node_type) AS
		(
			SELECT  e.id,
					e.NAME,
					e.entity_children as parent_id,
					0 AS level,
					'item' AS node_type
			FROM    entities e
			JOIN    entity_types et ON et.id = e.entity_type_entities
			WHERE   et.is_location = false
			AND     e.entity_children IN (SELECT id FROM entity_tree)

			UNION ALL

			SELECT  c.id,
					c.NAME,
					c.entity_children AS parent_id,
					level + 1,
					'item' AS node_type
			FROM    entities c
			JOIN    entity_types ct ON ct.id = c.entity_type_entities
			JOIN    item_tree p
			ON      c.entity_children = p.id
			WHERE   ct.is_location = false
			AND     level < 10 -- prevent infinite loop & excessive recursion
		)`

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

	var flatItems []FlatTreeItem
	for rows.Next() {
		var item FlatTreeItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Level, &item.ParentID, &item.Type); err != nil {
			return nil, err
		}
		flatItems = append(flatItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ConvertEntitiesToTree(flatItems), nil
}

func ConvertEntitiesToTree(items []FlatTreeItem) []TreeItem {
	itemMap := make(map[uuid.UUID]*TreeItem, len(items))

	var rootIds []uuid.UUID

	for _, item := range items {
		node := &TreeItem{
			ID:       item.ID,
			Name:     item.Name,
			Type:     item.Type,
			Children: []*TreeItem{},
		}

		itemMap[item.ID] = node
		if item.ParentID != uuid.Nil {
			parent, ok := itemMap[item.ParentID]
			if ok {
				parent.Children = append(parent.Children, node)
			}
		} else {
			rootIds = append(rootIds, item.ID)
		}
	}

	return lo.Map(rootIds, func(id uuid.UUID, _ int) TreeItem {
		return *itemMap[id]
	})
}
