package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/labeltemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
)

type LabelTemplatesRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	LabelTemplateCreate struct {
		Name         string                 `json:"name"                validate:"required,min=1,max=255"`
		Description  string                 `json:"description"         validate:"max=1000"`
		Width        float64                `json:"width"               validate:"required,gt=0"`
		Height       float64                `json:"height"              validate:"required,gt=0"`
		Preset       *string                `json:"preset,omitempty"    extensions:"x-nullable"`
		IsShared     bool                   `json:"isShared"`
		CanvasData   map[string]interface{} `json:"canvasData"`
		OutputFormat string                 `json:"outputFormat"        validate:"oneof=png pdf"`
		DPI          int                    `json:"dpi"                 validate:"gte=72,lte=1200"`
		MediaType    *string                `json:"mediaType,omitempty" extensions:"x-nullable"`
	}

	LabelTemplateUpdate struct {
		ID           uuid.UUID              `json:"id"`
		Name         string                 `json:"name"                validate:"required,min=1,max=255"`
		Description  string                 `json:"description"         validate:"max=1000"`
		Width        float64                `json:"width"               validate:"required,gt=0"`
		Height       float64                `json:"height"              validate:"required,gt=0"`
		Preset       *string                `json:"preset,omitempty"    extensions:"x-nullable"`
		IsShared     bool                   `json:"isShared"`
		CanvasData   map[string]interface{} `json:"canvasData"`
		OutputFormat string                 `json:"outputFormat"        validate:"oneof=png pdf"`
		DPI          int                    `json:"dpi"                 validate:"gte=72,lte=1200"`
		MediaType    *string                `json:"mediaType,omitempty" extensions:"x-nullable"`
	}

	LabelTemplateSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Width       float64   `json:"width"`
		Height      float64   `json:"height"`
		Preset      *string   `json:"preset,omitempty"`
		IsShared    bool      `json:"isShared"`
		IsOwner     bool      `json:"isOwner"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	LabelTemplateOut struct {
		ID           uuid.UUID              `json:"id"`
		Name         string                 `json:"name"`
		Description  string                 `json:"description"`
		Width        float64                `json:"width"`
		Height       float64                `json:"height"`
		Preset       *string                `json:"preset,omitempty"`
		IsShared     bool                   `json:"isShared"`
		IsOwner      bool                   `json:"isOwner"`
		CanvasData   map[string]interface{} `json:"canvasData"`
		OutputFormat string                 `json:"outputFormat"`
		DPI          int                    `json:"dpi"`
		MediaType    string                 `json:"mediaType"`
		OwnerID      uuid.UUID              `json:"ownerId"`
		CreatedAt    time.Time              `json:"createdAt"`
		UpdatedAt    time.Time              `json:"updatedAt"`
	}
)

func mapLabelTemplateSummary(template *ent.LabelTemplate, currentUserID uuid.UUID) LabelTemplateSummary {
	var preset *string
	if template.Preset != "" {
		preset = &template.Preset
	}

	return LabelTemplateSummary{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		Width:       template.Width,
		Height:      template.Height,
		Preset:      preset,
		IsShared:    template.IsShared,
		IsOwner:     template.OwnerID == currentUserID,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func mapLabelTemplateOut(template *ent.LabelTemplate, currentUserID uuid.UUID) LabelTemplateOut {
	var preset *string
	if template.Preset != "" {
		preset = &template.Preset
	}

	canvasData := template.CanvasData
	if canvasData == nil {
		canvasData = make(map[string]interface{})
	}

	return LabelTemplateOut{
		ID:           template.ID,
		Name:         template.Name,
		Description:  template.Description,
		Width:        template.Width,
		Height:       template.Height,
		Preset:       preset,
		IsShared:     template.IsShared,
		IsOwner:      template.OwnerID == currentUserID,
		CanvasData:   canvasData,
		OutputFormat: template.OutputFormat,
		DPI:          template.Dpi,
		MediaType:    template.MediaType,
		OwnerID:      template.OwnerID,
		CreatedAt:    template.CreatedAt,
		UpdatedAt:    template.UpdatedAt,
	}
}

func (r *LabelTemplatesRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventLabelMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// GetAll returns templates visible to the user (own templates + shared templates in group)
func (r *LabelTemplatesRepository) GetAll(ctx context.Context, gid, uid uuid.UUID) ([]LabelTemplateSummary, error) {
	// Query for templates that are either:
	// 1. Owned by the current user
	// 2. Shared within the group
	templates, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.Or(
				labeltemplate.OwnerID(uid),
				labeltemplate.IsShared(true),
			),
		).
		Order(ent.Asc(labeltemplate.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]LabelTemplateSummary, len(templates))
	for i, template := range templates {
		result[i] = mapLabelTemplateSummary(template, uid)
	}

	return result, nil
}

// GetOne returns a single template by ID if the user has access
func (r *LabelTemplatesRepository) GetOne(ctx context.Context, gid, uid, id uuid.UUID) (LabelTemplateOut, error) {
	template, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.ID(id),
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.Or(
				labeltemplate.OwnerID(uid),
				labeltemplate.IsShared(true),
			),
		).
		Only(ctx)

	if err != nil {
		return LabelTemplateOut{}, err
	}

	return mapLabelTemplateOut(template, uid), nil
}

// Create creates a new template owned by the user
func (r *LabelTemplatesRepository) Create(ctx context.Context, gid, uid uuid.UUID, data LabelTemplateCreate) (LabelTemplateOut, error) {
	q := r.db.LabelTemplate.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetWidth(data.Width).
		SetHeight(data.Height).
		SetIsShared(data.IsShared).
		SetOutputFormat(data.OutputFormat).
		SetDpi(data.DPI).
		SetOwnerID(uid).
		SetGroupID(gid)

	if data.Preset != nil && *data.Preset != "" {
		q.SetPreset(*data.Preset)
	}

	if data.CanvasData != nil {
		q.SetCanvasData(data.CanvasData)
	}

	if data.MediaType != nil && *data.MediaType != "" {
		q.SetMediaType(*data.MediaType)
	}

	template, err := q.Save(ctx)
	if err != nil {
		return LabelTemplateOut{}, err
	}

	r.publishMutationEvent(gid)
	return mapLabelTemplateOut(template, uid), nil
}

// Update updates an existing template (only owner can update)
func (r *LabelTemplatesRepository) Update(ctx context.Context, gid, uid uuid.UUID, data LabelTemplateUpdate) (LabelTemplateOut, error) {
	// Verify template belongs to group and user is owner
	template, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.ID(data.ID),
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.OwnerID(uid),
		).
		Only(ctx)

	if err != nil {
		return LabelTemplateOut{}, err
	}

	// Update template
	updateQ := template.Update().
		SetName(data.Name).
		SetDescription(data.Description).
		SetWidth(data.Width).
		SetHeight(data.Height).
		SetIsShared(data.IsShared).
		SetOutputFormat(data.OutputFormat).
		SetDpi(data.DPI)

	if data.Preset != nil && *data.Preset != "" {
		updateQ.SetPreset(*data.Preset)
	} else {
		updateQ.ClearPreset()
	}

	if data.CanvasData != nil {
		updateQ.SetCanvasData(data.CanvasData)
	}

	if data.MediaType != nil && *data.MediaType != "" {
		updateQ.SetMediaType(*data.MediaType)
	} else {
		updateQ.ClearMediaType()
	}

	_, err = updateQ.Save(ctx)
	if err != nil {
		return LabelTemplateOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, uid, template.ID)
}

// Delete deletes a template (only owner can delete)
func (r *LabelTemplatesRepository) Delete(ctx context.Context, gid, uid, id uuid.UUID) error {
	// Verify template belongs to group and user is owner
	_, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.ID(id),
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.OwnerID(uid),
		).
		Only(ctx)

	if err != nil {
		return err
	}

	err = r.db.LabelTemplate.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return nil
}

// Duplicate creates a copy of a template for the current user
func (r *LabelTemplatesRepository) Duplicate(ctx context.Context, gid, uid, id uuid.UUID) (LabelTemplateOut, error) {
	// Get the source template (must be accessible to user)
	source, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.ID(id),
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.Or(
				labeltemplate.OwnerID(uid),
				labeltemplate.IsShared(true),
			),
		).
		Only(ctx)

	if err != nil {
		return LabelTemplateOut{}, err
	}

	// Create a copy with new ID, owned by current user
	// Truncate name if needed to fit " (Copy)" suffix within 255 char limit
	const maxNameLen = 255
	const suffix = " (Copy)"
	newName := source.Name
	if len(newName)+len(suffix) > maxNameLen {
		newName = newName[:maxNameLen-len(suffix)]
	}
	newName += suffix

	q := r.db.LabelTemplate.Create().
		SetName(newName).
		SetDescription(source.Description).
		SetWidth(source.Width).
		SetHeight(source.Height).
		SetIsShared(false). // Copies are private by default
		SetOutputFormat(source.OutputFormat).
		SetDpi(source.Dpi).
		SetOwnerID(uid).
		SetGroupID(gid)

	if source.Preset != "" {
		q.SetPreset(source.Preset)
	}

	if source.CanvasData != nil {
		q.SetCanvasData(source.CanvasData)
	}

	newTemplate, err := q.Save(ctx)
	if err != nil {
		return LabelTemplateOut{}, err
	}

	r.publishMutationEvent(gid)
	return mapLabelTemplateOut(newTemplate, uid), nil
}

// GetByOwner returns all templates owned by a specific user
func (r *LabelTemplatesRepository) GetByOwner(ctx context.Context, gid, uid uuid.UUID) ([]LabelTemplateSummary, error) {
	templates, err := r.db.LabelTemplate.Query().
		Where(
			labeltemplate.HasGroupWith(group.ID(gid)),
			labeltemplate.HasOwnerWith(user.ID(uid)),
		).
		Order(ent.Asc(labeltemplate.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]LabelTemplateSummary, len(templates))
	for i, template := range templates {
		result[i] = mapLabelTemplateSummary(template, uid)
	}

	return result, nil
}
