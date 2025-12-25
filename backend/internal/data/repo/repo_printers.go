package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/printer"
)

type PrintersRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	PrinterCreate struct {
		Name          string   `json:"name"                    validate:"required,min=1,max=255"`
		Description   string   `json:"description"             validate:"max=1000"`
		PrinterType   string   `json:"printerType"             validate:"required,oneof=ipp cups brother_raster"`
		Address       string   `json:"address"                 validate:"required,min=1,max=512"`
		IsDefault     bool     `json:"isDefault"`
		LabelWidthMM  *float64 `json:"labelWidthMm,omitempty"  extensions:"x-nullable"`
		LabelHeightMM *float64 `json:"labelHeightMm,omitempty" extensions:"x-nullable"`
		DPI           int      `json:"dpi"                     validate:"gte=72,lte=1200"`
		MediaType     *string  `json:"mediaType,omitempty"     extensions:"x-nullable"`
	}

	PrinterUpdate struct {
		ID            uuid.UUID `json:"id"`
		Name          string    `json:"name"                    validate:"required,min=1,max=255"`
		Description   string    `json:"description"             validate:"max=1000"`
		PrinterType   string    `json:"printerType"             validate:"required,oneof=ipp cups brother_raster"`
		Address       string    `json:"address"                 validate:"required,min=1,max=512"`
		IsDefault     bool      `json:"isDefault"`
		LabelWidthMM  *float64  `json:"labelWidthMm,omitempty"  extensions:"x-nullable"`
		LabelHeightMM *float64  `json:"labelHeightMm,omitempty" extensions:"x-nullable"`
		DPI           int       `json:"dpi"                     validate:"gte=72,lte=1200"`
		MediaType     *string   `json:"mediaType,omitempty"     extensions:"x-nullable"`
	}

	PrinterSummary struct {
		ID            uuid.UUID `json:"id"`
		Name          string    `json:"name"`
		Description   string    `json:"description"`
		PrinterType   string    `json:"printerType"`
		Address       string    `json:"address"`
		IsDefault     bool      `json:"isDefault"`
		LabelWidthMM  *float64  `json:"labelWidthMm,omitempty"`
		LabelHeightMM *float64  `json:"labelHeightMm,omitempty"`
		DPI           int       `json:"dpi"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"createdAt"`
		UpdatedAt     time.Time `json:"updatedAt"`
	}

	PrinterOut struct {
		ID              uuid.UUID  `json:"id"`
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		PrinterType     string     `json:"printerType"`
		Address         string     `json:"address"`
		IsDefault       bool       `json:"isDefault"`
		LabelWidthMM    *float64   `json:"labelWidthMm,omitempty"`
		LabelHeightMM   *float64   `json:"labelHeightMm,omitempty"`
		DPI             int        `json:"dpi"`
		MediaType       *string    `json:"mediaType,omitempty"`
		Status          string     `json:"status"`
		LastStatusCheck *time.Time `json:"lastStatusCheck,omitempty"`
		CreatedAt       time.Time  `json:"createdAt"`
		UpdatedAt       time.Time  `json:"updatedAt"`
	}
)

func mapPrinterSummary(p *ent.Printer) PrinterSummary {
	var labelWidth, labelHeight *float64
	if p.LabelWidthMm != 0 {
		labelWidth = &p.LabelWidthMm
	}
	if p.LabelHeightMm != 0 {
		labelHeight = &p.LabelHeightMm
	}

	return PrinterSummary{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		PrinterType:   string(p.PrinterType),
		Address:       p.Address,
		IsDefault:     p.IsDefault,
		LabelWidthMM:  labelWidth,
		LabelHeightMM: labelHeight,
		DPI:           p.Dpi,
		Status:        string(p.Status),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func mapPrinterOut(p *ent.Printer) PrinterOut {
	var labelWidth, labelHeight *float64
	if p.LabelWidthMm != 0 {
		labelWidth = &p.LabelWidthMm
	}
	if p.LabelHeightMm != 0 {
		labelHeight = &p.LabelHeightMm
	}

	var mediaType *string
	if p.MediaType != "" {
		mediaType = &p.MediaType
	}

	return PrinterOut{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		PrinterType:     string(p.PrinterType),
		Address:         p.Address,
		IsDefault:       p.IsDefault,
		LabelWidthMM:    labelWidth,
		LabelHeightMM:   labelHeight,
		DPI:             p.Dpi,
		MediaType:       mediaType,
		Status:          string(p.Status),
		LastStatusCheck: p.LastStatusCheck,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func (r *PrintersRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventLabelMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// GetAll returns all printers in the group
func (r *PrintersRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]PrinterSummary, error) {
	printers, err := r.db.Printer.Query().
		Where(printer.HasGroupWith(group.ID(gid))).
		Order(ent.Asc(printer.FieldName)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]PrinterSummary, len(printers))
	for i, p := range printers {
		result[i] = mapPrinterSummary(p)
	}

	return result, nil
}

// GetOne returns a single printer by ID
func (r *PrintersRepository) GetOne(ctx context.Context, gid, id uuid.UUID) (PrinterOut, error) {
	p, err := r.db.Printer.Query().
		Where(
			printer.ID(id),
			printer.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return PrinterOut{}, err
	}

	return mapPrinterOut(p), nil
}

// GetDefault returns the default printer for the group, if one exists
func (r *PrintersRepository) GetDefault(ctx context.Context, gid uuid.UUID) (PrinterOut, error) {
	p, err := r.db.Printer.Query().
		Where(
			printer.HasGroupWith(group.ID(gid)),
			printer.IsDefault(true),
		).
		Only(ctx)

	if err != nil {
		return PrinterOut{}, err
	}

	return mapPrinterOut(p), nil
}

// Create creates a new printer
func (r *PrintersRepository) Create(ctx context.Context, gid uuid.UUID, data PrinterCreate) (PrinterOut, error) {
	// If this printer is set as default, clear any existing default
	if data.IsDefault {
		err := r.clearDefault(ctx, gid)
		if err != nil {
			return PrinterOut{}, err
		}
	}

	q := r.db.Printer.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetPrinterType(printer.PrinterType(data.PrinterType)).
		SetAddress(data.Address).
		SetIsDefault(data.IsDefault).
		SetDpi(data.DPI).
		SetGroupID(gid)

	if data.LabelWidthMM != nil {
		q.SetLabelWidthMm(*data.LabelWidthMM)
	}
	if data.LabelHeightMM != nil {
		q.SetLabelHeightMm(*data.LabelHeightMM)
	}
	if data.MediaType != nil {
		q.SetMediaType(*data.MediaType)
	}

	p, err := q.Save(ctx)
	if err != nil {
		return PrinterOut{}, err
	}

	r.publishMutationEvent(gid)
	return mapPrinterOut(p), nil
}

// Update updates an existing printer
func (r *PrintersRepository) Update(ctx context.Context, gid uuid.UUID, data PrinterUpdate) (PrinterOut, error) {
	// Verify printer belongs to group
	p, err := r.db.Printer.Query().
		Where(
			printer.ID(data.ID),
			printer.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return PrinterOut{}, err
	}

	// If this printer is being set as default, clear any existing default
	if data.IsDefault && !p.IsDefault {
		err := r.clearDefault(ctx, gid)
		if err != nil {
			return PrinterOut{}, err
		}
	}

	updateQ := p.Update().
		SetName(data.Name).
		SetDescription(data.Description).
		SetPrinterType(printer.PrinterType(data.PrinterType)).
		SetAddress(data.Address).
		SetIsDefault(data.IsDefault).
		SetDpi(data.DPI)

	if data.LabelWidthMM != nil {
		updateQ.SetLabelWidthMm(*data.LabelWidthMM)
	} else {
		updateQ.ClearLabelWidthMm()
	}
	if data.LabelHeightMM != nil {
		updateQ.SetLabelHeightMm(*data.LabelHeightMM)
	} else {
		updateQ.ClearLabelHeightMm()
	}
	if data.MediaType != nil {
		updateQ.SetMediaType(*data.MediaType)
	} else {
		updateQ.ClearMediaType()
	}

	_, err = updateQ.Save(ctx)
	if err != nil {
		return PrinterOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, p.ID)
}

// Delete deletes a printer
func (r *PrintersRepository) Delete(ctx context.Context, gid, id uuid.UUID) error {
	// Verify printer belongs to group
	_, err := r.db.Printer.Query().
		Where(
			printer.ID(id),
			printer.HasGroupWith(group.ID(gid)),
		).
		Only(ctx)

	if err != nil {
		return err
	}

	err = r.db.Printer.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return nil
}

// SetDefault sets a printer as the default for the group
func (r *PrintersRepository) SetDefault(ctx context.Context, gid, id uuid.UUID) error {
	// Verify printer exists and belongs to group before clearing defaults
	exists, err := r.db.Printer.Query().
		Where(
			printer.ID(id),
			printer.HasGroupWith(group.ID(gid)),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("printer not found")
	}

	// Clear existing default
	err = r.clearDefault(ctx, gid)
	if err != nil {
		return err
	}

	// Set new default
	_, err = r.db.Printer.UpdateOneID(id).
		Where(printer.HasGroupWith(group.ID(gid))).
		SetIsDefault(true).
		Save(ctx)

	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)
	return nil
}

// UpdateStatus updates the printer status
func (r *PrintersRepository) UpdateStatus(ctx context.Context, gid, id uuid.UUID, status string) error {
	now := time.Now()
	_, err := r.db.Printer.UpdateOneID(id).
		Where(printer.HasGroupWith(group.ID(gid))).
		SetStatus(printer.Status(status)).
		SetLastStatusCheck(now).
		Save(ctx)

	return err
}

// clearDefault clears the default flag from all printers in the group
func (r *PrintersRepository) clearDefault(ctx context.Context, gid uuid.UUID) error {
	_, err := r.db.Printer.Update().
		Where(
			printer.HasGroupWith(group.ID(gid)),
			printer.IsDefault(true),
		).
		SetIsDefault(false).
		Save(ctx)

	return err
}
