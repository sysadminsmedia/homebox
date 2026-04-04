package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/templatefield"
)

type EntityTemplatesRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	TemplateField struct {
		ID        *uuid.UUID `json:"id,omitempty"`
		Type      string     `json:"type"`
		Name      string     `json:"name"`
		TextValue string     `json:"textValue"`
	}

	TemplateTagSummary struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	TemplateLocationSummary struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	EntityTemplateCreate struct {
		Name        string `json:"name"        validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"max=1000"`
		Notes       string `json:"notes"       validate:"max=1000"`

		// Default values for entities
		DefaultQuantity         *float64 `json:"defaultQuantity,omitempty"        extensions:"x-nullable"`
		DefaultInsured          bool     `json:"defaultInsured"`
		DefaultName             *string  `json:"defaultName,omitempty"            validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultDescription      *string  `json:"defaultDescription,omitempty"     validate:"omitempty,max=1000" extensions:"x-nullable"`
		DefaultManufacturer     *string  `json:"defaultManufacturer,omitempty"    validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultModelNumber      *string  `json:"defaultModelNumber,omitempty"     validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultLifetimeWarranty bool     `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  *string  `json:"defaultWarrantyDetails,omitempty" validate:"omitempty,max=1000" extensions:"x-nullable"`

		// Default location and tags
		DefaultLocationID uuid.UUID    `json:"defaultLocationId,omitempty" extensions:"x-nullable"`
		DefaultTagIDs     *[]uuid.UUID `json:"defaultTagIds,omitempty"     extensions:"x-nullable"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}

	EntityTemplateUpdate struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Notes       string    `json:"notes"       validate:"max=1000"`

		// Default values for entities
		DefaultQuantity         *float64 `json:"defaultQuantity,omitempty"        extensions:"x-nullable"`
		DefaultInsured          bool     `json:"defaultInsured"`
		DefaultName             *string  `json:"defaultName,omitempty"            validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultDescription      *string  `json:"defaultDescription,omitempty"     validate:"omitempty,max=1000" extensions:"x-nullable"`
		DefaultManufacturer     *string  `json:"defaultManufacturer,omitempty"    validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultModelNumber      *string  `json:"defaultModelNumber,omitempty"     validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultLifetimeWarranty bool     `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  *string  `json:"defaultWarrantyDetails,omitempty" validate:"omitempty,max=1000" extensions:"x-nullable"`

		// Default location and tags
		DefaultLocationID uuid.UUID    `json:"defaultLocationId,omitempty" extensions:"x-nullable"`
		DefaultTagIDs     *[]uuid.UUID `json:"defaultTagIds,omitempty"     extensions:"x-nullable"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}

	EntityTemplateSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	EntityTemplateOut struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Notes       string    `json:"notes"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`

		// Default values for entities
		DefaultQuantity         float64 `json:"defaultQuantity"`
		DefaultInsured          bool    `json:"defaultInsured"`
		DefaultName             string  `json:"defaultName"`
		DefaultDescription      string  `json:"defaultDescription"`
		DefaultManufacturer     string  `json:"defaultManufacturer"`
		DefaultModelNumber      string  `json:"defaultModelNumber"`
		DefaultLifetimeWarranty bool    `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  string  `json:"defaultWarrantyDetails"`

		// Default location and tags
		DefaultLocation *TemplateLocationSummary `json:"defaultLocation"`
		DefaultTags     []TemplateTagSummary     `json:"defaultTags"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}
)

func mapTemplateField(field *ent.TemplateField) TemplateField {
	return TemplateField{
		ID:        lo.ToPtr(field.ID),
		Type:      string(field.Type),
		Name:      field.Name,
		TextValue: field.TextValue,
	}
}

func mapTemplateFieldSlice(fields []*ent.TemplateField) []TemplateField {
	return lo.Map(fields, func(field *ent.TemplateField, _ int) TemplateField {
		return mapTemplateField(field)
	})
}

func mapEntityTemplateSummary(template *ent.EntityTemplate) EntityTemplateSummary {
	return EntityTemplateSummary{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func (r *EntityTemplatesRepository) mapTemplateOut(ctx context.Context, template *ent.EntityTemplate) EntityTemplateOut {
	fields := make([]TemplateField, 0)
	if template.Edges.Fields != nil {
		fields = mapTemplateFieldSlice(template.Edges.Fields)
	}

	// Map location if present
	var location *TemplateLocationSummary
	if template.Edges.Location != nil {
		location = &TemplateLocationSummary{
			ID:   template.Edges.Location.ID,
			Name: template.Edges.Location.Name,
		}
	}

	// Fetch tags from database using stored IDs
	tags := make([]TemplateTagSummary, 0)
	if len(template.DefaultTagIds) > 0 {
		tagEntities, err := r.db.Tag.Query().
			Where(tag.IDIn(template.DefaultTagIds...)).
			All(ctx)
		if err == nil {
			tags = lo.Map(tagEntities, func(l *ent.Tag, _ int) TemplateTagSummary {
				return TemplateTagSummary{
					ID:   l.ID,
					Name: l.Name,
				}
			})
		}
	}

	return EntityTemplateOut{
		ID:                      template.ID,
		Name:                    template.Name,
		Description:             template.Description,
		Notes:                   template.Notes,
		CreatedAt:               template.CreatedAt,
		UpdatedAt:               template.UpdatedAt,
		DefaultQuantity:         template.DefaultQuantity,
		DefaultInsured:          template.DefaultInsured,
		DefaultName:             template.DefaultName,
		DefaultDescription:      template.DefaultDescription,
		DefaultManufacturer:     template.DefaultManufacturer,
		DefaultModelNumber:      template.DefaultModelNumber,
		DefaultLifetimeWarranty: template.DefaultLifetimeWarranty,
		DefaultWarrantyDetails:  template.DefaultWarrantyDetails,
		DefaultLocation:         location,
		DefaultTags:             tags,
		IncludeWarrantyFields:   template.IncludeWarrantyFields,
		IncludePurchaseFields:   template.IncludePurchaseFields,
		IncludeSoldFields:       template.IncludeSoldFields,
		Fields:                  fields,
	}
}

func (r *EntityTemplatesRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventEntityMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// GetAll returns all templates for a group
func (r *EntityTemplatesRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]EntityTemplateSummary, error) {
	templates, err := r.db.EntityTemplate.Query().
		Where(entitytemplate.HasGroupWith(group.ID(gid))).
		Order(ent.Asc(entitytemplate.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := lo.Map(templates, func(template *ent.EntityTemplate, _ int) EntityTemplateSummary {
		return mapEntityTemplateSummary(template)
	})

	return result, nil
}

// GetOne returns a single template by ID, verified to belong to the specified group
func (r *EntityTemplatesRepository) GetOne(ctx context.Context, gid uuid.UUID, id uuid.UUID) (EntityTemplateOut, error) {
	template, err := r.db.EntityTemplate.Query().
		Where(
			entitytemplate.ID(id),
			entitytemplate.HasGroupWith(group.ID(gid)),
		).
		WithFields().
		WithLocation().
		Only(ctx)

	if err != nil {
		return EntityTemplateOut{}, err
	}

	return r.mapTemplateOut(ctx, template), nil
}

// Create creates a new template
func (r *EntityTemplatesRepository) Create(ctx context.Context, gid uuid.UUID, data EntityTemplateCreate) (EntityTemplateOut, error) {
	q := r.db.EntityTemplate.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetNotes(data.Notes).
		SetNillableDefaultQuantity(data.DefaultQuantity).
		SetDefaultInsured(data.DefaultInsured).
		SetNillableDefaultName(data.DefaultName).
		SetNillableDefaultDescription(data.DefaultDescription).
		SetNillableDefaultManufacturer(data.DefaultManufacturer).
		SetNillableDefaultModelNumber(data.DefaultModelNumber).
		SetDefaultLifetimeWarranty(data.DefaultLifetimeWarranty).
		SetNillableDefaultWarrantyDetails(data.DefaultWarrantyDetails).
		SetIncludeWarrantyFields(data.IncludeWarrantyFields).
		SetIncludePurchaseFields(data.IncludePurchaseFields).
		SetIncludeSoldFields(data.IncludeSoldFields).
		SetGroupID(gid)

	if data.DefaultLocationID != uuid.Nil {
		q.SetLocationID(data.DefaultLocationID)
	}
	if data.DefaultTagIDs != nil && len(*data.DefaultTagIDs) > 0 {
		q.SetDefaultTagIds(*data.DefaultTagIDs)
	}

	template, err := q.Save(ctx)
	if err != nil {
		return EntityTemplateOut{}, err
	}

	// Create template fields
	for _, field := range data.Fields {
		_, err = r.db.TemplateField.Create().
			SetEntityTemplateID(template.ID).
			SetType(templatefield.Type(field.Type)).
			SetName(field.Name).
			SetTextValue(field.TextValue).
			Save(ctx)

		if err != nil {
			log.Err(err).Msg("failed to create template field")
			return EntityTemplateOut{}, err
		}
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, template.ID)
}

// Update updates an existing template
func (r *EntityTemplatesRepository) Update(ctx context.Context, gid uuid.UUID, data EntityTemplateUpdate) (EntityTemplateOut, error) {
	// Verify template belongs to group
	template, err := r.db.EntityTemplate.Query().
		Where(
			entitytemplate.ID(data.ID),
			entitytemplate.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return EntityTemplateOut{}, err
	}

	// Update template
	updateQ := template.Update().
		SetName(data.Name).
		SetDescription(data.Description).
		SetNotes(data.Notes).
		SetNillableDefaultQuantity(data.DefaultQuantity).
		SetDefaultInsured(data.DefaultInsured).
		SetNillableDefaultName(data.DefaultName).
		SetNillableDefaultDescription(data.DefaultDescription).
		SetNillableDefaultManufacturer(data.DefaultManufacturer).
		SetNillableDefaultModelNumber(data.DefaultModelNumber).
		SetDefaultLifetimeWarranty(data.DefaultLifetimeWarranty).
		SetNillableDefaultWarrantyDetails(data.DefaultWarrantyDetails).
		SetIncludeWarrantyFields(data.IncludeWarrantyFields).
		SetIncludePurchaseFields(data.IncludePurchaseFields).
		SetIncludeSoldFields(data.IncludeSoldFields)

	// Update location: set when provided (not uuid.Nil), otherwise clear
	if data.DefaultLocationID != uuid.Nil {
		updateQ.SetLocationID(data.DefaultLocationID)
	} else {
		updateQ.ClearLocation()
	}

	// Update default tag IDs (stored as JSON)
	if data.DefaultTagIDs != nil && len(*data.DefaultTagIDs) > 0 {
		updateQ.SetDefaultTagIds(*data.DefaultTagIDs)
	} else {
		updateQ.ClearDefaultTagIds()
	}

	_, err = updateQ.Save(ctx)
	if err != nil {
		return EntityTemplateOut{}, err
	}

	// Get existing fields
	existingFields, err := r.db.TemplateField.Query().
		Where(templatefield.HasEntityTemplateWith(entitytemplate.ID(data.ID))).
		All(ctx)

	if err != nil {
		return EntityTemplateOut{}, err
	}

	// Track which fields are being updated
	updatedFieldIDs := make(map[uuid.UUID]bool)

	// Create or update fields
	for _, field := range data.Fields {
		if field.ID == nil || *field.ID == uuid.Nil {
			// Create new field
			_, err = r.db.TemplateField.Create().
				SetEntityTemplateID(data.ID).
				SetType(templatefield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				Save(ctx)

			if err != nil {
				log.Err(err).Msg("failed to create template field")
				return EntityTemplateOut{}, err
			}
		} else {
			// Update existing field
			updatedFieldIDs[*field.ID] = true
			_, err = r.db.TemplateField.Update().
				Where(
					templatefield.ID(*field.ID),
					templatefield.HasEntityTemplateWith(entitytemplate.ID(data.ID)),
				).
				SetType(templatefield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				Save(ctx)

			if err != nil {
				log.Err(err).Msg("failed to update template field")
				return EntityTemplateOut{}, err
			}
		}
	}

	// Delete fields that are no longer present
	for _, field := range existingFields {
		if !updatedFieldIDs[field.ID] {
			err = r.db.TemplateField.DeleteOne(field).Exec(ctx)
			if err != nil {
				log.Err(err).Msg("failed to delete template field")
			}
		}
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, template.ID)
}

// Delete deletes a template
func (r *EntityTemplatesRepository) Delete(ctx context.Context, gid uuid.UUID, id uuid.UUID) error {
	// Verify template belongs to group
	_, err := r.db.EntityTemplate.Query().
		Where(
			entitytemplate.ID(id),
			entitytemplate.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return err
	}

	// Delete template (fields will be cascade deleted)
	err = r.db.EntityTemplate.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return nil
}
