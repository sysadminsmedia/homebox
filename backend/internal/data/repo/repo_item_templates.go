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
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/templatefield"
)

type ItemTemplatesRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	TemplateField struct {
		ID        uuid.UUID `json:"id,omitempty"`
		Type      string    `json:"type"`
		Name      string    `json:"name"`
		TextValue string    `json:"textValue"`
	}

	ItemTemplateCreate struct {
		Name        string `json:"name"        validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"max=1000"`
		Notes       string `json:"notes"       validate:"max=1000"`

		// Default values for items
		DefaultQuantity         int    `json:"defaultQuantity"`
		DefaultInsured          bool   `json:"defaultInsured"`
		DefaultManufacturer     string `json:"defaultManufacturer"     validate:"max=255"`
		DefaultLifetimeWarranty bool   `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  string `json:"defaultWarrantyDetails"  validate:"max=1000"`

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
		DefaultQuantity         int    `json:"defaultQuantity"`
		DefaultInsured          bool   `json:"defaultInsured"`
		DefaultManufacturer     string `json:"defaultManufacturer"     validate:"max=255"`
		DefaultLifetimeWarranty bool   `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  string `json:"defaultWarrantyDetails"  validate:"max=1000"`

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
		DefaultManufacturer     string `json:"defaultManufacturer"`
		DefaultLifetimeWarranty bool   `json:"defaultLifetimeWarranty"`
		DefaultWarrantyDetails  string `json:"defaultWarrantyDetails"`

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
		ID:        field.ID,
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

func mapTemplateOut(template *ent.ItemTemplate) ItemTemplateOut {
	fields := make([]TemplateField, 0)
	if template.Edges.Fields != nil {
		fields = mapTemplateFieldSlice(template.Edges.Fields)
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
		DefaultManufacturer:     template.DefaultManufacturer,
		DefaultLifetimeWarranty: template.DefaultLifetimeWarranty,
		DefaultWarrantyDetails:  template.DefaultWarrantyDetails,
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

// GetOne returns a single template by ID
func (r *ItemTemplatesRepository) GetOne(ctx context.Context, id uuid.UUID) (ItemTemplateOut, error) {
	template, err := r.db.ItemTemplate.Query().
		Where(itemtemplate.ID(id)).
		WithFields().
		Only(ctx)

	if err != nil {
		return ItemTemplateOut{}, err
	}

	return mapTemplateOut(template), nil
}

// Create creates a new template
func (r *ItemTemplatesRepository) Create(ctx context.Context, gid uuid.UUID, data ItemTemplateCreate) (ItemTemplateOut, error) {
	q := r.db.ItemTemplate.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetNotes(data.Notes).
		SetDefaultQuantity(data.DefaultQuantity).
		SetDefaultInsured(data.DefaultInsured).
		SetDefaultManufacturer(data.DefaultManufacturer).
		SetDefaultLifetimeWarranty(data.DefaultLifetimeWarranty).
		SetDefaultWarrantyDetails(data.DefaultWarrantyDetails).
		SetIncludeWarrantyFields(data.IncludeWarrantyFields).
		SetIncludePurchaseFields(data.IncludePurchaseFields).
		SetIncludeSoldFields(data.IncludeSoldFields).
		SetGroupID(gid)

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
	return r.GetOne(ctx, template.ID)
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
	_, err = template.Update().
		SetName(data.Name).
		SetDescription(data.Description).
		SetNotes(data.Notes).
		SetDefaultQuantity(data.DefaultQuantity).
		SetDefaultInsured(data.DefaultInsured).
		SetDefaultManufacturer(data.DefaultManufacturer).
		SetDefaultLifetimeWarranty(data.DefaultLifetimeWarranty).
		SetDefaultWarrantyDetails(data.DefaultWarrantyDetails).
		SetIncludeWarrantyFields(data.IncludeWarrantyFields).
		SetIncludePurchaseFields(data.IncludePurchaseFields).
		SetIncludeSoldFields(data.IncludeSoldFields).
		Save(ctx)

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
		if field.ID == uuid.Nil {
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
			updatedFieldIDs[field.ID] = true
			_, err = r.db.TemplateField.Update().
				Where(
					templatefield.ID(field.ID),
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
	return r.GetOne(ctx, template.ID)
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
