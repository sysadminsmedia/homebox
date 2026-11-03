package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"github.com/sysadminsmedia/homebox/backend/pkgs/labelmaker"
)

// mimePNG is the content type used for rendered label images.
const mimePNG = "image/png"

// HandleLabelTemplatesGetAll godoc
//
//	@Summary	Get All Label Templates
//	@Tags		Label Templates
//	@Produce	json
//	@Success	200	{object}	[]repo.LabelTemplateSummary
//	@Router		/v1/label-templates [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.LabelTemplateSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.LabelTemplates.GetAll(r.Context(), auth.GID, auth.UID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleLabelTemplatesGet godoc
//
//	@Summary	Get Label Template
//	@Tags		Label Templates
//	@Produce	json
//	@Param		id	path		string	true	"Template ID"
//	@Success	200	{object}	repo.LabelTemplateOut
//	@Router		/v1/label-templates/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.LabelTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleLabelTemplatesCreate godoc
//
//	@Summary	Create Label Template
//	@Tags		Label Templates
//	@Produce	json
//	@Param		payload	body		repo.LabelTemplateCreate	true	"Template Data"
//	@Success	201		{object}	repo.LabelTemplateOut
//	@Router		/v1/label-templates [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.LabelTemplateCreate) (repo.LabelTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.LabelTemplates.Create(r.Context(), auth.GID, auth.UID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleLabelTemplatesUpdate godoc
//
//	@Summary	Update Label Template
//	@Tags		Label Templates
//	@Produce	json
//	@Param		id		path		string						true	"Template ID"
//	@Param		payload	body		repo.LabelTemplateUpdate	true	"Template Data"
//	@Success	200		{object}	repo.LabelTemplateOut
//	@Router		/v1/label-templates/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.LabelTemplateUpdate) (repo.LabelTemplateOut, error) {
		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.LabelTemplates.Update(r.Context(), auth.GID, auth.UID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleLabelTemplatesDelete godoc
//
//	@Summary	Delete Label Template
//	@Tags		Label Templates
//	@Produce	json
//	@Param		id	path	string	true	"Template ID"
//	@Success	204
//	@Router		/v1/label-templates/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.LabelTemplates.Delete(r.Context(), auth.GID, auth.UID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleLabelTemplatesDuplicate godoc
//
//	@Summary	Duplicate Label Template
//	@Tags		Label Templates
//	@Produce	json
//	@Param		id	path		string	true	"Template ID"
//	@Success	201	{object}	repo.LabelTemplateOut
//	@Router		/v1/label-templates/{id}/duplicate [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesDuplicate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.LabelTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.LabelTemplates.Duplicate(r.Context(), auth.GID, auth.UID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusCreated)
}

// HandleLabelTemplatesPresets godoc
//
//	@Summary	Get Label Size Presets
//	@Tags		Label Templates
//	@Produce	json
//	@Success	200	{object}	[]labelmaker.LabelPreset
//	@Router		/v1/label-templates/presets [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesPresets() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]labelmaker.LabelPreset, error) {
		return labelmaker.LabelPresets, nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleLabelTemplatesBarcodeFormats godoc
//
//	@Summary	Get Supported Barcode Formats
//	@Tags		Label Templates
//	@Produce	json
//	@Success	200	{object}	[]labelmaker.BarcodeFormatInfo
//	@Router		/v1/label-templates/barcode-formats [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesBarcodeFormats() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]labelmaker.BarcodeFormatInfo, error) {
		return labelmaker.GetBarcodeFormatInfo(), nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleLabelTemplatesPreview godoc
//
//	@Summary	Preview Label Template
//	@Tags		Label Templates
//	@Produce	image/png
//	@Param		id	path	string	true	"Template ID"
//	@Success	200	{file}	binary
//	@Router		/v1/label-templates/{id}/preview [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesPreview() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: template.CanvasData,
		}

		pngData, err := renderer.RenderPreview(templateData)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mimePNG)
		_, err = w.Write(pngData)
		return err
	}
}

// LabelTemplateRenderRequest represents a request to render labels
type LabelTemplateRenderRequest struct {
	ItemIDs       []uuid.UUID `json:"itemIds"              validate:"required,min=1"`
	Format        string      `json:"format"`               // "png" or "pdf", defaults to "png"
	PageSize      string      `json:"pageSize"`             // "Letter", "A4", or "Custom" for PDF
	ShowCutGuides bool        `json:"showCutGuides"`        // Draw light borders around labels for cutting
	SkipLabels    int         `json:"skipLabels"`           // Leave this many positions empty on the first sheet
	CanvasData    string      `json:"canvasData,omitempty"` // Optional: canvas data for live preview (overrides saved template)
}

// HandleLabelTemplatesRender godoc
//
//	@Summary	Render Label for Items
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	image/png,application/pdf
//	@Param		id		path	string						true	"Template ID"
//	@Param		payload	body	LabelTemplateRenderRequest	true	"Items to render"
//	@Success	200		{file}	binary
//	@Router		/v1/label-templates/{id}/render [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesRender() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		req, err := adapters.DecodeBody[LabelTemplateRenderRequest](r)
		if err != nil {
			return err
		}

		if len(req.ItemIDs) == 0 {
			return fmt.Errorf("at least one item ID is required")
		}
		if len(req.ItemIDs) > maxPrintLabelsPerRequest {
			return validate.NewRequestError(fmt.Errorf("too many items in one render request (max %d)", maxPrintLabelsPerRequest), http.StatusBadRequest)
		}

		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		// Use provided canvas data for live preview, or fall back to saved template
		canvasData := template.CanvasData
		if req.CanvasData != "" {
			var parsedCanvasData map[string]interface{}
			if err := json.Unmarshal([]byte(req.CanvasData), &parsedCanvasData); err == nil {
				canvasData = parsedCanvasData
			}
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: canvasData,
		}

		baseURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)
		sheetLayout := sheetLayoutFor(template, templateData)

		// Handle PDF format
		format := req.Format
		if format == "" {
			format = "png"
		}

		if format == "pdf" {
			entities := ctrl.fetchEntitiesData(r.Context(), auth.GID, req.ItemIDs, baseURL)
			if len(entities) == 0 {
				return fmt.Errorf("no valid items found")
			}
			pdfOpts := pdfOptionsFor(template, req.PageSize, req.ShowCutGuides, req.SkipLabels)
			return renderLabelsPDF(w, renderer, templateData, entities, pdfOpts, sheetLayout, fmt.Sprintf("labels-%d.pdf", len(entities)))
		}

		// Handle PNG sheet format (multiple items with sheet layout)
		if sheetLayout != nil && len(req.ItemIDs) > 1 {
			entities := ctrl.fetchEntitiesData(r.Context(), auth.GID, req.ItemIDs, baseURL)
			if len(entities) == 0 {
				return fmt.Errorf("no valid items found")
			}
			return renderLabelsPNGSheet(w, renderer, templateData, entities, sheetLayout)
		}

		// Default: render single entity as PNG
		entity, err := ctrl.buildEntityData(r.Context(), auth.GID, req.ItemIDs[0], baseURL)
		if err != nil {
			return err
		}

		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Entity:   entity,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mimePNG)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=label-%s.png", entity.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// buildEntityData fetches an entity and converts it to labelmaker.EntityData.
// Items and locations are both entities and expose the same label fields; a
// location describes itself, an item describes the location it resolves to.
func (ctrl *V1Controller) buildEntityData(ctx context.Context, gid, id uuid.UUID, baseURL string) (*labelmaker.EntityData, error) {
	e, err := ctrl.repo.Entities.GetOneByGroup(ctx, gid, id)
	if err != nil {
		return nil, err
	}

	isLocation := e.EntityType != nil && e.EntityType.IsLocation

	var locationID uuid.UUID
	var locationName string
	switch {
	case isLocation:
		locationID, locationName = e.ID, e.Name
	case e.Location != nil:
		locationID, locationName = e.Location.ID, e.Location.Name
	case e.Parent != nil:
		locationID, locationName = e.Parent.ID, e.Parent.Name
	}

	locationPath := []string{}
	if locationID != uuid.Nil {
		path, err := ctrl.repo.Entities.PathForEntity(ctx, gid, locationID)
		if err != nil {
			log.Warn().Err(err).Str("entityID", id.String()).Msg("failed to resolve location path for label rendering")
		}
		for _, p := range path {
			locationPath = append(locationPath, p.Name)
		}
	}

	itemCount, err := ctrl.repo.Entities.GetItemCountByGroup(ctx, gid, e.ID)
	if err != nil {
		log.Warn().Err(err).Str("entityID", id.String()).Msg("failed to count items for label rendering")
	}

	tags := make([]string, len(e.Tags))
	for i, t := range e.Tags {
		tags[i] = t.Name
	}

	customFields := make(map[string]string, len(e.Fields))
	for _, f := range e.Fields {
		customFields[f.Name] = f.TextValue
	}

	return &labelmaker.EntityData{
		ID:           e.ID,
		Name:         e.Name,
		Description:  e.Description,
		AssetID:      e.AssetID.String(),
		SerialNumber: e.SerialNumber,
		ModelNumber:  e.ModelNumber,
		Manufacturer: e.Manufacturer,
		IsLocation:   isLocation,
		LocationName: locationName,
		LocationPath: locationPath,
		Tags:         tags,
		CustomFields: customFields,
		ItemCount:    itemCount,
		Quantity:     e.Quantity,
		Notes:        e.Notes,
		ItemURL:      fmt.Sprintf("%s/item/%s", baseURL, e.ID.String()),
		LocationURL:  fmt.Sprintf("%s/location/%s", baseURL, e.ID.String()),
		AssetURL:     assetURL(baseURL, e.AssetID),
	}, nil
}

// assetURL returns the stable /a/<asset id> link, which keeps resolving after
// a tag is printed even if the entity it points at changes. Empty without an asset ID.
func assetURL(baseURL string, aid repo.AssetID) string {
	if aid.Nil() {
		return ""
	}
	return fmt.Sprintf("%s/a/%s", baseURL, aid.String())
}

// fetchEntitiesData fetches entities by ID and converts them to labelmaker.EntityData,
// skipping any that can't be loaded
func (ctrl *V1Controller) fetchEntitiesData(ctx context.Context, gid uuid.UUID, ids []uuid.UUID, baseURL string) []*labelmaker.EntityData {
	entities := make([]*labelmaker.EntityData, 0, len(ids))
	for _, id := range ids {
		entity, err := ctrl.buildEntityData(ctx, gid, id, baseURL)
		if err != nil {
			log.Warn().Err(err).Str("entityID", id.String()).Msg("failed to fetch entity for label rendering")
			continue
		}
		entities = append(entities, entity)
	}
	return entities
}

// sheetLayoutFor returns the page layout for printing several labels per
// sheet: the template's own layout if it has one, else its preset's. The
// template's print offset is folded into the margins. Returns nil when the
// template isn't laid out on sheets. Sheet presets also fix the label size.
func sheetLayoutFor(template repo.LabelTemplateOut, templateData *labelmaker.TemplateData) *labelmaker.SheetLayout {
	var base *labelmaker.SheetLayout
	if template.Preset != nil && *template.Preset != "" {
		if preset := labelmaker.GetPresetByKey(*template.Preset); preset != nil && preset.SheetLayout != nil {
			// Use preset's exact dimensions for proper sheet alignment
			templateData.Width = preset.Width
			templateData.Height = preset.Height
			base = preset.SheetLayout
		}
	}
	if template.SheetLayout != nil {
		base = template.SheetLayout
	}
	if base == nil {
		return nil
	}

	// Copy so the shared preset isn't modified
	layout := *base
	layout.MarginLeft += template.OffsetX
	layout.MarginTop += template.OffsetY
	return &layout
}

// pdfOptionsFor returns the PDF options for a render request. Templates
// without a sheet layout use the plain grid, shifted by the print offset.
func pdfOptionsFor(template repo.LabelTemplateOut, pageSize string, showCutGuides bool, skipLabels int) labelmaker.PDFRenderOptions {
	pdfOpts := labelmaker.DefaultPDFOptions()
	if pageSize == "A4" {
		pdfOpts.PageSize = labelmaker.PDFPageA4
	} else {
		pdfOpts.PageSize = labelmaker.PDFPageLetter
	}
	pdfOpts.ShowCutGuides = showCutGuides
	pdfOpts.SkipLabels = skipLabels
	pdfOpts.MarginLeft += template.OffsetX
	pdfOpts.MarginTop += template.OffsetY
	return pdfOpts
}

// renderLabelsPDF renders labels to PDF format
func renderLabelsPDF(w http.ResponseWriter, renderer *labelmaker.TemplateRenderer, templateData *labelmaker.TemplateData, entities []*labelmaker.EntityData, pdfOpts labelmaker.PDFRenderOptions, sheetLayout *labelmaker.SheetLayout, filename string) error {
	pdfData, err := renderer.RenderToPDFWithSheetLayout(templateData, entities, pdfOpts, sheetLayout)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, err = w.Write(pdfData)
	return err
}

// renderLabelsPNGSheet renders multiple labels to a PNG sheet
func renderLabelsPNGSheet(w http.ResponseWriter, renderer *labelmaker.TemplateRenderer, templateData *labelmaker.TemplateData, items []*labelmaker.EntityData, sheetLayout *labelmaker.SheetLayout) error {
	sheets, err := renderer.RenderToSheets(templateData, items, sheetLayout)
	if err != nil {
		return err
	}

	if len(sheets) == 0 {
		return fmt.Errorf("no sheets generated")
	}

	// Return the first sheet (could be extended to return ZIP for multiple sheets)
	w.Header().Set("Content-Type", mimePNG)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=labels-sheet-%d.png", len(items)))
	if len(sheets) > 1 {
		w.Header().Set("X-Labels-Warning", fmt.Sprintf("Only returning first sheet of %d. Use PDF format for all sheets.", len(sheets)))
	}
	_, err = w.Write(sheets[0])
	return err
}

// printTemplateLabel sends a rendered label PNG to the admin-configured print
// command. Width and Height are the rendered image size in pixels at the
// template DPI, WidthMm and HeightMm come from the template itself.
func printTemplateLabel(ctx context.Context, cfg *config.Config, tmpl repo.LabelTemplateOut, pngData []byte, title, labelURL string, copies int) error {
	dpi := tmpl.DPI
	if dpi <= 0 {
		dpi = 300
	}
	data := map[string]string{
		"TitleText": title,
		"URL":       labelURL,
		"Dpi":       strconv.Itoa(dpi),
		"WidthMm":   strconv.FormatFloat(tmpl.Width, 'f', -1, 64),
		"HeightMm":  strconv.FormatFloat(tmpl.Height, 'f', -1, 64),
	}
	if cfgImg, err := png.DecodeConfig(bytes.NewReader(pngData)); err == nil {
		data["Width"] = strconv.Itoa(cfgImg.Width)
		data["Height"] = strconv.Itoa(cfgImg.Height)
	}
	return labelmaker.PrintPNG(ctx, cfg, pngData, data, copies)
}

// LabelPrintItem represents an item to print with its quantity
type LabelPrintItem struct {
	ID       uuid.UUID `json:"id"       validate:"required"`
	Quantity int       `json:"quantity"` // Number of copies for this item
}

// Limits for a single print request. Each copy is a separate render
// and print job, so unbounded values would tie up the server and the printer.
const (
	maxPrintCopies           = 100
	maxPrintLabelsPerRequest = 500
	maxPrintTotalLabels      = 500 // copies summed over every entity in the request
)

// totalLabelCount sums the copies a request will print, applying the same
// clamping as the print loop, so a request can be rejected before any work.
func totalLabelCount(defaultCopies int, quantities []int) int {
	total := 0
	for _, q := range quantities {
		if q > 0 {
			total += clampCopies(q)
		} else {
			total += defaultCopies
		}
	}
	return total
}

func clampCopies(n int) int {
	if n <= 0 {
		return 1
	}
	if n > maxPrintCopies {
		return maxPrintCopies
	}
	return n
}

// LabelTemplatePrintRequest represents a request to print labels with the configured print command
type LabelTemplatePrintRequest struct {
	ItemIDs []uuid.UUID      `json:"itemIds,omitempty"` // Simple list (1 copy each) - for backward compatibility
	Items   []LabelPrintItem `json:"items,omitempty"`   // Items with individual quantities
	Copies  int              `json:"copies"`            // Default copies per label (used if item.quantity is 0)
}

// LabelTemplatePrintResponse represents the result of a direct print operation
type LabelTemplatePrintResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	LabelCount int    `json:"labelCount"`
}

// HandleLabelTemplatesPrint godoc
//
//	@Summary	Print Labels
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"Template ID"
//	@Param		payload	body		LabelTemplatePrintRequest	true	"Print request"
//	@Success	200		{object}	LabelTemplatePrintResponse
//	@Router		/v1/label-templates/{id}/print [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesPrint() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		// Parse request body
		req, err := adapters.DecodeBody[LabelTemplatePrintRequest](r)
		if err != nil {
			return err
		}

		// Build items list from either Items (with quantities) or ItemIDs (backward compatible)
		var itemsToPrint []LabelPrintItem
		switch {
		case len(req.Items) > 0:
			itemsToPrint = req.Items
		case len(req.ItemIDs) > 0:
			// Convert ItemIDs to Items with default quantity
			for _, id := range req.ItemIDs {
				itemsToPrint = append(itemsToPrint, LabelPrintItem{ID: id, Quantity: 0})
			}
		default:
			return fmt.Errorf("at least one item is required (provide itemIds or items)")
		}
		if len(itemsToPrint) > maxPrintLabelsPerRequest {
			return validate.NewRequestError(fmt.Errorf("too many items in one print request (max %d)", maxPrintLabelsPerRequest), http.StatusBadRequest)
		}
		itemQuantities := make([]int, len(itemsToPrint))
		for i, it := range itemsToPrint {
			itemQuantities[i] = it.Quantity
		}
		if total := totalLabelCount(clampCopies(req.Copies), itemQuantities); total > maxPrintTotalLabels {
			return validate.NewRequestError(fmt.Errorf("too many labels in one print request (%d, max %d)", total, maxPrintTotalLabels), http.StatusBadRequest)
		}

		if cmd := ctrl.config.LabelMaker.PrintCommand; cmd == nil || *cmd == "" {
			return validate.NewRequestError(errors.New("server-side label printing is not configured; set HBOX_LABEL_MAKER_PRINT_COMMAND"), http.StatusBadRequest)
		}

		// Get the template
		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		// Create renderer
		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: template.CanvasData,
		}

		// Default copies (used when item quantity is 0 or not specified)
		defaultCopies := clampCopies(req.Copies)

		labelsPrinted := 0
		labelsSkipped := 0

		for _, printItem := range itemsToPrint {
			itemData, err := ctrl.buildEntityData(r.Context(), auth.GID, printItem.ID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
			if err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Msg("failed to fetch item for printing")
				labelsSkipped++
				continue
			}

			// Render label as PNG
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Entity:   itemData,
				Template: templateData,
			})
			if err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Str("itemName", itemData.Name).Msg("failed to render label for printing")
				labelsSkipped++
				continue
			}

			// Determine copies for this item (use per-item quantity if set, otherwise default)
			copies := defaultCopies
			if printItem.Quantity > 0 {
				copies = clampCopies(printItem.Quantity)
			}

			if err := printTemplateLabel(r.Context(), ctrl.config, template, pngData, itemData.Name, itemData.URL(), copies); err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Msg("print command failed")
				return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
					Success:    false,
					Message:    fmt.Sprintf("Print failed after %d labels: %v", labelsPrinted, err),
					LabelCount: labelsPrinted,
				})
			}

			labelsPrinted += copies
		}

		if labelsPrinted == 0 {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success: false,
				Message: "No labels were printed. Check that the items exist.",
			})
		}

		message := fmt.Sprintf("Successfully sent %d label(s) to the print command", labelsPrinted)
		if labelsSkipped > 0 {
			message = fmt.Sprintf("Sent %d label(s) to the print command (%d skipped due to errors)", labelsPrinted, labelsSkipped)
		}

		return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
			Success:    true,
			Message:    message,
			LabelCount: labelsPrinted,
		})
	}
}

// maxAssetLabelRange caps the number of labels one asset range request renders
const maxAssetLabelRange = 1000

// LabelTemplateRenderAssetsRequest represents a request to render labels for a range of asset IDs
type LabelTemplateRenderAssetsRequest struct {
	StartAssetID  int    `json:"startAssetId"  validate:"min=1"`
	EndAssetID    int    `json:"endAssetId"    validate:"min=1"`
	SkipLabels    int    `json:"skipLabels"    validate:"min=0,max=1000"` // Empty slots before the first label, for partly used sheets
	Format        string `json:"format"`                                  // "png" or "pdf", defaults to "pdf"
	PageSize      string `json:"pageSize"`                                // "Letter" or "A4" for PDF
	ShowCutGuides bool   `json:"showCutGuides"`                           // Draw light borders around labels for cutting
}

// HandleLabelTemplatesRenderAssets godoc
//
//	@Summary		Render Labels for an Asset ID Range
//	@Description	Renders one label per asset ID in the range. Asset IDs that no entity has yet get a label with only the asset ID and asset URL, for pre-printing tags.
//	@Tags			Label Templates
//	@Accept			json
//	@Produce		image/png,application/pdf
//	@Param			id		path	string								true	"Template ID"
//	@Param			payload	body	LabelTemplateRenderAssetsRequest	true	"Asset range to render"
//	@Success		200		{file}	binary
//	@Router			/v1/label-templates/{id}/render-assets [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleLabelTemplatesRenderAssets() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		req, err := adapters.DecodeBody[LabelTemplateRenderAssetsRequest](r)
		if err != nil {
			return err
		}

		if req.EndAssetID < req.StartAssetID {
			return validate.NewRequestError(errors.New("end asset ID must not be before the start"), http.StatusBadRequest)
		}
		if req.EndAssetID-req.StartAssetID+1 > maxAssetLabelRange {
			return validate.NewRequestError(fmt.Errorf("at most %d asset labels can be rendered at once", maxAssetLabelRange), http.StatusBadRequest)
		}

		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: template.CanvasData,
		}
		sheetLayout := sheetLayoutFor(template, templateData)
		baseURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)

		start, end := repo.AssetID(req.StartAssetID), repo.AssetID(req.EndAssetID)
		existing, err := ctrl.repo.Entities.GetIDsByAssetIDRange(r.Context(), auth.GID, start, end)
		if err != nil {
			return err
		}

		labels := make([]*labelmaker.EntityData, 0, int(end-start)+1)
		for aid := start; aid <= end; aid++ {
			if entityID, ok := existing[aid]; ok {
				entity, err := ctrl.buildEntityData(r.Context(), auth.GID, entityID, baseURL)
				if err == nil {
					labels = append(labels, entity)
					continue
				}
				log.Warn().Err(err).Str("entityID", entityID.String()).Msg("failed to fetch entity for asset label, rendering it blank")
			}

			// No entity has this asset ID yet: a blank tag carrying only the ID
			labels = append(labels, &labelmaker.EntityData{
				AssetID:      aid.String(),
				LocationPath: []string{},
				CustomFields: map[string]string{},
				AssetURL:     assetURL(baseURL, aid),
			})
		}

		if req.Format == "png" {
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Entity:   labels[0],
				Template: templateData,
			})
			if err != nil {
				return err
			}
			w.Header().Set("Content-Type", mimePNG)
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=asset-label-%s.png", start.String()))
			_, err = w.Write(pngData)
			return err
		}

		pdfOpts := pdfOptionsFor(template, req.PageSize, req.ShowCutGuides, req.SkipLabels)
		return renderLabelsPDF(w, renderer, templateData, labels, pdfOpts, sheetLayout,
			fmt.Sprintf("asset-labels-%s-%s.pdf", start.String(), end.String()))
	}
}

// LabelTemplateRenderLocationsRequest represents a request to render labels for locations
type LabelTemplateRenderLocationsRequest struct {
	LocationIDs   []uuid.UUID `json:"locationIds"   validate:"required,min=1"`
	Format        string      `json:"format"`        // "png" or "pdf", defaults to "png"
	PageSize      string      `json:"pageSize"`      // "Letter", "A4", or "Custom" for PDF
	ShowCutGuides bool        `json:"showCutGuides"` // Draw light borders around labels for cutting
	SkipLabels    int         `json:"skipLabels"`    // Leave this many positions empty on the first sheet
}

// HandleLabelTemplatesRenderLocations godoc
//
//	@Summary	Render Label for Locations
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	image/png,application/pdf
//	@Param		id		path	string								true	"Template ID"
//	@Param		payload	body	LabelTemplateRenderLocationsRequest	true	"Locations to render"
//	@Success	200		{file}	binary
//	@Router		/v1/label-templates/{id}/render-locations [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesRenderLocations() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		// Parse request body
		req, err := adapters.DecodeBody[LabelTemplateRenderLocationsRequest](r)
		if err != nil {
			return err
		}

		if len(req.LocationIDs) == 0 {
			return fmt.Errorf("at least one location ID is required")
		}
		if len(req.LocationIDs) > maxPrintLabelsPerRequest {
			return validate.NewRequestError(fmt.Errorf("too many locations in one render request (max %d)", maxPrintLabelsPerRequest), http.StatusBadRequest)
		}

		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: template.CanvasData,
		}

		// Determine output format
		format := req.Format
		if format == "" {
			format = "png"
		}

		if format == "pdf" {
			baseURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)
			locations := ctrl.fetchEntitiesData(r.Context(), auth.GID, req.LocationIDs, baseURL)
			if len(locations) == 0 {
				return fmt.Errorf("no valid locations found")
			}

			sheetLayout := sheetLayoutFor(template, templateData)
			pdfOpts := pdfOptionsFor(template, req.PageSize, req.ShowCutGuides, req.SkipLabels)
			return renderLabelsPDF(w, renderer, templateData, locations, pdfOpts, sheetLayout, fmt.Sprintf("location-labels-%d.pdf", len(locations)))
		}

		// PNG format: render single location
		locationData, err := ctrl.buildEntityData(r.Context(), auth.GID, req.LocationIDs[0], GetHBURL(r, &ctrl.config.Options, ctrl.url))
		if err != nil {
			return err
		}

		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Entity:   locationData,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mimePNG)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=location-label-%s.png", locationData.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// LabelPrintLocation represents a location to print with its quantity
type LabelPrintLocation struct {
	ID       uuid.UUID `json:"id"       validate:"required"`
	Quantity int       `json:"quantity"` // Number of copies for this location
}

// LabelTemplatePrintLocationsRequest represents a request to print location labels directly
type LabelTemplatePrintLocationsRequest struct {
	LocationIDs []uuid.UUID          `json:"locationIds,omitempty"` // Simple list (1 copy each)
	Locations   []LabelPrintLocation `json:"locations,omitempty"`   // Locations with individual quantities
	Copies      int                  `json:"copies"`                // Default copies per label
}

// HandleLabelTemplatesPrintLocations godoc
//
//	@Summary	Print Location Labels
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string								true	"Template ID"
//	@Param		payload	body		LabelTemplatePrintLocationsRequest	true	"Print request"
//	@Success	200		{object}	LabelTemplatePrintResponse
//	@Router		/v1/label-templates/{id}/print-locations [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesPrintLocations() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		// Parse request body
		req, err := adapters.DecodeBody[LabelTemplatePrintLocationsRequest](r)
		if err != nil {
			return err
		}

		// Build locations list from either Locations (with quantities) or LocationIDs
		var locationsToPrint []LabelPrintLocation
		switch {
		case len(req.Locations) > 0:
			locationsToPrint = req.Locations
		case len(req.LocationIDs) > 0:
			for _, locID := range req.LocationIDs {
				locationsToPrint = append(locationsToPrint, LabelPrintLocation{ID: locID, Quantity: 0})
			}
		default:
			return fmt.Errorf("at least one location is required (provide locationIds or locations)")
		}
		if len(locationsToPrint) > maxPrintLabelsPerRequest {
			return validate.NewRequestError(fmt.Errorf("too many locations in one print request (max %d)", maxPrintLabelsPerRequest), http.StatusBadRequest)
		}
		locationQuantities := make([]int, len(locationsToPrint))
		for i, loc := range locationsToPrint {
			locationQuantities[i] = loc.Quantity
		}
		if total := totalLabelCount(clampCopies(req.Copies), locationQuantities); total > maxPrintTotalLabels {
			return validate.NewRequestError(fmt.Errorf("too many labels in one print request (%d, max %d)", total, maxPrintTotalLabels), http.StatusBadRequest)
		}

		if cmd := ctrl.config.LabelMaker.PrintCommand; cmd == nil || *cmd == "" {
			return validate.NewRequestError(errors.New("server-side label printing is not configured; set HBOX_LABEL_MAKER_PRINT_COMMAND"), http.StatusBadRequest)
		}

		// Get the template
		template, err := ctrl.repo.LabelTemplates.GetOne(r.Context(), auth.GID, auth.UID, id)
		if err != nil {
			return err
		}

		// Create renderer
		renderer, err := labelmaker.NewTemplateRenderer()
		if err != nil {
			return err
		}

		templateData := &labelmaker.TemplateData{
			ID:         template.ID,
			Name:       template.Name,
			Width:      template.Width,
			Height:     template.Height,
			DPI:        template.DPI,
			CanvasData: template.CanvasData,
		}

		// Default copies
		defaultCopies := clampCopies(req.Copies)

		labelsPrinted := 0
		labelsSkipped := 0

		for _, printLoc := range locationsToPrint {
			locationData, err := ctrl.buildEntityData(r.Context(), auth.GID, printLoc.ID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
			if err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Msg("failed to fetch location for printing")
				labelsSkipped++
				continue
			}

			// Render label as PNG
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Entity:   locationData,
				Template: templateData,
			})
			if err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Str("locationName", locationData.Name).Msg("failed to render location label for printing")
				labelsSkipped++
				continue
			}

			// Determine copies for this location
			copies := defaultCopies
			if printLoc.Quantity > 0 {
				copies = clampCopies(printLoc.Quantity)
			}

			if err := printTemplateLabel(r.Context(), ctrl.config, template, pngData, locationData.Name, locationData.URL(), copies); err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Msg("print command failed")
				return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
					Success:    false,
					Message:    fmt.Sprintf("Print failed after %d labels: %v", labelsPrinted, err),
					LabelCount: labelsPrinted,
				})
			}

			labelsPrinted += copies
		}

		if labelsPrinted == 0 {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success: false,
				Message: "No labels were printed. Check that the locations exist.",
			})
		}

		message := fmt.Sprintf("Successfully sent %d location label(s) to the print command", labelsPrinted)
		if labelsSkipped > 0 {
			message = fmt.Sprintf("Sent %d location label(s) to the print command (%d skipped due to errors)", labelsPrinted, labelsSkipped)
		}

		return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
			Success:    true,
			Message:    message,
			LabelCount: labelsPrinted,
		})
	}
}
