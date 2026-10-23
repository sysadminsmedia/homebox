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
	"strings"

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
		sheetLayout := applyPresetAndGetSheetLayout(template, templateData)

		// Handle PDF format
		format := req.Format
		if format == "" {
			format = "png"
		}

		if format == "pdf" {
			items := ctrl.fetchItemsData(r.Context(), auth.GID, req.ItemIDs, baseURL)
			if len(items) == 0 {
				return fmt.Errorf("no valid items found")
			}
			return renderLabelsPDF(w, renderer, templateData, items, req, sheetLayout)
		}

		// Handle PNG sheet format (multiple items with sheet layout)
		if sheetLayout != nil && len(req.ItemIDs) > 1 {
			items := ctrl.fetchItemsData(r.Context(), auth.GID, req.ItemIDs, baseURL)
			if len(items) == 0 {
				return fmt.Errorf("no valid items found")
			}
			return renderLabelsPNGSheet(w, renderer, templateData, items, sheetLayout)
		}

		// Default: render single item as PNG
		itemData, err := ctrl.loadItemData(r.Context(), auth.GID, req.ItemIDs[0], baseURL)
		if err != nil {
			return err
		}

		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Item:     itemData,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mimePNG)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=label-%s.png", itemData.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// buildLocationData converts a container entity and its path to labelmaker.LocationData
func buildLocationData(loc repo.EntityOut, path []repo.EntityPath, baseURL string) *labelmaker.LocationData {
	// Build the path names
	pathNames := make([]string, 0, len(path))
	for _, p := range path {
		if p.ID != loc.ID { // Exclude the current location from path
			pathNames = append(pathNames, p.Name)
		}
	}

	// Build full path string
	fullPath := loc.Name
	if len(pathNames) > 0 {
		fullPath = strings.Join(pathNames, " > ") + " > " + loc.Name
	}

	return &labelmaker.LocationData{
		ID:          loc.ID,
		Name:        loc.Name,
		Description: loc.Description,
		Path:        pathNames,
		FullPath:    fullPath,
		ItemCount:   containerItemCount(loc),
		LocationURL: fmt.Sprintf("%s/location/%s", baseURL, loc.ID.String()),
	}
}

// containerItemCount sums the quantity of non-location children of a container.
func containerItemCount(loc repo.EntityOut) float64 {
	var total float64
	for _, child := range loc.Children {
		if child.Archived {
			continue
		}
		if child.EntityType != nil && child.EntityType.IsLocation {
			continue
		}
		total += child.Quantity
	}
	return total
}

// buildItemData converts a repo.EntityOut (plus its ancestor path) to labelmaker.ItemData
func buildItemData(item repo.EntityOut, path []repo.EntityPath, baseURL string) *labelmaker.ItemData {
	tags := make([]string, len(item.Tags))
	for i, t := range item.Tags {
		tags[i] = t.Name
	}

	// Ancestor names, root first, excluding the item itself
	locationPath := make([]string, 0, len(path))
	for _, p := range path {
		if p.ID != item.ID {
			locationPath = append(locationPath, p.Name)
		}
	}

	customFields := make(map[string]string)
	for _, f := range item.Fields {
		customFields[f.Name] = f.TextValue
	}

	var locationName string
	switch {
	case item.Location != nil:
		locationName = item.Location.Name
	case item.Parent != nil:
		locationName = item.Parent.Name
	}

	return &labelmaker.ItemData{
		ID:           item.ID,
		Name:         item.Name,
		Description:  item.Description,
		AssetID:      item.AssetID.String(),
		SerialNumber: item.SerialNumber,
		ModelNumber:  item.ModelNumber,
		Manufacturer: item.Manufacturer,
		LocationName: locationName,
		LocationPath: locationPath,
		Tags:         tags,
		CustomFields: customFields,
		ItemURL:      fmt.Sprintf("%s/item/%s", baseURL, item.ID.String()),
		Quantity:     item.Quantity,
		Notes:        item.Notes,
	}
}

// loadItemData fetches an entity and its ancestor path and converts it to labelmaker.ItemData
func (ctrl *V1Controller) loadItemData(ctx context.Context, gid, id uuid.UUID, baseURL string) (*labelmaker.ItemData, error) {
	item, err := ctrl.repo.Entities.GetOneByGroup(ctx, gid, id)
	if err != nil {
		return nil, err
	}
	path, err := ctrl.repo.Entities.PathForEntity(ctx, gid, id)
	if err != nil {
		log.Warn().Err(err).Str("itemID", id.String()).Msg("failed to resolve item path for label rendering")
		path = nil
	}
	return buildItemData(item, path, baseURL), nil
}

// fetchItemsData fetches items by ID and converts them to labelmaker.ItemData
func (ctrl *V1Controller) fetchItemsData(ctx context.Context, gid uuid.UUID, itemIDs []uuid.UUID, baseURL string) []*labelmaker.ItemData {
	items := make([]*labelmaker.ItemData, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		itemData, err := ctrl.loadItemData(ctx, gid, itemID, baseURL)
		if err != nil {
			log.Warn().Err(err).Str("itemID", itemID.String()).Msg("failed to fetch item for label rendering")
			continue
		}
		items = append(items, itemData)
	}
	return items
}

// applyPresetAndGetSheetLayout extracts sheet layout from template preset if available
func applyPresetAndGetSheetLayout(template repo.LabelTemplateOut, templateData *labelmaker.TemplateData) *labelmaker.SheetLayout {
	if template.Preset == nil || *template.Preset == "" {
		return nil
	}
	preset := labelmaker.GetPresetByKey(*template.Preset)
	if preset == nil || preset.SheetLayout == nil {
		return nil
	}
	// Use preset's exact dimensions for proper sheet alignment
	templateData.Width = preset.Width
	templateData.Height = preset.Height
	return preset.SheetLayout
}

// renderLabelsPDF renders labels to PDF format
func renderLabelsPDF(w http.ResponseWriter, renderer *labelmaker.TemplateRenderer, templateData *labelmaker.TemplateData, items []*labelmaker.ItemData, req LabelTemplateRenderRequest, sheetLayout *labelmaker.SheetLayout) error {
	pdfOpts := labelmaker.DefaultPDFOptions()
	if req.PageSize == "A4" {
		pdfOpts.PageSize = labelmaker.PDFPageA4
	} else {
		pdfOpts.PageSize = labelmaker.PDFPageLetter
	}
	pdfOpts.ShowCutGuides = req.ShowCutGuides
	pdfOpts.SkipLabels = req.SkipLabels

	pdfData, err := renderer.RenderToPDFWithSheetLayout(templateData, items, pdfOpts, sheetLayout)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=labels-%d.pdf", len(items)))
	_, err = w.Write(pdfData)
	return err
}

// renderLabelsPNGSheet renders multiple labels to a PNG sheet
func renderLabelsPNGSheet(w http.ResponseWriter, renderer *labelmaker.TemplateRenderer, templateData *labelmaker.TemplateData, items []*labelmaker.ItemData, sheetLayout *labelmaker.SheetLayout) error {
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
func printTemplateLabel(cfg *config.Config, tmpl repo.LabelTemplateOut, pngData []byte, title, labelURL string, copies int) error {
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
	return labelmaker.PrintPNG(cfg, pngData, data, copies)
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
)

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
			itemData, err := ctrl.loadItemData(r.Context(), auth.GID, printItem.ID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
			if err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Msg("failed to fetch item for printing")
				labelsSkipped++
				continue
			}

			// Render label as PNG
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Item:     itemData,
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

			if err := printTemplateLabel(ctrl.config, template, pngData, itemData.Name, itemData.ItemURL, copies); err != nil {
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
			// Render multiple locations to PDF
			locations := make([]*labelmaker.LocationData, 0, len(req.LocationIDs))
			for _, locID := range req.LocationIDs {
				loc, err := ctrl.getLocationWithCount(r.Context(), auth.GID, locID)
				if err != nil {
					log.Warn().Err(err).Str("locationID", locID.String()).Msg("failed to fetch location for label rendering")
					continue
				}
				path, _ := ctrl.repo.Entities.PathForEntity(r.Context(), auth.GID, locID)
				locations = append(locations, buildLocationData(loc, path, GetHBURL(r, &ctrl.config.Options, ctrl.url)))
			}

			if len(locations) == 0 {
				return fmt.Errorf("no valid locations found")
			}

			// Configure PDF options
			pdfOpts := labelmaker.DefaultPDFOptions()
			if req.PageSize == "A4" {
				pdfOpts.PageSize = labelmaker.PDFPageA4
			} else {
				pdfOpts.PageSize = labelmaker.PDFPageLetter
			}
			pdfOpts.ShowCutGuides = req.ShowCutGuides
			pdfOpts.SkipLabels = req.SkipLabels

			// Check if template has a preset with sheet layout (Avery-style)
			var sheetLayout *labelmaker.SheetLayout
			if template.Preset != nil && *template.Preset != "" {
				preset := labelmaker.GetPresetByKey(*template.Preset)
				if preset != nil && preset.SheetLayout != nil {
					sheetLayout = preset.SheetLayout
					templateData.Width = preset.Width
					templateData.Height = preset.Height
				}
			}

			// Use sheet layout if available, otherwise use default grid layout
			pdfData, err := renderer.RenderLocationsToPDFWithSheetLayout(templateData, locations, pdfOpts, sheetLayout)
			if err != nil {
				return err
			}

			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=location-labels-%d.pdf", len(locations)))
			_, err = w.Write(pdfData)
			return err
		}

		// PNG format: render single location
		loc, err := ctrl.getLocationWithCount(r.Context(), auth.GID, req.LocationIDs[0])
		if err != nil {
			return err
		}

		path, _ := ctrl.repo.Entities.PathForEntity(r.Context(), auth.GID, req.LocationIDs[0])
		locationData := buildLocationData(loc, path, GetHBURL(r, &ctrl.config.Options, ctrl.url))

		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Location: locationData,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mimePNG)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=location-label-%s.png", loc.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// getLocationWithCount fetches a container entity with its children so an item count can be derived
func (ctrl *V1Controller) getLocationWithCount(ctx context.Context, gid, locID uuid.UUID) (repo.EntityOut, error) {
	return ctrl.repo.Entities.GetContainerByGroup(ctx, gid, locID)
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
			loc, err := ctrl.getLocationWithCount(r.Context(), auth.GID, printLoc.ID)
			if err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Msg("failed to fetch location for printing")
				labelsSkipped++
				continue
			}

			path, _ := ctrl.repo.Entities.PathForEntity(r.Context(), auth.GID, printLoc.ID)
			locationData := buildLocationData(loc, path, GetHBURL(r, &ctrl.config.Options, ctrl.url))

			// Render label as PNG
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Location: locationData,
				Template: templateData,
			})
			if err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Str("locationName", loc.Name).Msg("failed to render location label for printing")
				labelsSkipped++
				continue
			}

			// Determine copies for this location
			copies := defaultCopies
			if printLoc.Quantity > 0 {
				copies = clampCopies(printLoc.Quantity)
			}

			if err := printTemplateLabel(ctrl.config, template, pngData, loc.Name, locationData.LocationURL, copies); err != nil {
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
