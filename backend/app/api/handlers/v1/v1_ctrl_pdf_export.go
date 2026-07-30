package v1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// sanitizeFilename removes or escapes characters that could cause issues
// in HTTP Content-Disposition headers (e.g., header injection via quotes,
// newlines, or semicolons in user-controlled asset IDs).
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"\"", "'",
		"\n", "",
		"\r", "",
		";", "_",
	)
	return replacer.Replace(name)
}

// HandleEntityExportPDF godoc
//
//	@Summary	Export Single Entity as PDF
//	@Tags		Entities
//	@Produce	application/pdf
//	@Param		id		path	string	true	"Entity ID"
//	@Param		theme	query	string	false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool	false	"Include photos in export (default: true)"
//	@Param		owner	query	string	false	"Owner name for cover page"
//	@Success	200		{file}	file	"PDF document"
//	@Router		/v1/entities/{id}/export/pdf [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityExportPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Parse the entity ID from the URL path
		entityID, err := ctrl.routeID(r)
		if err != nil {
			return err
		}

		ctx := services.NewContext(r.Context())

		// Build export options from query parameters
		opts := services.PDFExportOptions{
			Theme:         r.URL.Query().Get("theme"),
			IncludePhotos: r.URL.Query().Get("photos") != "false", // default true
			OwnerName:     r.URL.Query().Get("owner"),
		}

		// Generate the PDF using the export service
		pdfSvc := services.NewPDFExportService(ctrl.repo)
		pdfBytes, filename, err := pdfSvc.ExportSingleItem(ctx, ctx.GID, entityID, opts)
		if err != nil {
			log.Err(err).Msg("failed to export entity as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		writePDFResponse(w, pdfBytes, filename)
		return nil
	}
}

// multiExportRequest is the JSON body for bulk PDF export requests.
// Clients send a list of entity IDs to include in the report.
type multiExportRequest struct {
	ItemIDs []string `json:"itemIds" validate:"required,min=1"`
}

// HandleEntitiesExportPDF godoc
//
//	@Summary	Export Multiple Entities as PDF
//	@Tags		Entities
//	@Accept		json
//	@Produce	application/pdf
//	@Param		payload	body	multiExportRequest	true	"Entity IDs to export"
//	@Param		theme	query	string				false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool				false	"Include photos in export (default: true)"
//	@Param		owner	query	string				false	"Owner name for cover page"
//	@Success	200		{file}	file				"PDF document"
//	@Router		/v1/entities/export/pdf [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesExportPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		// Decode the request body containing entity IDs
		var body multiExportRequest
		if err := server.Decode(r, &body); err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		if len(body.ItemIDs) == 0 {
			return validate.NewRequestError(
				fmt.Errorf("at least one entity ID is required"),
				http.StatusBadRequest,
			)
		}

		// Enforce the configurable export limit
		maxItems := ctrl.config.Options.PDFExportMaxItems
		if len(body.ItemIDs) > maxItems {
			return validate.NewRequestError(
				fmt.Errorf("too many items to export (%d); maximum is %d (configurable via HBOX_OPTIONS_PDF_EXPORT_MAX_ITEMS)", len(body.ItemIDs), maxItems),
				http.StatusBadRequest,
			)
		}

		// Parse string UUIDs into uuid.UUID values
		entityIDs, err := parseUUIDs(body.ItemIDs)
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		// Build export options from query parameters
		opts := services.PDFExportOptions{
			Theme:         r.URL.Query().Get("theme"),
			IncludePhotos: r.URL.Query().Get("photos") != "false",
			OwnerName:     r.URL.Query().Get("owner"),
		}

		// Generate the multi-entity PDF report
		pdfSvc := services.NewPDFExportService(ctrl.repo)
		pdfBytes, filename, err := pdfSvc.ExportMultipleItems(ctx, ctx.GID, entityIDs, opts)
		if err != nil {
			log.Err(err).Msg("failed to export entities as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		writePDFResponse(w, pdfBytes, filename)
		return nil
	}
}

// HandleEntitiesExportAllPDF godoc
//
//	@Summary	Export All Entities as PDF
//	@Tags		Entities
//	@Produce	application/pdf
//	@Param		theme	query	string	false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool	false	"Include photos in export (default: true)"
//	@Param		owner	query	string	false	"Owner name for cover page"
//	@Success	200		{file}	file	"PDF document"
//	@Router		/v1/entities/export/pdf [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesExportAllPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		// Query all items (not locations) for the user's group with
		// pagination disabled (-1 = all results)
		itemsOnly := false
		allItems, err := ctrl.repo.Entities.QueryByGroup(ctx, ctx.GID, repo.EntityQuery{
			IsLocation: &itemsOnly,
			Page:       -1,
			PageSize:   -1,
		})
		if err != nil {
			log.Err(err).Msg("failed to query entities for PDF export")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		if len(allItems.Items) == 0 {
			return validate.NewRequestError(
				fmt.Errorf("no items found to export"),
				http.StatusNotFound,
			)
		}

		// Enforce the configurable export limit to prevent excessive memory
		// usage and timeouts
		maxItems := ctrl.config.Options.PDFExportMaxItems
		if len(allItems.Items) > maxItems {
			return validate.NewRequestError(
				fmt.Errorf("too many items to export (%d); maximum is %d (configurable via HBOX_OPTIONS_PDF_EXPORT_MAX_ITEMS) — use bulk export with specific item IDs instead", len(allItems.Items), maxItems),
				http.StatusBadRequest,
			)
		}

		// Collect all entity IDs from the query result
		entityIDs := make([]uuid.UUID, len(allItems.Items))
		for i, item := range allItems.Items {
			entityIDs[i] = item.ID
		}

		// Build export options from query parameters
		opts := services.PDFExportOptions{
			Theme:         r.URL.Query().Get("theme"),
			IncludePhotos: r.URL.Query().Get("photos") != "false",
			OwnerName:     r.URL.Query().Get("owner"),
		}

		// Generate the full-inventory PDF report
		pdfSvc := services.NewPDFExportService(ctrl.repo)
		pdfBytes, filename, err := pdfSvc.ExportMultipleItems(ctx, ctx.GID, entityIDs, opts)
		if err != nil {
			log.Err(err).Msg("failed to export all entities as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		writePDFResponse(w, pdfBytes, filename)
		return nil
	}
}

// HandlePDFThemes godoc
//
//	@Summary	Get Available PDF Themes
//	@Tags		Entities
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/v1/entities/export/pdf/themes [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePDFThemes() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Return a map of theme key -> display name for the frontend to render
		themes := make(map[string]string)
		for key, theme := range services.PDFThemes {
			themes[key] = theme.Name
		}

		return server.JSON(w, http.StatusOK, themes)
	}
}

// writePDFResponse sets download headers and writes the PDF bytes to the response.
func writePDFResponse(w http.ResponseWriter, pdfBytes []byte, filename string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sanitizeFilename(filename)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	if _, err := w.Write(pdfBytes); err != nil {
		log.Err(err).Msg("failed to write PDF export response")
	}
}

// parseUUIDs converts a slice of string UUIDs to uuid.UUID values.
// Returns an error if any string is not a valid UUID.
func parseUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %s", s)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
