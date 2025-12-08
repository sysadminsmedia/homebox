package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/itemtemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/label"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/templatefield"
)

type ItemTemplatesRepository struct {
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

	TemplateLabelSummary struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	TemplateLocationSummary struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	ItemTemplateCreate struct {
		Name        string `json:"name"        validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"max=1000"`
		Notes       string `json:"notes"       validate:"max=1000"`

		// Default values for items
		DefaultQuantity         *int    `json:"defaultQuantity,omitempty"        extensions:"x-nullable"`
		DefaultInsured          bool    `json:"defaultInsured"`
		DefaultName             *string `json:"defaultName,omitempty"            validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultDescription      *string `json:"defaultDescription,omitempty"     validate:"omitempty,max=1000" extensions:"x-nullable"`
		DefaultManufacturer     *string `json:"defaultManufacturer,omitempty"    validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultModelNumber      *string `json:"defaultModelNumber,omitempty"     validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultLifetimeWarranty bool    `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  *string `json:"defaultWarrantyDetails,omitempty" validate:"omitempty,max=1000" extensions:"x-nullable"`

		// Default location and labels
		DefaultLocationID *uuid.UUID   `json:"defaultLocationId,omitempty" extensions:"x-nullable"`
		DefaultLabelIDs   *[]uuid.UUID `json:"defaultLabelIds,omitempty"   extensions:"x-nullable"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}

	ItemTemplateUpdate struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Notes       string    `json:"notes"       validate:"max=1000"`

		// Default values for items
		DefaultQuantity         *int    `json:"defaultQuantity,omitempty"        extensions:"x-nullable"`
		DefaultInsured          bool    `json:"defaultInsured"`
		DefaultName             *string `json:"defaultName,omitempty"            validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultDescription      *string `json:"defaultDescription,omitempty"     validate:"omitempty,max=1000" extensions:"x-nullable"`
		DefaultManufacturer     *string `json:"defaultManufacturer,omitempty"    validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultModelNumber      *string `json:"defaultModelNumber,omitempty"     validate:"omitempty,max=255"  extensions:"x-nullable"`
		DefaultLifetimeWarranty bool    `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  *string `json:"defaultWarrantyDetails,omitempty" validate:"omitempty,max=1000" extensions:"x-nullable"`

		// Default location and labels
		DefaultLocationID *uuid.UUID   `json:"defaultLocationId,omitempty" extensions:"x-nullable"`
		DefaultLabelIDs   *[]uuid.UUID `json:"defaultLabelIds,omitempty"   extensions:"x-nullable"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}

	ItemTemplateSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	ItemTemplateOut struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Notes       string    `json:"notes"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`

		// Default values for items
		DefaultQuantity         int    `json:"defaultQuantity"`
		DefaultInsured          bool   `json:"defaultInsured"`
		DefaultName             string `json:"defaultName"`
		DefaultDescription      string `json:"defaultDescription"`
		DefaultManufacturer     string `json:"defaultManufacturer"`
		DefaultModelNumber      string `json:"defaultModelNumber"`
		DefaultLifetimeWarranty bool   `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  string `json:"defaultWarrantyDetails"`

		// Default location and labels
		DefaultLocation *TemplateLocationSummary `json:"defaultLocation"`
		DefaultLabels   []TemplateLabelSummary   `json:"defaultLabels"`

		// Metadata flags
		IncludeWarrantyFields bool `json:"includeWarrantyFields"`
		IncludePurchaseFields bool `json:"includePurchaseFields"`
		IncludeSoldFields     bool `json:"includeSoldFields"`

		// Custom fields
		Fields []TemplateField `json:"fields"`
	}
)

func mapTemplateField(field *ent.TemplateField) TemplateField {
	id := field.ID
	return TemplateField{
		ID:        &id,
		Type:      string(field.Type),
		Name:      field.Name,
		TextValue: field.TextValue,
	}
}

func mapTemplateFieldSlice(fields []*ent.TemplateField) []TemplateField {
	result := make([]TemplateField, len(fields))
	for i, field := range fields {
		result[i] = mapTemplateField(field)
	}
	return result
}

func mapTemplateSummary(template *ent.ItemTemplate) ItemTemplateSummary {
	return ItemTemplateSummary{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func (r *ItemTemplatesRepository) mapTemplateOut(ctx context.Context, template *ent.ItemTemplate) ItemTemplateOut {
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

	// Fetch labels from database using stored IDs
	labels := make([]TemplateLabelSummary, 0)
	if len(template.DefaultLabelIds) > 0 {
		labelEntities, err := r.db.Label.Query().
			Where(label.IDIn(template.DefaultLabelIds...)).
			All(ctx)
		if err == nil {
			for _, l := range labelEntities {
				labels = append(labels, TemplateLabelSummary{
					ID:   l.ID,
					Name: l.Name,
				})
			}
		}
	}

	return ItemTemplateOut{
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
		DefaultLabels:           labels,
		IncludeWarrantyFields:   template.IncludeWarrantyFields,
		IncludePurchaseFields:   template.IncludePurchaseFields,
		IncludeSoldFields:       template.IncludeSoldFields,
		Fields:                  fields,
	}
}

func (r *ItemTemplatesRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventItemMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// GetAll returns all templates for a group
func (r *ItemTemplatesRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]ItemTemplateSummary, error) {
	templates, err := r.db.ItemTemplate.Query().
		Where(itemtemplate.HasGroupWith(group.ID(gid))).
		Order(ent.Asc(itemtemplate.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]ItemTemplateSummary, len(templates))
	for i, template := range templates {
		result[i] = mapTemplateSummary(template)
	}

	return result, nil
}

// GetOne returns a single template by ID, verified to belong to the specified group
func (r *ItemTemplatesRepository) GetOne(ctx context.Context, gid uuid.UUID, id uuid.UUID) (ItemTemplateOut, error) {
	template, err := r.db.ItemTemplate.Query().
		Where(
			itemtemplate.ID(id),
			itemtemplate.HasGroupWith(group.ID(gid)),
		).
		WithFields().
		WithLocation().
		Only(ctx)

	if err != nil {
		return ItemTemplateOut{}, err
	}

	return r.mapTemplateOut(ctx, template), nil
}

// Create creates a new template
func (r *ItemTemplatesRepository) Create(ctx context.Context, gid uuid.UUID, data ItemTemplateCreate) (ItemTemplateOut, error) {
	q := r.db.ItemTemplate.Create().
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
		SetGroupID(gid).
		SetNillableLocationID(data.DefaultLocationID)

	// Set default label IDs (stored as JSON)
	if data.DefaultLabelIDs != nil && len(*data.DefaultLabelIDs) > 0 {
		q.SetDefaultLabelIds(*data.DefaultLabelIDs)
	}

	template, err := q.Save(ctx)
	if err != nil {
		return ItemTemplateOut{}, err
	}

	// Create template fields
	for _, field := range data.Fields {
		_, err = r.db.TemplateField.Create().
			SetItemTemplateID(template.ID).
			SetType(templatefield.Type(field.Type)).
			SetName(field.Name).
			SetTextValue(field.TextValue).
			Save(ctx)

		if err != nil {
			log.Err(err).Msg("failed to create template field")
			return ItemTemplateOut{}, err
		}
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, template.ID)
}

// Update updates an existing template
func (r *ItemTemplatesRepository) Update(ctx context.Context, gid uuid.UUID, data ItemTemplateUpdate) (ItemTemplateOut, error) {
	// Verify template belongs to group
	template, err := r.db.ItemTemplate.Query().
		Where(
			itemtemplate.ID(data.ID),
			itemtemplate.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return ItemTemplateOut{}, err
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

	// Update location
	if data.DefaultLocationID != nil {
		updateQ.SetLocationID(*data.DefaultLocationID)
	} else {
		updateQ.ClearLocation()
	}

	// Update default label IDs (stored as JSON)
	if data.DefaultLabelIDs != nil && len(*data.DefaultLabelIDs) > 0 {
		updateQ.SetDefaultLabelIds(*data.DefaultLabelIDs)
	} else {
		updateQ.ClearDefaultLabelIds()
	}

	_, err = updateQ.Save(ctx)
	if err != nil {
		return ItemTemplateOut{}, err
	}

	// Get existing fields
	existingFields, err := r.db.TemplateField.Query().
		Where(templatefield.HasItemTemplateWith(itemtemplate.ID(data.ID))).
		All(ctx)

	if err != nil {
		return ItemTemplateOut{}, err
	}

	// Create a map of existing field IDs for quick lookup
	existingFieldMap := make(map[uuid.UUID]bool)
	for _, field := range existingFields {
		existingFieldMap[field.ID] = true
	}

	// Track which fields are being updated
	updatedFieldIDs := make(map[uuid.UUID]bool)

	// Create or update fields
	for _, field := range data.Fields {
		if field.ID == nil || *field.ID == uuid.Nil {
			// Create new field
			_, err = r.db.TemplateField.Create().
				SetItemTemplateID(data.ID).
				SetType(templatefield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				Save(ctx)

			if err != nil {
				log.Err(err).Msg("failed to create template field")
				return ItemTemplateOut{}, err
			}
		} else {
			// Update existing field
			updatedFieldIDs[*field.ID] = true
			_, err = r.db.TemplateField.Update().
				Where(
					templatefield.ID(*field.ID),
					templatefield.HasItemTemplateWith(itemtemplate.ID(data.ID)),
				).
				SetType(templatefield.Type(field.Type)).
				SetName(field.Name).
				SetTextValue(field.TextValue).
				Save(ctx)

			if err != nil {
				log.Err(err).Msg("failed to update template field")
				return ItemTemplateOut{}, err
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
func (r *ItemTemplatesRepository) Delete(ctx context.Context, gid uuid.UUID, id uuid.UUID) error {
	// Verify template belongs to group
	_, err := r.db.ItemTemplate.Query().
		Where(
			itemtemplate.ID(id),
			itemtemplate.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return err
	}

	// Delete template (fields will be cascade deleted)
	err = r.db.ItemTemplate.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return nil
}
