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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func entityTracer() trace.Tracer {
	return otel.Tracer("data")
}

func recordSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

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
		IsLocation       *bool        `json:"isLocation"`     // nil=all, true=locations only, false=items only
		FilterChildren   bool         `json:"filterChildren"` // when true, only return root entities (no parent)
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
		ParentID     uuid.UUID `json:"parentId"     extensions:"x-nullable"`
		Name         string    `json:"name"         validate:"required,min=1,max=255"`
		Quantity     float64   `json:"quantity"`
		Description  string    `json:"description"  validate:"max=1000"`
		AssetID      AssetID   `json:"-"`
		EntityTypeID uuid.UUID `json:"entityTypeId"`

		// Edges
		TagIDs []uuid.UUID `json:"tagIds"`
	}

	EntityUpdate struct {
		ParentID                 uuid.UUID `json:"parentId"                 extensions:"x-nullable,x-omitempty"`
		ID                       uuid.UUID `json:"id"`
		AssetID                  AssetID   `json:"assetId"                  swaggertype:"string"`
		Name                     string    `json:"name"                     validate:"required,min=1,max=255"`
		Description              string    `json:"description"              validate:"max=1000"`
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
		PurchaseDate  types.Date `json:"purchaseDate"`
		PurchaseFrom  string     `json:"purchaseFrom"  validate:"max=255"`
		PurchasePrice float64    `json:"purchasePrice" extensions:"x-nullable,x-omitempty"`

		// Sold
		SoldDate  types.Date `json:"soldDate"`
		SoldTo    string     `json:"soldTo"    validate:"max=255"`
		SoldPrice float64    `json:"soldPrice" extensions:"x-nullable,x-omitempty"`
		SoldNotes string     `json:"soldNotes"`

		// Extras
		Notes  string            `json:"notes"`
		Fields []EntityFieldData `json:"fields"`
	}

	EntityPatch struct {
		ID           uuid.UUID   `json:"id"`
		Quantity     *float64    `json:"quantity,omitempty" extensions:"x-nullable,x-omitempty"`
		ImportRef    *string     `json:"-"                  extensions:"x-nullable,x-omitempty"`
		ParentID     uuid.UUID   `json:"parentId"           extensions:"x-nullable,x-omitempty"`
		EntityTypeID uuid.UUID   `json:"entityTypeId"       extensions:"x-nullable,x-omitempty"`
		TagIDs       []uuid.UUID `json:"tagIds"             extensions:"x-nullable,x-omitempty"`
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
		SoldDate types.Date `json:"soldDate"`

		// Container-specific (populated when querying locations)
		ItemCount float64 `json:"itemCount,omitempty"`
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
		PurchaseDate types.Date `json:"purchaseDate"`
		PurchaseFrom string     `json:"purchaseFrom"`

		// Sold
		SoldDate  types.Date `json:"soldDate"`
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

		// Sale
		SoldDate: types.DateFromTime(e.SoldDate),
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
		PurchaseDate: types.DateFromTime(e.PurchaseDate),
		PurchaseFrom: e.PurchaseFrom,

		// Sold
		SoldDate:  types.DateFromTime(e.SoldDate),
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
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.resolveDefaultEntityType",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Bool("entity_type.is_location", isLocation),
		))
	defer span.End()

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
			createCtx, createSpan := entityTracer().Start(ctx, "repo.EntityRepository.resolveDefaultEntityType.create",
				trace.WithAttributes(attribute.String("entity_type.name", name)))
			created, err := r.db.EntityType.Create().
				SetName(name).
				SetDescription("").
				SetIsLocation(isLocation).
				SetGroupID(gid).
				Save(createCtx)
			if err != nil {
				recordSpanError(createSpan, err)
				createSpan.End()
				recordSpanError(span, err)
				return uuid.Nil, err
			}
			createSpan.SetAttributes(attribute.String("entity_type.id", created.ID.String()))
			createSpan.End()
			span.SetAttributes(attribute.String("entity_type.id", created.ID.String()))
			return created.ID, nil
		}
		recordSpanError(span, err)
		return uuid.Nil, err
	}
	span.SetAttributes(attribute.String("entity_type.id", et.ID.String()))
	return et.ID, nil
}

func (r *EntityRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventEntityMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (r *EntityRepository) getOneTx(ctx context.Context, tx *ent.Tx, where ...predicate.Entity) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.getOneTx",
		trace.WithAttributes(
			attribute.Bool("tx", tx != nil),
			attribute.Int("predicate.count", len(where)),
		))
	defer span.End()

	var q *ent.EntityQuery
	if tx != nil {
		q = tx.Entity.Query().Where(where...)
	} else {
		q = r.db.Entity.Query().Where(where...)
	}

	out, err := mapEntityOutErr(q.
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
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(
		attribute.String("entity.id", out.ID.String()),
		attribute.Int("entity.fields.count", len(out.Fields)),
		attribute.Int("entity.tags.count", len(out.Tags)),
		attribute.Int("entity.attachments.count", len(out.Attachments)),
		attribute.Int("entity.children.count", len(out.Children)),
	)
	return out, nil
}

func (r *EntityRepository) getOne(ctx context.Context, where ...predicate.Entity) (EntityOut, error) {
	return r.getOneTx(ctx, nil, where...)
}

// GetOne returns a single entity by ID. If the entity does not exist, an error is returned.
func (r *EntityRepository) GetOne(ctx context.Context, id uuid.UUID) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetOne",
		trace.WithAttributes(attribute.String("entity.id", id.String())))
	defer span.End()

	out, err := r.getOne(ctx, entity.ID(id))
	recordSpanError(span, err)
	return out, err
}

func (r *EntityRepository) CheckRef(ctx context.Context, gid uuid.UUID, ref string) (bool, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.CheckRef",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.import_ref", ref),
		))
	defer span.End()

	q := r.db.Entity.Query().Where(entity.HasGroupWith(group.ID(gid)))
	exists, err := q.Where(entity.ImportRef(ref)).Exist(ctx)
	if err != nil {
		recordSpanError(span, err)
		return exists, err
	}
	span.SetAttributes(attribute.Bool("entity.exists", exists))
	return exists, nil
}

func (r *EntityRepository) GetByRef(ctx context.Context, gid uuid.UUID, ref string) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetByRef",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.import_ref", ref),
		))
	defer span.End()

	out, err := r.getOne(ctx, entity.ImportRef(ref), entity.HasGroupWith(group.ID(gid)))
	recordSpanError(span, err)
	return out, err
}

// GetOneByGroup returns a single entity by ID, verified to belong to a specific group.
func (r *EntityRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetOneByGroup",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	out, err := r.getOne(ctx, entity.ID(id), entity.HasGroupWith(group.ID(gid)))
	recordSpanError(span, err)
	return out, err
}

func entityQuerySpanAttrs(gid uuid.UUID, q EntityQuery) []attribute.KeyValue {
	isLocSet := q.IsLocation != nil
	isLocValue := false
	if isLocSet {
		isLocValue = *q.IsLocation
	}
	return []attribute.KeyValue{
		attribute.String("group.id", gid.String()),
		attribute.Int("query.page", q.Page),
		attribute.Int("query.page_size", q.PageSize),
		attribute.String("query.search", q.Search),
		attribute.Int("query.tag_ids.count", len(q.TagIDs)),
		attribute.Bool("query.negate_tags", q.NegateTags),
		attribute.Int("query.parent_ids.count", len(q.ParentIDs)),
		attribute.Int("query.parent_item_ids.count", len(q.ParentItemIDs)),
		attribute.Int("query.fields.count", len(q.Fields)),
		attribute.Bool("query.only_with_photo", q.OnlyWithPhoto),
		attribute.Bool("query.only_without_photo", q.OnlyWithoutPhoto),
		attribute.Bool("query.include_archived", q.IncludeArchived),
		attribute.Bool("query.filter_children", q.FilterChildren),
		attribute.String("query.order_by", q.OrderBy),
		attribute.Bool("query.is_location.set", isLocSet),
		attribute.Bool("query.is_location.value", isLocValue),
		attribute.Bool("query.asset_id.set", !q.AssetID.Nil()),
	}
}

// QueryByGroup returns a list of entities that belong to a specific group based on the provided query.
func (r *EntityRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q EntityQuery) (PaginationResult[EntitySummary], error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.QueryByGroup",
		trace.WithAttributes(entityQuerySpanAttrs(gid, q)...))
	defer span.End()

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

	if q.FilterChildren {
		qb = qb.Where(entity.Not(entity.HasParent()))
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
		qb = qb.Where(entity.AssetID(int64(q.AssetID)))
	}

	var andPredicates []predicate.Entity
	{
		if len(q.TagIDs) > 0 {
			tagRepo := &TagRepository{r.db, r.bus}
			ctxDescendants, descSpan := entityTracer().Start(ctx, "repo.EntityRepository.QueryByGroup.tagDescendants",
				trace.WithAttributes(attribute.Int("query.tag_ids.count", len(q.TagIDs))))
			descendants, err := tagRepo.GetDescendantTagIDs(ctxDescendants, q.TagIDs)
			if err != nil {
				recordSpanError(descSpan, err)
				log.Warn().Err(err).Msg("failed to get descendant tags, using only direct tags")
				descendants = q.TagIDs
			} else if len(descendants) == 0 {
				descendants = q.TagIDs
			}
			descSpan.SetAttributes(attribute.Int("query.tag_descendants.count", len(descendants)))
			descSpan.End()

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

	span.SetAttributes(attribute.Int("query.predicates.and.count", len(andPredicates)))

	countCtx, countSpan := entityTracer().Start(ctx, "repo.EntityRepository.QueryByGroup.count")
	count, err := qb.Count(countCtx)
	if err != nil {
		recordSpanError(countSpan, err)
		countSpan.End()
		recordSpanError(span, err)
		return PaginationResult[EntitySummary]{}, err
	}
	countSpan.SetAttributes(attribute.Int("query.total.count", count))
	countSpan.End()

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

	fetchCtx, fetchSpan := entityTracer().Start(ctx, "repo.EntityRepository.QueryByGroup.fetch")
	entities, err := mapEntitiesSummaryErr(qb.All(fetchCtx))
	if err != nil {
		recordSpanError(fetchSpan, err)
		fetchSpan.End()
		recordSpanError(span, err)
		return PaginationResult[EntitySummary]{}, err
	}
	fetchSpan.SetAttributes(attribute.Int("query.results.count", len(entities)))
	fetchSpan.End()

	// Populate ItemCount for location-type entities
	if q.IsLocation != nil && *q.IsLocation && len(entities) > 0 {
		childCtx, childSpan := entityTracer().Start(ctx, "repo.EntityRepository.QueryByGroup.childItemCounts",
			trace.WithAttributes(attribute.Int("locations.count", len(entities))))
		ids := lo.Map(entities, func(e EntitySummary, _ int) uuid.UUID { return e.ID })
		counts, cErr := r.getChildItemCounts(childCtx, gid, ids)
		if cErr != nil {
			recordSpanError(childSpan, cErr)
		} else {
			for i := range entities {
				if c, ok := counts[entities[i].ID]; ok {
					entities[i].ItemCount = c
				}
			}
		}
		childSpan.End()
	}

	span.SetAttributes(
		attribute.Int("query.results.count", len(entities)),
		attribute.Int("query.total.count", count),
	)

	return PaginationResult[EntitySummary]{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    count,
		Items:    entities,
	}, nil
}

// getChildItemCounts returns a map of entity ID → sum of child item quantities for the given location IDs.
func (r *EntityRepository) getChildItemCounts(ctx context.Context, gid uuid.UUID, locationIDs []uuid.UUID) (map[uuid.UUID]float64, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.getChildItemCounts",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Int("locations.count", len(locationIDs)),
		))
	defer span.End()

	if len(locationIDs) == 0 {
		return nil, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(locationIDs))
	args := make([]any, 0, len(locationIDs)+1)
	args = append(args, gid)
	for i, id := range locationIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT e.entity_children, COALESCE(SUM(e.quantity), 0)
		FROM entities e
		JOIN entity_types et ON et.id = e.entity_type_entities
		WHERE e.group_entities = $1
			AND et.is_location = false
			AND e.archived = false
			AND e.entity_children IN (%s)
		GROUP BY e.entity_children
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Sql().QueryContext(ctx, query, args...)
	if err != nil {
		recordSpanError(span, err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[uuid.UUID]float64)
	for rows.Next() {
		var parentID uuid.UUID
		var count float64
		if err := rows.Scan(&parentID, &count); err != nil {
			recordSpanError(span, err)
			return nil, err
		}
		result[parentID] = count
	}
	if err := rows.Err(); err != nil {
		recordSpanError(span, err)
		return result, err
	}
	span.SetAttributes(attribute.Int("results.count", len(result)))
	return result, nil
}

// QueryByAssetID returns entities by asset ID.
func (r *EntityRepository) QueryByAssetID(ctx context.Context, gid uuid.UUID, assetID AssetID, page int, pageSize int) (PaginationResult[EntitySummary], error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.QueryByAssetID",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Int64("entity.asset_id", int64(assetID)),
			attribute.Int("query.page", page),
			attribute.Int("query.page_size", pageSize),
		))
	defer span.End()

	qb := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(int64(assetID)),
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
		recordSpanError(span, err)
		return PaginationResult[EntitySummary]{}, err
	}

	span.SetAttributes(attribute.Int("query.results.count", len(entities)))
	return PaginationResult[EntitySummary]{
		Page:     page,
		PageSize: pageSize,
		Total:    len(entities),
		Items:    entities,
	}, nil
}

// GetAll returns all the entities in the database with the Tags, Parent, and EntityType eager loaded.
func (r *EntityRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAll",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	out, err := mapEntitiesOutErr(r.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid))).
		WithTag().
		WithParent().
		WithEntityType().
		WithFields().
		All(ctx))
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("entities.count", len(out)))
	return out, nil
}

func (r *EntityRepository) GetAllZeroAssetID(ctx context.Context, gid uuid.UUID) ([]EntitySummary, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAllZeroAssetID",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	q := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.AssetID(0),
	).Order(
		ent.Asc(entity.FieldCreatedAt),
	)

	out, err := mapEntitiesSummaryErr(q.All(ctx))
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("entities.count", len(out)))
	return out, nil
}

func (r *EntityRepository) GetHighestAssetIDTx(ctx context.Context, tx *ent.Tx, gid uuid.UUID) (AssetID, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetHighestAssetIDTx",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Bool("tx", tx != nil),
		))
	defer span.End()

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
		recordSpanError(span, err)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("entity.asset_id.highest", result.AssetID))
	return AssetID(result.AssetID), nil
}

func (r *EntityRepository) GetHighestAssetID(ctx context.Context, gid uuid.UUID) (AssetID, error) {
	return r.GetHighestAssetIDTx(ctx, nil, gid)
}

func (r *EntityRepository) SetAssetID(ctx context.Context, gid uuid.UUID, id uuid.UUID, assetID AssetID) error {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.SetAssetID",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
			attribute.Int64("entity.asset_id", int64(assetID)),
		))
	defer span.End()

	q := r.db.Entity.Update().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.ID(id),
	)

	_, err := q.SetAssetID(int64(assetID)).Save(ctx)
	recordSpanError(span, err)
	return err
}

func validateQuantity(op string, quantity float64) error {
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return fmt.Errorf("%s: invalid quantity: must be a finite number", op)
	}

	return nil
}

func (r *EntityRepository) Create(ctx context.Context, gid uuid.UUID, data EntityCreate) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.Create",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.name", data.Name),
			attribute.Float64("entity.quantity", data.Quantity),
			attribute.Bool("entity.parent_id.set", data.ParentID != uuid.Nil),
			attribute.Bool("entity.entity_type_id.set", data.EntityTypeID != uuid.Nil),
			attribute.Int("entity.tags.count", len(data.TagIDs)),
			attribute.Int64("entity.asset_id", int64(data.AssetID)),
		))
	defer span.End()

	if err := validateQuantity("create entity", data.Quantity); err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	q := r.db.Entity.Create().
		SetImportRef(data.ImportRef).
		SetName(data.Name).
		SetQuantity(data.Quantity).
		SetDescription(data.Description).
		SetGroupID(gid).
		SetAssetID(int64(data.AssetID))

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	} else {
		// Auto-resolve default "Item" entity type for the group
		etID, err := r.resolveDefaultEntityType(ctx, gid, false)
		if err != nil {
			recordSpanError(span, err)
			return EntityOut{}, err
		}
		q.SetEntityTypeID(etID)
	}

	if len(data.TagIDs) > 0 {
		q.AddTagIDs(data.TagIDs...)
	}

	result, err := q.Save(ctx)
	if err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	span.SetAttributes(attribute.String("entity.id", result.ID.String()))
	r.publishMutationEvent(gid)
	out, err := r.GetOne(ctx, result.ID)
	recordSpanError(span, err)
	return out, err
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
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.CreateFromTemplate",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.name", data.Name),
			attribute.Float64("entity.quantity", data.Quantity),
			attribute.Bool("entity.parent_id.set", data.ParentID != uuid.Nil),
			attribute.Bool("entity.entity_type_id.set", data.EntityTypeID != uuid.Nil),
			attribute.Int("entity.tags.count", len(data.TagIDs)),
			attribute.Int("entity.fields.count", len(data.Fields)),
		))
	defer span.End()

	if err := validateQuantity("create entity from template", data.Quantity); err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		recordSpanError(span, err)
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
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	nextAssetID++
	span.SetAttributes(attribute.Int64("entity.asset_id", int64(nextAssetID)))

	// Create entity with all template data
	newEntityID := uuid.New()
	span.SetAttributes(attribute.String("entity.id", newEntityID.String()))

	entityCtx, entitySpan := entityTracer().Start(ctx, "repo.EntityRepository.CreateFromTemplate.entity")
	entityBuilder := tx.Entity.Create().
		SetID(newEntityID).
		SetName(data.Name).
		SetDescription(data.Description).
		SetQuantity(data.Quantity).
		SetGroupID(gid).
		SetAssetID(int64(nextAssetID)).
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

	_, err = entityBuilder.Save(entityCtx)
	if err != nil {
		recordSpanError(entitySpan, err)
		entitySpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	entitySpan.End()

	if len(data.Fields) > 0 {
		fieldsCtx, fieldsSpan := entityTracer().Start(ctx, "repo.EntityRepository.CreateFromTemplate.fields",
			trace.WithAttributes(attribute.Int("fields.count", len(data.Fields))))
		for _, field := range data.Fields {
			_, err = tx.EntityField.Create().
				SetEntityID(newEntityID).
				SetType(entityfield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				Save(fieldsCtx)
			if err != nil {
				wrapped := fmt.Errorf("failed to create field %s: %w", field.Name, err)
				recordSpanError(fieldsSpan, wrapped)
				fieldsSpan.End()
				recordSpanError(span, wrapped)
				return EntityOut{}, wrapped
			}
		}
		fieldsSpan.End()
	}

	_, commitSpan := entityTracer().Start(ctx, "repo.EntityRepository.CreateFromTemplate.commit")
	if err = tx.Commit(); err != nil {
		recordSpanError(commitSpan, err)
		commitSpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	commitSpan.End()
	committed = true

	r.publishMutationEvent(gid)
	out, err := r.GetOne(ctx, newEntityID)
	recordSpanError(span, err)
	return out, err
}

func (r *EntityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.Delete",
		trace.WithAttributes(attribute.String("entity.id", id.String())))
	defer span.End()

	loadCtx, loadSpan := entityTracer().Start(ctx, "repo.EntityRepository.Delete.load")
	e, err := r.db.Entity.Query().
		Where(entity.ID(id)).
		WithGroup().
		WithAttachments().
		Only(loadCtx)
	if err != nil {
		recordSpanError(loadSpan, err)
		loadSpan.End()
		recordSpanError(span, err)
		return err
	}
	loadSpan.End()

	// Get the group ID for attachment deletion
	var gid uuid.UUID
	if e.Edges.Group != nil {
		gid = e.Edges.Group.ID
	}
	span.SetAttributes(
		attribute.String("group.id", gid.String()),
		attribute.Int("entity.attachments.count", len(e.Edges.Attachments)),
	)

	if len(e.Edges.Attachments) > 0 {
		attCtx, attSpan := entityTracer().Start(ctx, "repo.EntityRepository.Delete.attachments",
			trace.WithAttributes(attribute.Int("attachments.count", len(e.Edges.Attachments))))
		for _, att := range e.Edges.Attachments {
			err := r.attachments.Delete(attCtx, gid, att.ID)
			if err != nil {
				recordSpanError(attSpan, err)
				log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during entity deletion")
			}
		}
		attSpan.End()
	}

	_, deleteSpan := entityTracer().Start(ctx, "repo.EntityRepository.Delete.entity")
	err = r.db.Entity.DeleteOneID(id).Exec(ctx)
	if err != nil {
		recordSpanError(deleteSpan, err)
		deleteSpan.End()
		recordSpanError(span, err)
		return err
	}
	deleteSpan.End()

	r.publishMutationEvent(id)
	return nil
}

func (r *EntityRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.DeleteByGroup",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	loadCtx, loadSpan := entityTracer().Start(ctx, "repo.EntityRepository.DeleteByGroup.load")
	e, err := r.db.Entity.Query().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).
		WithAttachments().
		Only(loadCtx)
	if err != nil {
		recordSpanError(loadSpan, err)
		loadSpan.End()
		recordSpanError(span, err)
		return err
	}
	loadSpan.End()

	span.SetAttributes(attribute.Int("entity.attachments.count", len(e.Edges.Attachments)))

	if len(e.Edges.Attachments) > 0 {
		attCtx, attSpan := entityTracer().Start(ctx, "repo.EntityRepository.DeleteByGroup.attachments",
			trace.WithAttributes(attribute.Int("attachments.count", len(e.Edges.Attachments))))
		for _, att := range e.Edges.Attachments {
			err := r.attachments.Delete(attCtx, gid, att.ID)
			if err != nil {
				recordSpanError(attSpan, err)
				log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during entity deletion")
			}
		}
		attSpan.End()
	}

	_, deleteSpan := entityTracer().Start(ctx, "repo.EntityRepository.DeleteByGroup.entity")
	_, err = r.db.Entity.
		Delete().
		Where(
			entity.ID(id),
			entity.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		recordSpanError(deleteSpan, err)
		deleteSpan.End()
		recordSpanError(span, err)
		return err
	}
	deleteSpan.End()

	r.publishMutationEvent(gid)
	return nil
}

func (r *EntityRepository) WipeInventory(ctx context.Context, gid uuid.UUID, wipeTags bool, wipeContainers bool, wipeMaintenance bool) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Bool("wipe.tags", wipeTags),
			attribute.Bool("wipe.containers", wipeContainers),
			attribute.Bool("wipe.maintenance", wipeMaintenance),
		))
	defer span.End()

	deleted := 0

	// Wipe maintenance records if requested
	// IMPORTANT: Must delete maintenance records BEFORE entities since they are linked to entities
	if wipeMaintenance {
		maintCtx, maintSpan := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory.maintenance")
		maintenanceCount, err := r.db.MaintenanceEntry.Delete().
			Where(maintenanceentry.HasEntityWith(entity.HasGroupWith(group.ID(gid)))).
			Exec(maintCtx)
		if err != nil {
			recordSpanError(maintSpan, err)
			log.Err(err).Msg("failed to delete maintenance entries during wipe inventory")
		} else {
			maintSpan.SetAttributes(attribute.Int("deleted.count", maintenanceCount))
			log.Info().Int("count", maintenanceCount).Msg("deleted maintenance entries during wipe inventory")
			deleted += maintenanceCount
		}
		maintSpan.End()
	}

	loadCtx, loadSpan := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory.loadEntities")
	entities, err := r.db.Entity.Query().
		Where(entity.HasGroupWith(group.ID(gid))).
		WithAttachments().
		All(loadCtx)
	if err != nil {
		recordSpanError(loadSpan, err)
		loadSpan.End()
		recordSpanError(span, err)
		return 0, err
	}
	loadSpan.SetAttributes(attribute.Int("entities.count", len(entities)))
	loadSpan.End()

	entCtx, entSpan := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory.deleteEntities",
		trace.WithAttributes(attribute.Int("entities.count", len(entities))))
	entityDeleted := 0
	attachmentsDeleted := 0
	for _, e := range entities {
		for _, att := range e.Edges.Attachments {
			err := r.attachments.Delete(entCtx, gid, att.ID)
			if err != nil {
				recordSpanError(entSpan, err)
				log.Err(err).Str("attachment_id", att.ID.String()).Msg("failed to delete attachment during wipe inventory")
				continue
			}
			attachmentsDeleted++
		}

		_, err = r.db.Entity.
			Delete().
			Where(
				entity.ID(e.ID),
				entity.HasGroupWith(group.ID(gid)),
			).Exec(entCtx)
		if err != nil {
			recordSpanError(entSpan, err)
			log.Err(err).Str("entity_id", e.ID.String()).Msg("failed to delete entity during wipe inventory")
			continue
		}

		entityDeleted++
	}
	entSpan.SetAttributes(
		attribute.Int("entities.deleted.count", entityDeleted),
		attribute.Int("attachments.deleted.count", attachmentsDeleted),
	)
	entSpan.End()
	deleted += entityDeleted

	// Wipe tags if requested
	if wipeTags {
		tagCtx, tagSpan := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory.tags")
		tagCount, err := r.db.Tag.Delete().Where(tag.HasGroupWith(group.ID(gid))).Exec(tagCtx)
		if err != nil {
			recordSpanError(tagSpan, err)
			log.Err(err).Msg("failed to delete tags during wipe inventory")
		} else {
			tagSpan.SetAttributes(attribute.Int("deleted.count", tagCount))
			log.Info().Int("count", tagCount).Msg("deleted tags during wipe inventory")
			deleted += tagCount
		}
		tagSpan.End()
	}

	// Wipe containers (location-type entities) if requested
	if wipeContainers {
		containerCtx, containerSpan := entityTracer().Start(ctx, "repo.EntityRepository.WipeInventory.containers")
		containerCount, err := r.db.Entity.Delete().
			Where(
				entity.HasGroupWith(group.ID(gid)),
				entity.HasEntityTypeWith(entitytype.IsLocation(true)),
			).Exec(containerCtx)
		if err != nil {
			recordSpanError(containerSpan, err)
			log.Err(err).Msg("failed to delete containers during wipe inventory")
		} else {
			containerSpan.SetAttributes(attribute.Int("deleted.count", containerCount))
			log.Info().Int("count", containerCount).Msg("deleted containers during wipe inventory")
			deleted += containerCount
		}
		containerSpan.End()
	}

	span.SetAttributes(attribute.Int("deleted.count.total", deleted))
	r.publishMutationEvent(gid)
	return deleted, nil
}

func (r *EntityRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data EntityUpdate) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.UpdateByGroup",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", data.ID.String()),
			attribute.String("entity.name", data.Name),
			attribute.Float64("entity.quantity", data.Quantity),
			attribute.Bool("entity.archived", data.Archived),
			attribute.Bool("entity.parent_id.set", data.ParentID != uuid.Nil),
			attribute.Bool("entity.entity_type_id.set", data.EntityTypeID != uuid.Nil),
			attribute.Bool("entity.sync_child_locations", data.SyncChildEntityLocations),
			attribute.Int("entity.tags.count", len(data.TagIDs)),
			attribute.Int("entity.fields.count", len(data.Fields)),
		))
	defer span.End()

	if err := validateQuantity("update entity", data.Quantity); err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	q := r.db.Entity.Update().Where(entity.ID(data.ID), entity.HasGroupWith(group.ID(gid))).
		SetName(data.Name).
		SetDescription(data.Description).
		SetSerialNumber(data.SerialNumber).
		SetModelNumber(data.ModelNumber).
		SetManufacturer(data.Manufacturer).
		SetArchived(data.Archived).
		SetPurchaseFrom(data.PurchaseFrom).
		SetPurchasePrice(data.PurchasePrice).
		SetSoldTo(data.SoldTo).
		SetSoldPrice(data.SoldPrice).
		SetSoldNotes(data.SoldNotes).
		SetNotes(data.Notes).
		SetLifetimeWarranty(data.LifetimeWarranty).
		SetInsured(data.Insured).
		SetWarrantyDetails(data.WarrantyDetails).
		SetQuantity(data.Quantity).
		SetAssetID(int64(data.AssetID)).
		SetSyncChildEntityLocations(data.SyncChildEntityLocations)

	// Date fields are nullable. Writing types.Date{}.Time() would persist
	// the 0001-01-01 sentinel that ZeroOutTimeFields then has to chase —
	// clear the column instead so absent dates round-trip as NULL/"".
	if t := data.PurchaseDate.Time(); t.IsZero() {
		q.ClearPurchaseDate()
	} else {
		q.SetPurchaseDate(t)
	}
	if t := data.SoldDate.Time(); t.IsZero() {
		q.ClearSoldDate()
	} else {
		q.SetSoldDate(t)
	}
	if t := data.WarrantyExpires.Time(); t.IsZero() {
		q.ClearWarrantyExpires()
	} else {
		q.SetWarrantyExpires(t)
	}

	if data.EntityTypeID != uuid.Nil {
		q.SetEntityTypeID(data.EntityTypeID)
	}

	tagsCtx, tagsSpan := entityTracer().Start(ctx, "repo.EntityRepository.UpdateByGroup.tags")
	currentTags, err := r.db.Entity.Query().Where(entity.ID(data.ID)).QueryTag().All(tagsCtx)
	if err != nil {
		recordSpanError(tagsSpan, err)
		tagsSpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	set := newIDSet(currentTags)

	added := 0
	for _, l := range data.TagIDs {
		if set.Contains(l) {
			set.Remove(l)
			continue
		}
		q.AddTagIDs(l)
		added++
	}

	removed := set.Len()
	if removed > 0 {
		q.RemoveTagIDs(set.Slice()...)
	}
	tagsSpan.SetAttributes(
		attribute.Int("tags.added.count", added),
		attribute.Int("tags.removed.count", removed),
	)
	tagsSpan.End()

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	if data.SyncChildEntityLocations {
		syncCtx, syncSpan := entityTracer().Start(ctx, "repo.EntityRepository.UpdateByGroup.syncChildLocations")
		children, err := r.db.Entity.Query().Where(entity.ID(data.ID)).QueryChildren().All(syncCtx)
		if err != nil {
			recordSpanError(syncSpan, err)
			syncSpan.End()
			recordSpanError(span, err)
			return EntityOut{}, err
		}

		syncSpan.SetAttributes(attribute.Int("children.count", len(children)))
		updatedCount := 0
		for _, child := range children {
			if data.ParentID != uuid.Nil {
				childParent, err := child.QueryParent().First(syncCtx)
				if err != nil || childParent.ID != data.ParentID {
					err = child.Update().SetParentID(data.ParentID).Exec(syncCtx)
					if err != nil {
						recordSpanError(syncSpan, err)
						syncSpan.End()
						recordSpanError(span, err)
						return EntityOut{}, err
					}
					updatedCount++
				}
			}
		}
		syncSpan.SetAttributes(attribute.Int("children.updated.count", updatedCount))
		syncSpan.End()
	}

	_, execSpan := entityTracer().Start(ctx, "repo.EntityRepository.UpdateByGroup.exec")
	err = q.Exec(ctx)
	if err != nil {
		recordSpanError(execSpan, err)
		execSpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	execSpan.End()

	fieldsCtx, fieldsSpan := entityTracer().Start(ctx, "repo.EntityRepository.UpdateByGroup.fields",
		trace.WithAttributes(attribute.Int("fields.input.count", len(data.Fields))))
	fields, err := r.db.EntityField.Query().Where(entityfield.HasEntityWith(entity.ID(data.ID))).All(fieldsCtx)
	if err != nil {
		recordSpanError(fieldsSpan, err)
		fieldsSpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	fieldIds := newIDSet(fields)
	fieldsSpan.SetAttributes(attribute.Int("fields.existing.count", len(fields)))

	createdFields := 0
	updatedFields := 0
	// Update Existing Fields
	for _, f := range data.Fields {
		if f.ID == uuid.Nil {
			_, err = r.db.EntityField.Create().
				SetEntityID(data.ID).
				SetType(entityfield.Type(f.Type)).
				SetName(f.Name).
				SetTextValue(f.TextValue).
				SetNumberValue(f.NumberValue).
				SetBooleanValue(f.BooleanValue).
				Save(fieldsCtx)
			if err != nil {
				recordSpanError(fieldsSpan, err)
				fieldsSpan.End()
				recordSpanError(span, err)
				return EntityOut{}, err
			}
			createdFields++
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

		_, err = opt.Save(fieldsCtx)
		if err != nil {
			recordSpanError(fieldsSpan, err)
			fieldsSpan.End()
			recordSpanError(span, err)
			return EntityOut{}, err
		}
		updatedFields++

		fieldIds.Remove(f.ID)
		continue
	}

	deletedFields := 0
	if fieldIds.Len() > 0 {
		deletedFields, err = r.db.EntityField.Delete().
			Where(
				entityfield.IDIn(fieldIds.Slice()...),
				entityfield.HasEntityWith(entity.ID(data.ID)),
			).Exec(fieldsCtx)
		if err != nil {
			recordSpanError(fieldsSpan, err)
			fieldsSpan.End()
			recordSpanError(span, err)
			return EntityOut{}, err
		}
	}
	fieldsSpan.SetAttributes(
		attribute.Int("fields.created.count", createdFields),
		attribute.Int("fields.updated.count", updatedFields),
		attribute.Int("fields.deleted.count", deletedFields),
	)
	fieldsSpan.End()

	r.publishMutationEvent(gid)
	out, err := r.GetOne(ctx, data.ID)
	recordSpanError(span, err)
	return out, err
}

func (r *EntityRepository) GetAllZeroImportRef(ctx context.Context, gid uuid.UUID) ([]uuid.UUID, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAllZeroImportRef",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

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
		recordSpanError(span, err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("entities.count", len(ids)))
	return ids, nil
}

func (r *EntityRepository) Patch(ctx context.Context, gid, id uuid.UUID, data EntityPatch) error {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.Patch",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
			attribute.Bool("patch.import_ref.set", data.ImportRef != nil),
			attribute.Bool("patch.quantity.set", data.Quantity != nil),
			attribute.Bool("patch.parent_id.set", data.ParentID != uuid.Nil),
			attribute.Bool("patch.entity_type_id.set", data.EntityTypeID != uuid.Nil),
			attribute.Bool("patch.tag_ids.set", data.TagIDs != nil),
		))
	defer span.End()

	tx, err := r.db.Tx(ctx)
	if err != nil {
		recordSpanError(span, err)
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
			recordSpanError(span, err)
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

	_, execSpan := entityTracer().Start(ctx, "repo.EntityRepository.Patch.exec")
	err = q.Exec(ctx)
	if err != nil {
		recordSpanError(execSpan, err)
		execSpan.End()
		recordSpanError(span, err)
		return err
	}
	execSpan.End()

	if data.TagIDs != nil {
		tagsCtx, tagsSpan := entityTracer().Start(ctx, "repo.EntityRepository.Patch.tags",
			trace.WithAttributes(attribute.Int("tags.input.count", len(data.TagIDs))))
		currentTags, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryTag().All(tagsCtx)
		if err != nil {
			recordSpanError(tagsSpan, err)
			tagsSpan.End()
			recordSpanError(span, err)
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
				Exec(tagsCtx); err != nil {
				recordSpanError(tagsSpan, err)
				tagsSpan.End()
				recordSpanError(span, err)
				return err
			}
		}
		if set.Len() > 0 {
			if err := tx.Entity.Update().
				Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
				RemoveTagIDs(set.Slice()...).
				Exec(tagsCtx); err != nil {
				recordSpanError(tagsSpan, err)
				tagsSpan.End()
				recordSpanError(span, err)
				return err
			}
		}
		tagsSpan.SetAttributes(
			attribute.Int("tags.added.count", len(addTags)),
			attribute.Int("tags.removed.count", set.Len()),
		)
		tagsSpan.End()
	}

	if data.ParentID != uuid.Nil {
		syncCtx, syncSpan := entityTracer().Start(ctx, "repo.EntityRepository.Patch.syncChildLocations")
		entityEnt, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).Only(syncCtx)
		if err != nil {
			recordSpanError(syncSpan, err)
			syncSpan.End()
			recordSpanError(span, err)
			return err
		}
		syncSpan.SetAttributes(attribute.Bool("entity.sync_child_locations", entityEnt.SyncChildEntityLocations))
		if entityEnt.SyncChildEntityLocations {
			children, err := tx.Entity.Query().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).QueryChildren().All(syncCtx)
			if err != nil {
				recordSpanError(syncSpan, err)
				syncSpan.End()
				recordSpanError(span, err)
				return err
			}
			updatedCount := 0
			for _, child := range children {
				childParent, err := child.QueryParent().First(syncCtx)
				if err != nil || childParent.ID != data.ParentID {
					err = child.Update().SetParentID(data.ParentID).Exec(syncCtx)
					if err != nil {
						recordSpanError(syncSpan, err)
						syncSpan.End()
						recordSpanError(span, err)
						return err
					}
					updatedCount++
				}
			}
			syncSpan.SetAttributes(
				attribute.Int("children.count", len(children)),
				attribute.Int("children.updated.count", updatedCount),
			)
		}
		syncSpan.End()
	}

	_, commitSpan := entityTracer().Start(ctx, "repo.EntityRepository.Patch.commit")
	if err := tx.Commit(); err != nil {
		recordSpanError(commitSpan, err)
		commitSpan.End()
		recordSpanError(span, err)
		return err
	}
	commitSpan.End()
	committed = true

	r.publishMutationEvent(gid)
	return nil
}

func (r *EntityRepository) GetAllCustomFieldValues(ctx context.Context, gid uuid.UUID, name string) ([]string, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAllCustomFieldValues",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("field.name", name),
		))
	defer span.End()

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
		wrapped := fmt.Errorf("failed to get field values: %w", err)
		recordSpanError(span, wrapped)
		return nil, wrapped
	}

	valueStrings := lo.Map(values, func(f st, _ int) string {
		return f.Value
	})

	span.SetAttributes(attribute.Int("values.count", len(valueStrings)))
	return valueStrings, nil
}

func (r *EntityRepository) GetAllCustomFieldNames(ctx context.Context, gid uuid.UUID) ([]string, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAllCustomFieldNames",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

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
		wrapped := fmt.Errorf("failed to get custom fields: %w", err)
		recordSpanError(span, wrapped)
		return nil, wrapped
	}

	fieldNames := lo.Map(fields, func(f st, _ int) string {
		return f.Name
	})

	span.SetAttributes(attribute.Int("names.count", len(fieldNames)))
	return fieldNames, nil
}

// ZeroOutTimeFields sets all date fields to the beginning of the day.
func (r *EntityRepository) ZeroOutTimeFields(ctx context.Context, gid uuid.UUID) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.ZeroOutTimeFields",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	q := r.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.Or(
			entity.PurchaseDateNotNil(),
			entity.PurchaseFromLT("0002-01-01"),
			entity.SoldDateNotNil(),
			entity.SoldToLT("0002-01-01"),
			entity.WarrantyExpiresNotNil(),
			entity.WarrantyDetailsLT("0002-01-01"),
		),
	)

	loadCtx, loadSpan := entityTracer().Start(ctx, "repo.EntityRepository.ZeroOutTimeFields.load")
	entities, err := q.All(loadCtx)
	if err != nil {
		wrapped := fmt.Errorf("ZeroOutTimeFields() -> failed to get entities: %w", err)
		recordSpanError(loadSpan, wrapped)
		loadSpan.End()
		recordSpanError(span, wrapped)
		return -1, wrapped
	}
	loadSpan.SetAttributes(attribute.Int("entities.count", len(entities)))
	loadSpan.End()

	toDateOnly := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	_, updateSpan := entityTracer().Start(ctx, "repo.EntityRepository.ZeroOutTimeFields.update",
		trace.WithAttributes(attribute.Int("entities.count", len(entities))))
	defer func() {
		updateSpan.End()
	}()

	updated := 0

	for _, e := range entities {
		updateQ := r.db.Entity.Update().Where(entity.ID(e.ID))

		if !e.PurchaseDate.IsZero() {
			switch {
			case e.PurchaseDate.Year() < 100:
				updateQ.ClearPurchaseDate()
			default:
				updateQ.SetPurchaseDate(toDateOnly(e.PurchaseDate))
			}
		} else {
			updateQ.ClearPurchaseDate()
		}

		if !e.SoldDate.IsZero() {
			switch {
			case e.SoldDate.Year() < 100:
				updateQ.ClearSoldDate()
			default:
				updateQ.SetSoldDate(toDateOnly(e.SoldDate))
			}
		} else {
			updateQ.ClearSoldDate()
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
			wrapped := fmt.Errorf("ZeroOutTimeFields() -> failed to update entity: %w", err)
			recordSpanError(updateSpan, wrapped)
			recordSpanError(span, wrapped)
			return updated, wrapped
		}

		updated++
	}

	updateSpan.SetAttributes(attribute.Int("entities.updated.count", updated))
	span.SetAttributes(attribute.Int("entities.updated.count", updated))
	return updated, nil
}

func (r *EntityRepository) SetPrimaryPhotos(ctx context.Context, gid uuid.UUID) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.SetPrimaryPhotos",
		trace.WithAttributes(attribute.String("group.id", gid.String())))
	defer span.End()

	loadCtx, loadSpan := entityTracer().Start(ctx, "repo.EntityRepository.SetPrimaryPhotos.load")
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
		IDs(loadCtx)
	if err != nil {
		recordSpanError(loadSpan, err)
		loadSpan.End()
		recordSpanError(span, err)
		return -1, err
	}
	loadSpan.SetAttributes(attribute.Int("entities.count", len(entityIDs)))
	loadSpan.End()

	updateCtx, updateSpan := entityTracer().Start(ctx, "repo.EntityRepository.SetPrimaryPhotos.update",
		trace.WithAttributes(attribute.Int("entities.count", len(entityIDs))))
	defer updateSpan.End()

	updated := 0
	for _, id := range entityIDs {
		a, err := r.db.Attachment.Query().
			Where(
				attachment.HasEntityWith(entity.ID(id)),
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Primary(false),
			).
			First(updateCtx)
		if err != nil {
			recordSpanError(updateSpan, err)
			recordSpanError(span, err)
			return updated, err
		}

		_, err = r.db.Attachment.UpdateOne(a).
			SetPrimary(true).
			Save(updateCtx)
		if err != nil {
			recordSpanError(updateSpan, err)
			recordSpanError(span, err)
			return updated, err
		}

		updated++
	}

	updateSpan.SetAttributes(attribute.Int("entities.updated.count", updated))
	span.SetAttributes(attribute.Int("entities.updated.count", updated))
	return updated, nil
}

// Duplicate creates a copy of an entity with configurable options for what data to copy.
func (r *EntityRepository) Duplicate(ctx context.Context, gid, id uuid.UUID, options DuplicateOptions) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.source_id", id.String()),
			attribute.Bool("options.copy_maintenance", options.CopyMaintenance),
			attribute.Bool("options.copy_attachments", options.CopyAttachments),
			attribute.Bool("options.copy_custom_fields", options.CopyCustomFields),
			attribute.String("options.copy_prefix", options.CopyPrefix),
		))
	defer span.End()

	tx, err := r.db.Tx(ctx)
	if err != nil {
		recordSpanError(span, err)
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
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	nextAssetID, err := r.GetHighestAssetIDTx(ctx, tx, gid)
	if err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	nextAssetID++

	if options.CopyPrefix == "" {
		options.CopyPrefix = "Copy of "
	}

	newEntityID := uuid.New()
	span.SetAttributes(
		attribute.String("entity.new_id", newEntityID.String()),
		attribute.Int64("entity.new_asset_id", int64(nextAssetID)),
	)

	entityCtx, entitySpan := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate.entity")
	entityBuilder := tx.Entity.Create().
		SetID(newEntityID).
		SetName(options.CopyPrefix + originalEntity.Name).
		SetDescription(originalEntity.Description).
		SetQuantity(originalEntity.Quantity).
		SetGroupID(gid).
		SetAssetID(int64(nextAssetID)).
		SetSerialNumber(originalEntity.SerialNumber).
		SetModelNumber(originalEntity.ModelNumber).
		SetManufacturer(originalEntity.Manufacturer).
		SetLifetimeWarranty(originalEntity.LifetimeWarranty).
		SetWarrantyDetails(originalEntity.WarrantyDetails).
		SetPurchaseFrom(originalEntity.PurchaseFrom).
		SetPurchasePrice(originalEntity.PurchasePrice).
		SetSoldTo(originalEntity.SoldTo).
		SetSoldPrice(originalEntity.SoldPrice).
		SetSoldNotes(originalEntity.SoldNotes).
		SetNotes(originalEntity.Notes).
		SetInsured(originalEntity.Insured).
		SetArchived(originalEntity.Archived).
		SetSyncChildEntityLocations(originalEntity.SyncChildEntityLocations)

	// Skip Set on zero dates so the duplicate's nullable date columns end up
	// NULL rather than the 0001-01-01 sentinel.
	if t := originalEntity.PurchaseDate.Time(); !t.IsZero() {
		entityBuilder.SetPurchaseDate(t)
	}
	if t := originalEntity.SoldDate.Time(); !t.IsZero() {
		entityBuilder.SetSoldDate(t)
	}
	if t := originalEntity.WarrantyExpires.Time(); !t.IsZero() {
		entityBuilder.SetWarrantyExpires(t)
	}

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

	_, err = entityBuilder.Save(entityCtx)
	if err != nil {
		recordSpanError(entitySpan, err)
		entitySpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	entitySpan.End()

	// Copy custom fields if requested
	if options.CopyCustomFields {
		fieldsCtx, fieldsSpan := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate.fields",
			trace.WithAttributes(attribute.Int("fields.count", len(originalEntity.Fields))))
		copied := 0
		for _, field := range originalEntity.Fields {
			_, err = tx.EntityField.Create().
				SetEntityID(newEntityID).
				SetType(entityfield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				SetNumberValue(field.NumberValue).
				SetBooleanValue(field.BooleanValue).
				Save(fieldsCtx)
			if err != nil {
				recordSpanError(fieldsSpan, err)
				log.Warn().Err(err).Str("field_name", field.Name).Msg("failed to copy custom field during duplication")
				continue
			}
			copied++
		}
		fieldsSpan.SetAttributes(attribute.Int("fields.copied.count", copied))
		fieldsSpan.End()
	}

	// Copy attachments if requested
	if options.CopyAttachments {
		attCtx, attSpan := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate.attachments",
			trace.WithAttributes(attribute.Int("attachments.count", len(originalEntity.Attachments))))
		copied := 0
		for _, att := range originalEntity.Attachments {
			originalAttachment, err := tx.Attachment.Query().
				Where(attachment.ID(att.ID)).
				Only(attCtx)
			if err != nil {
				recordSpanError(attSpan, err)
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
				Save(attCtx)
			if err != nil {
				recordSpanError(attSpan, err)
				log.Warn().Err(err).Str("original_attachment_id", att.ID.String()).Msg("failed to copy attachment during duplication")
				continue
			}
			copied++
		}
		attSpan.SetAttributes(attribute.Int("attachments.copied.count", copied))
		attSpan.End()
	}

	// Copy maintenance entries if requested
	if options.CopyMaintenance {
		maintCtx, maintSpan := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate.maintenance")
		maintenanceEntries, err := tx.MaintenanceEntry.Query().
			Where(maintenanceentry.HasEntityWith(entity.ID(id))).
			All(maintCtx)
		if err != nil {
			recordSpanError(maintSpan, err)
		} else {
			maintSpan.SetAttributes(attribute.Int("maintenance.count", len(maintenanceEntries)))
			copied := 0
			for _, entry := range maintenanceEntries {
				_, err = tx.MaintenanceEntry.Create().
					SetEntityID(newEntityID).
					SetDate(entry.Date).
					SetScheduledDate(entry.ScheduledDate).
					SetName(entry.Name).
					SetDescription(entry.Description).
					SetCost(entry.Cost).
					Save(maintCtx)
				if err != nil {
					recordSpanError(maintSpan, err)
					log.Warn().Err(err).Str("maintenance_entry_id", entry.ID.String()).Msg("failed to copy maintenance entry during duplication")
					continue
				}
				copied++
			}
			maintSpan.SetAttributes(attribute.Int("maintenance.copied.count", copied))
		}
		maintSpan.End()
	}

	_, commitSpan := entityTracer().Start(ctx, "repo.EntityRepository.Duplicate.commit")
	if err := tx.Commit(); err != nil {
		recordSpanError(commitSpan, err)
		commitSpan.End()
		recordSpanError(span, err)
		return EntityOut{}, err
	}
	commitSpan.End()
	committed = true

	r.publishMutationEvent(gid)
	out, err := r.GetOne(ctx, newEntityID)
	recordSpanError(span, err)
	return out, err
}

// ============================================================================
// Container / Location methods (absorbed from LocationRepository)
// ============================================================================

type ContainerQuery struct {
	FilterChildren bool `json:"filterChildren" schema:"filterChildren"`
}

// GetAllContainers returns all container entities (entity_type.is_location = true) with child entity counts.
func (r *EntityRepository) GetAllContainers(ctx context.Context, gid uuid.UUID, filter ContainerQuery) ([]EntityOutCount, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetAllContainers",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Bool("filter.children", filter.FilterChildren),
		))
	defer span.End()

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
		recordSpanError(span, err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	list := []EntityOutCount{}
	for rows.Next() {
		var ct EntityOutCount

		var maybeCount *float64

		err := rows.Scan(&ct.ID, &ct.Name, &ct.Description, &ct.CreatedAt, &ct.UpdatedAt, &maybeCount)
		if err != nil {
			recordSpanError(span, err)
			return nil, err
		}

		if maybeCount != nil {
			ct.ItemCount = *maybeCount
		}

		list = append(list, ct)
	}

	if err := rows.Err(); err != nil {
		recordSpanError(span, err)
		return list, err
	}
	span.SetAttributes(attribute.Int("containers.count", len(list)))
	return list, nil
}

// GetContainerByGroup returns a single container entity by ID, verified to belong to a specific group.
func (r *EntityRepository) GetContainerByGroup(ctx context.Context, gid, id uuid.UUID) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.GetContainerByGroup",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	out, err := mapEntityOutErr(r.db.Entity.Query().
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
	if err != nil {
		recordSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("entity.children.count", len(out.Children)))
	return out, nil
}

// CreateContainer creates a container entity (with a location-type entity_type).
func (r *EntityRepository) CreateContainer(ctx context.Context, gid uuid.UUID, data EntityCreate) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.CreateContainer",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.name", data.Name),
			attribute.Bool("entity.parent_id.set", data.ParentID != uuid.Nil),
			attribute.Bool("entity.entity_type_id.set", data.EntityTypeID != uuid.Nil),
		))
	defer span.End()

	if data.ParentID != uuid.Nil {
		validateCtx, validateSpan := entityTracer().Start(ctx, "repo.EntityRepository.CreateContainer.validateParent")
		parentEntity, err := r.db.Entity.Query().
			Where(entity.ID(data.ParentID)).
			WithEntityType().
			Only(validateCtx)
		if err != nil {
			wrapped := fmt.Errorf("parent entity not found: %w", err)
			recordSpanError(validateSpan, wrapped)
			validateSpan.End()
			recordSpanError(span, wrapped)
			return EntityOut{}, wrapped
		}
		if parentEntity.Edges.EntityType == nil || !parentEntity.Edges.EntityType.IsLocation {
			wrapped := fmt.Errorf("locations can only have other locations as parents, not items")
			recordSpanError(validateSpan, wrapped)
			validateSpan.End()
			recordSpanError(span, wrapped)
			return EntityOut{}, wrapped
		}
		validateSpan.End()
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
			recordSpanError(span, err)
			return EntityOut{}, err
		}
		q.SetEntityTypeID(etID)
	}

	result, err := q.Save(ctx)
	if err != nil {
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	span.SetAttributes(attribute.String("entity.id", result.ID.String()))
	result.Edges.Group = &ent.Group{ID: gid}
	r.publishMutationEvent(gid)
	return mapEntityOut(result), nil
}

// UpdateContainer updates a container entity.
func (r *EntityRepository) UpdateContainer(ctx context.Context, gid, id uuid.UUID, data EntityUpdate) (EntityOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.UpdateContainer",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
			attribute.String("entity.name", data.Name),
			attribute.Bool("entity.parent_id.set", data.ParentID != uuid.Nil),
		))
	defer span.End()

	if data.ParentID != uuid.Nil {
		validateCtx, validateSpan := entityTracer().Start(ctx, "repo.EntityRepository.UpdateContainer.validateParent")
		parentEntity, err := r.db.Entity.Query().
			Where(entity.ID(data.ParentID)).
			WithEntityType().
			Only(validateCtx)
		if err != nil {
			wrapped := fmt.Errorf("parent entity not found: %w", err)
			recordSpanError(validateSpan, wrapped)
			validateSpan.End()
			recordSpanError(span, wrapped)
			return EntityOut{}, wrapped
		}
		if parentEntity.Edges.EntityType == nil || !parentEntity.Edges.EntityType.IsLocation {
			wrapped := fmt.Errorf("locations can only have other locations as parents, not items")
			recordSpanError(validateSpan, wrapped)
			validateSpan.End()
			recordSpanError(span, wrapped)
			return EntityOut{}, wrapped
		}
		validateSpan.End()
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
		recordSpanError(span, err)
		return EntityOut{}, err
	}

	r.publishMutationEvent(gid)
	out, err := r.GetOne(ctx, id)
	recordSpanError(span, err)
	return out, err
}

// DeleteContainerByGroup deletes a container entity by group.
func (r *EntityRepository) DeleteContainerByGroup(ctx context.Context, gid, id uuid.UUID) error {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.DeleteContainerByGroup",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	_, err := r.db.Entity.Delete().Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).Exec(ctx)
	if err != nil {
		recordSpanError(span, err)
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
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.PathForEntity",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

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

	queryCtx, querySpan := entityTracer().Start(ctx, "repo.EntityRepository.PathForEntity.query")
	rows, err := r.db.Sql().QueryContext(queryCtx, query, entityID, gid)
	if err != nil {
		recordSpanError(querySpan, err)
		querySpan.End()
		recordSpanError(span, err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var path []EntityPath

	for rows.Next() {
		var entry EntityPath
		entry.Type = EntityPathTypeLocation
		if err := rows.Scan(&entry.ID, &entry.Name); err != nil {
			recordSpanError(querySpan, err)
			querySpan.End()
			recordSpanError(span, err)
			return nil, err
		}
		path = append(path, entry)
	}

	if err := rows.Err(); err != nil {
		recordSpanError(querySpan, err)
		querySpan.End()
		recordSpanError(span, err)
		return nil, err
	}
	querySpan.SetAttributes(attribute.Int("path.depth", len(path)))
	querySpan.End()

	// Reverse the order so that the root is first
	mutable.Reverse(path)

	span.SetAttributes(attribute.Int("path.depth", len(path)))
	return path, nil
}

func (r *EntityRepository) Tree(ctx context.Context, gid uuid.UUID, tq TreeQuery) ([]TreeItem, error) {
	ctx, span := entityTracer().Start(ctx, "repo.EntityRepository.Tree",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.Bool("query.with_items", tq.WithItems),
		))
	defer span.End()

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

	queryCtx, querySpan := entityTracer().Start(ctx, "repo.EntityRepository.Tree.query")
	rows, err := r.db.Sql().QueryContext(queryCtx, query, gid)
	if err != nil {
		recordSpanError(querySpan, err)
		querySpan.End()
		recordSpanError(span, err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var flatItems []FlatTreeItem
	for rows.Next() {
		var item FlatTreeItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Level, &item.ParentID, &item.Type); err != nil {
			recordSpanError(querySpan, err)
			querySpan.End()
			recordSpanError(span, err)
			return nil, err
		}
		flatItems = append(flatItems, item)
	}

	if err := rows.Err(); err != nil {
		recordSpanError(querySpan, err)
		querySpan.End()
		recordSpanError(span, err)
		return nil, err
	}
	querySpan.SetAttributes(attribute.Int("flat_items.count", len(flatItems)))
	querySpan.End()

	_, buildSpan := entityTracer().Start(ctx, "repo.EntityRepository.Tree.build",
		trace.WithAttributes(attribute.Int("flat_items.count", len(flatItems))))
	tree := ConvertEntitiesToTree(flatItems)
	buildSpan.SetAttributes(attribute.Int("tree.roots.count", len(tree)))
	buildSpan.End()

	span.SetAttributes(
		attribute.Int("flat_items.count", len(flatItems)),
		attribute.Int("tree.roots.count", len(tree)),
	)
	return tree, nil
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
