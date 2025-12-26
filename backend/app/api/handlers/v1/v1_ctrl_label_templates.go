package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"github.com/sysadminsmedia/homebox/backend/pkgs/labelmaker"
	"github.com/sysadminsmedia/homebox/backend/pkgs/printer"
)

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
//	@Param		id	path		string	true	"Template ID"
//	@Success	200	{file}		binary
//	@Router		/v1/label-templates/{id}/preview [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesPreview() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		idParam := r.PathValue("id")
		id, err := uuid.Parse(idParam)
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

		w.Header().Set("Content-Type", "image/png")
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
	CanvasData    string      `json:"canvasData,omitempty"` // Optional: canvas data for live preview (overrides saved template)
}

// HandleLabelTemplatesRender godoc
//
//	@Summary	Render Label for Items
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	image/png,application/pdf
//	@Param		id		path		string						true	"Template ID"
//	@Param		payload	body		LabelTemplateRenderRequest	true	"Items to render"
//	@Success	200		{file}		binary
//	@Router		/v1/label-templates/{id}/render [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesRender() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		id, err := uuid.Parse(r.PathValue("id"))
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

		baseURL := GetHBURL(r.Header.Get("Referer"), ctrl.url)
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
		item, err := ctrl.repo.Items.GetOneByGroup(r.Context(), auth.GID, req.ItemIDs[0])
		if err != nil {
			return err
		}

		itemData := buildItemData(item, baseURL)
		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Item:     itemData,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=label-%s.png", item.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// buildLocationData converts a repo.LocationOutCount and path to labelmaker.LocationData
func buildLocationData(loc repo.LocationOutCount, path []repo.ItemPath, baseURL string) *labelmaker.LocationData {
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
		ItemCount:   loc.ItemCount,
		LocationURL: fmt.Sprintf("%s/location/%s", baseURL, loc.ID.String()),
	}
}

// buildItemData converts a repo.ItemOut to labelmaker.ItemData
func buildItemData(item repo.ItemOut, baseURL string) *labelmaker.ItemData {
	labels := make([]string, len(item.Labels))
	for i, l := range item.Labels {
		labels[i] = l.Name
	}

	locationPath := []string{}
	if item.Location != nil {
		locationPath = append(locationPath, item.Location.Name)
	}

	customFields := make(map[string]string)
	for _, f := range item.Fields {
		customFields[f.Name] = f.TextValue
	}

	var locationName string
	if item.Location != nil {
		locationName = item.Location.Name
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
		Labels:       labels,
		CustomFields: customFields,
		ItemURL:      fmt.Sprintf("%s/item/%s", baseURL, item.ID.String()),
		Quantity:     item.Quantity,
		Notes:        item.Notes,
	}
}

// fetchItemsData fetches items by ID and converts them to labelmaker.ItemData
func (ctrl *V1Controller) fetchItemsData(ctx context.Context, gid uuid.UUID, itemIDs []uuid.UUID, baseURL string) []*labelmaker.ItemData {
	items := make([]*labelmaker.ItemData, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		item, err := ctrl.repo.Items.GetOneByGroup(ctx, gid, itemID)
		if err != nil {
			log.Warn().Err(err).Str("itemID", itemID.String()).Msg("failed to fetch item for label rendering")
			continue
		}
		items = append(items, buildItemData(item, baseURL))
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
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=labels-sheet-%d.png", len(items)))
	if len(sheets) > 1 {
		w.Header().Set("X-Labels-Warning", fmt.Sprintf("Only returning first sheet of %d. Use PDF format for all sheets.", len(sheets)))
	}
	_, err = w.Write(sheets[0])
	return err
}

// LabelPrintItem represents an item to print with its quantity
type LabelPrintItem struct {
	ID       uuid.UUID `json:"id"       validate:"required"`
	Quantity int       `json:"quantity"` // Number of copies for this item
}

// LabelTemplatePrintRequest represents a request to print labels directly to a printer
type LabelTemplatePrintRequest struct {
	ItemIDs   []uuid.UUID      `json:"itemIds,omitempty"`   // Simple list (1 copy each) - for backward compatibility
	Items     []LabelPrintItem `json:"items,omitempty"`     // Items with individual quantities
	PrinterID *uuid.UUID       `json:"printerId,omitempty"` // If nil, uses default printer
	Copies    int              `json:"copies"`              // Default copies per label (used if item.quantity is 0)
}

// LabelTemplatePrintResponse represents the result of a direct print operation
type LabelTemplatePrintResponse struct {
	Success     bool   `json:"success"`
	JobID       int    `json:"jobId,omitempty"`
	Message     string `json:"message"`
	LabelCount  int    `json:"labelCount"`
	PrinterName string `json:"printerName"`
}

// HandleLabelTemplatesPrint godoc
//
//	@Summary	Print Labels Directly to Printer
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

		idParam := r.PathValue("id")
		id, err := uuid.Parse(idParam)
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

		// Get the printer
		var printerOut repo.PrinterOut
		if req.PrinterID != nil {
			printerOut, err = ctrl.repo.Printers.GetOne(r.Context(), auth.GID, *req.PrinterID)
		} else {
			printerOut, err = ctrl.repo.Printers.GetDefault(r.Context(), auth.GID)
		}
		if err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success: false,
				Message: "No printer found. Please configure a printer first.",
			})
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

		// Re-validate printer address for defense-in-depth (prevents SSRF if DB is compromised)
		if err := printer.ValidatePrinterAddress(printerOut.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "Printer address validation failed: " + err.Error(),
				PrinterName: printerOut.Name,
			})
		}

		// Create printer client
		printerClient, err := printer.NewPrinterClient(printer.PrinterType(printerOut.PrinterType), printerOut.Address)
		if err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "Failed to connect to printer: " + err.Error(),
				PrinterName: printerOut.Name,
			})
		}

		// Configure Brother raster client with media type from template
		if printer.PrinterType(printerOut.PrinterType) == printer.PrinterTypeBrotherRaster {
			if brotherClient, ok := printerClient.(*printer.BrotherRasterClient); ok {
				mediaType := template.MediaType
				if mediaType == "" {
					mediaType = "DK-22251" // Default to DK-22251
				}
				if err := brotherClient.SetMediaType(mediaType); err != nil {
					return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
						Success:     false,
						Message:     "Invalid media type: " + err.Error(),
						PrinterName: printerOut.Name,
					})
				}
			}
		}

		// Default copies (used when item quantity is 0 or not specified)
		defaultCopies := req.Copies
		if defaultCopies <= 0 {
			defaultCopies = 1
		}

		var lastJobID int
		labelsPrinted := 0
		labelsSkipped := 0

		for _, printItem := range itemsToPrint {
			item, err := ctrl.repo.Items.GetOneByGroup(r.Context(), auth.GID, printItem.ID)
			if err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Msg("failed to fetch item for printing")
				labelsSkipped++
				continue
			}

			itemData := buildItemData(item, GetHBURL(r.Header.Get("Referer"), ctrl.url))

			// Render label as PNG
			pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
				Item:     itemData,
				Template: templateData,
			})
			if err != nil {
				log.Warn().Err(err).Str("itemID", printItem.ID.String()).Str("itemName", item.Name).Msg("failed to render label for printing")
				labelsSkipped++
				continue
			}

			// Prepare print data based on printer type
			var printData []byte
			var contentType string

			if printer.PrinterType(printerOut.PrinterType) == printer.PrinterTypeBrotherRaster {
				// Brother raster printers: send PNG directly (client converts to raster)
				printData = pngData
				contentType = "image/png"
			} else {
				// IPP/CUPS printers: convert PNG to URF format
				dpi := template.DPI
				if dpi <= 0 {
					dpi = 300
				}
				urfData, err := printer.ConvertPNGToURF(pngData, dpi)
				if err != nil {
					log.Warn().Err(err).Str("itemID", printItem.ID.String()).Str("itemName", item.Name).Msg("failed to convert label to URF format")
					labelsSkipped++
					continue
				}
				printData = urfData
				contentType = "image/urf"
			}

			// Determine copies for this item (use per-item quantity if set, otherwise default)
			copies := printItem.Quantity
			if copies <= 0 {
				copies = defaultCopies
			}

			// Send to printer
			result, err := printerClient.Print(r.Context(), &printer.PrintJob{
				DocumentName: fmt.Sprintf("Label - %s", item.Name),
				ContentType:  contentType,
				Data:         printData,
				Copies:       copies,
			})
			if err != nil {
				return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
					Success:     false,
					Message:     fmt.Sprintf("Print failed after %d labels: %v", labelsPrinted, err),
					LabelCount:  labelsPrinted,
					PrinterName: printerOut.Name,
				})
			}

			lastJobID = result.JobID
			labelsPrinted += copies
		}

		if labelsPrinted == 0 {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "No labels were printed. Check that the items exist.",
				PrinterName: printerOut.Name,
			})
		}

		message := fmt.Sprintf("Successfully sent %d label(s) to printer", labelsPrinted)
		if labelsSkipped > 0 {
			message = fmt.Sprintf("Sent %d label(s) to printer (%d skipped due to errors)", labelsPrinted, labelsSkipped)
		}

		return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
			Success:     true,
			JobID:       lastJobID,
			Message:     message,
			LabelCount:  labelsPrinted,
			PrinterName: printerOut.Name,
		})
	}
}

// LabelTemplateRenderLocationsRequest represents a request to render labels for locations
type LabelTemplateRenderLocationsRequest struct {
	LocationIDs   []uuid.UUID `json:"locationIds"   validate:"required,min=1"`
	Format        string      `json:"format"`        // "png" or "pdf", defaults to "png"
	PageSize      string      `json:"pageSize"`      // "Letter", "A4", or "Custom" for PDF
	ShowCutGuides bool        `json:"showCutGuides"` // Draw light borders around labels for cutting
}

// HandleLabelTemplatesRenderLocations godoc
//
//	@Summary	Render Label for Locations
//	@Tags		Label Templates
//	@Accept		json
//	@Produce	image/png,application/pdf
//	@Param		id		path		string								true	"Template ID"
//	@Param		payload	body		LabelTemplateRenderLocationsRequest	true	"Locations to render"
//	@Success	200		{file}		binary
//	@Router		/v1/label-templates/{id}/render-locations [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelTemplatesRenderLocations() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		idParam := r.PathValue("id")
		id, err := uuid.Parse(idParam)
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
				path, _ := ctrl.repo.Locations.PathForLoc(r.Context(), auth.GID, locID)
				locations = append(locations, buildLocationData(loc, path, GetHBURL(r.Header.Get("Referer"), ctrl.url)))
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

		path, _ := ctrl.repo.Locations.PathForLoc(r.Context(), auth.GID, req.LocationIDs[0])
		locationData := buildLocationData(loc, path, GetHBURL(r.Header.Get("Referer"), ctrl.url))

		pngData, err := renderer.RenderTemplate(labelmaker.RenderContext{
			Location: locationData,
			Template: templateData,
		})
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=location-label-%s.png", loc.ID.String()))
		_, err = w.Write(pngData)
		return err
	}
}

// getLocationWithCount fetches a location and adds item count
func (ctrl *V1Controller) getLocationWithCount(ctx context.Context, gid, locID uuid.UUID) (repo.LocationOutCount, error) {
	return ctrl.repo.Locations.GetOneByGroupWithCount(ctx, gid, locID)
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
	PrinterID   *uuid.UUID           `json:"printerId,omitempty"`   // If nil, uses default printer
	Copies      int                  `json:"copies"`                // Default copies per label
}

// HandleLabelTemplatesPrintLocations godoc
//
//	@Summary	Print Location Labels Directly to Printer
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

		idParam := r.PathValue("id")
		id, err := uuid.Parse(idParam)
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

		// Get the printer
		var printerOut repo.PrinterOut
		if req.PrinterID != nil {
			printerOut, err = ctrl.repo.Printers.GetOne(r.Context(), auth.GID, *req.PrinterID)
		} else {
			printerOut, err = ctrl.repo.Printers.GetDefault(r.Context(), auth.GID)
		}
		if err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success: false,
				Message: "No printer found. Please configure a printer first.",
			})
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

		// Re-validate printer address for defense-in-depth (prevents SSRF if DB is compromised)
		if err := printer.ValidatePrinterAddress(printerOut.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "Printer address validation failed: " + err.Error(),
				PrinterName: printerOut.Name,
			})
		}

		// Create printer client
		printerClient, err := printer.NewPrinterClient(printer.PrinterType(printerOut.PrinterType), printerOut.Address)
		if err != nil {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "Failed to connect to printer: " + err.Error(),
				PrinterName: printerOut.Name,
			})
		}

		// Configure Brother raster client with media type from template
		if printer.PrinterType(printerOut.PrinterType) == printer.PrinterTypeBrotherRaster {
			if brotherClient, ok := printerClient.(*printer.BrotherRasterClient); ok {
				mediaType := template.MediaType
				if mediaType == "" {
					mediaType = "DK-22251"
				}
				if err := brotherClient.SetMediaType(mediaType); err != nil {
					return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
						Success:     false,
						Message:     "Invalid media type: " + err.Error(),
						PrinterName: printerOut.Name,
					})
				}
			}
		}

		// Default copies
		defaultCopies := req.Copies
		if defaultCopies <= 0 {
			defaultCopies = 1
		}

		var lastJobID int
		labelsPrinted := 0
		labelsSkipped := 0

		for _, printLoc := range locationsToPrint {
			loc, err := ctrl.getLocationWithCount(r.Context(), auth.GID, printLoc.ID)
			if err != nil {
				log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Msg("failed to fetch location for printing")
				labelsSkipped++
				continue
			}

			path, _ := ctrl.repo.Locations.PathForLoc(r.Context(), auth.GID, printLoc.ID)
			locationData := buildLocationData(loc, path, GetHBURL(r.Header.Get("Referer"), ctrl.url))

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

			// Prepare print data based on printer type
			var printData []byte
			var contentType string

			if printer.PrinterType(printerOut.PrinterType) == printer.PrinterTypeBrotherRaster {
				printData = pngData
				contentType = "image/png"
			} else {
				dpi := template.DPI
				if dpi <= 0 {
					dpi = 300
				}
				urfData, err := printer.ConvertPNGToURF(pngData, dpi)
				if err != nil {
					log.Warn().Err(err).Str("locationID", printLoc.ID.String()).Str("locationName", loc.Name).Msg("failed to convert location label to URF format")
					labelsSkipped++
					continue
				}
				printData = urfData
				contentType = "image/urf"
			}

			// Determine copies for this location
			copies := printLoc.Quantity
			if copies <= 0 {
				copies = defaultCopies
			}

			// Send to printer
			result, err := printerClient.Print(r.Context(), &printer.PrintJob{
				DocumentName: fmt.Sprintf("Location Label - %s", loc.Name),
				ContentType:  contentType,
				Data:         printData,
				Copies:       copies,
			})
			if err != nil {
				return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
					Success:     false,
					Message:     fmt.Sprintf("Print failed after %d labels: %v", labelsPrinted, err),
					LabelCount:  labelsPrinted,
					PrinterName: printerOut.Name,
				})
			}

			lastJobID = result.JobID
			labelsPrinted += copies
		}

		if labelsPrinted == 0 {
			return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
				Success:     false,
				Message:     "No labels were printed. Check that the locations exist.",
				PrinterName: printerOut.Name,
			})
		}

		message := fmt.Sprintf("Successfully sent %d location label(s) to printer", labelsPrinted)
		if labelsSkipped > 0 {
			message = fmt.Sprintf("Sent %d location label(s) to printer (%d skipped due to errors)", labelsPrinted, labelsSkipped)
		}

		return server.JSON(w, http.StatusOK, LabelTemplatePrintResponse{
			Success:     true,
			JobID:       lastJobID,
			Message:     message,
			LabelCount:  labelsPrinted,
			PrinterName: printerOut.Name,
		})
	}
}
