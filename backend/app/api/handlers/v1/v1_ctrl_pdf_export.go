package v1

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// HandleItemExportPDF godoc
//
//	@Summary	Export Single Item as PDF
//	@Tags		Items
//	@Produce	application/pdf
//	@Param		id		path	string	true	"Item ID"
//	@Param		theme	query	string	false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool	false	"Include photos in export (default: true)"
//	@Param		owner	query	string	false	"Owner name for cover page"
//	@Success	200		{file}	file	"PDF document"
//	@Router		/v1/items/{id}/export/pdf [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemExportPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Parse the item ID from the URL path
		itemID, err := ctrl.routeID(r)
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
		pdfBytes, filename, err := pdfSvc.ExportSingleItem(ctx, ctx.GID, itemID, opts)
		if err != nil {
			log.Err(err).Msg("failed to export item as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		// Set response headers for PDF file download
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

		_, err = w.Write(pdfBytes)
		return err
	}
}

// multiExportRequest is the JSON body for bulk PDF export requests.
// Clients send a list of item IDs to include in the report.
type multiExportRequest struct {
	ItemIDs []string `json:"itemIds" validate:"required,min=1"`
}

// HandleItemsExportPDF godoc
//
//	@Summary	Export Multiple Items as PDF
//	@Tags		Items
//	@Accept		json
//	@Produce	application/pdf
//	@Param		payload	body	multiExportRequest	true	"Item IDs to export"
//	@Param		theme	query	string				false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool				false	"Include photos in export (default: true)"
//	@Param		owner	query	string				false	"Owner name for cover page"
//	@Success	200		{file}	file				"PDF document"
//	@Router		/v1/items/export/pdf [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsExportPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		// Decode the request body containing item IDs
		var body multiExportRequest
		if err := server.Decode(r, &body); err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		if len(body.ItemIDs) == 0 {
			return validate.NewRequestError(
				fmt.Errorf("at least one item ID is required"),
				http.StatusBadRequest,
			)
		}

		// Parse string UUIDs into uuid.UUID values
		itemIDs, err := parseUUIDs(body.ItemIDs)
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		// Build export options from query parameters
		opts := services.PDFExportOptions{
			Theme:         r.URL.Query().Get("theme"),
			IncludePhotos: r.URL.Query().Get("photos") != "false",
			OwnerName:     r.URL.Query().Get("owner"),
		}

		// Generate the multi-item PDF report
		pdfSvc := services.NewPDFExportService(ctrl.repo)
		pdfBytes, filename, err := pdfSvc.ExportMultipleItems(ctx, ctx.GID, itemIDs, opts)
		if err != nil {
			log.Err(err).Msg("failed to export items as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		// Set response headers for PDF file download
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

		_, err = w.Write(pdfBytes)
		return err
	}
}

// HandleItemsExportAllPDF godoc
//
//	@Summary	Export All Items as PDF
//	@Tags		Items
//	@Produce	application/pdf
//	@Param		theme	query	string	false	"PDF theme (navy, modern, minimal, forest)"
//	@Param		photos	query	bool	false	"Include photos in export (default: true)"
//	@Param		owner	query	string	false	"Owner name for cover page"
//	@Success	200		{file}	file	"PDF document"
//	@Router		/v1/items/export/pdf [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsExportAllPDF() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		// Query all items for the user's group with pagination disabled (-1 = all results)
		allItems, err := ctrl.repo.Items.QueryByGroup(ctx, ctx.GID, repo.ItemQuery{
			Page:     -1,
			PageSize: -1,
		})
		if err != nil {
			log.Err(err).Msg("failed to query items for PDF export")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		if len(allItems.Items) == 0 {
			return validate.NewRequestError(
				fmt.Errorf("no items found to export"),
				http.StatusNotFound,
			)
		}

		// Collect all item IDs from the query result
		itemIDs := make([]uuid.UUID, len(allItems.Items))
		for i, item := range allItems.Items {
			itemIDs[i] = item.ID
		}

		// Build export options from query parameters
		opts := services.PDFExportOptions{
			Theme:         r.URL.Query().Get("theme"),
			IncludePhotos: r.URL.Query().Get("photos") != "false",
			OwnerName:     r.URL.Query().Get("owner"),
		}

		// Generate the full-inventory PDF report
		pdfSvc := services.NewPDFExportService(ctrl.repo)
		pdfBytes, filename, err := pdfSvc.ExportMultipleItems(ctx, ctx.GID, itemIDs, opts)
		if err != nil {
			log.Err(err).Msg("failed to export all items as PDF")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		// Set response headers for PDF file download
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

		_, err = w.Write(pdfBytes)
		return err
	}
}

// HandlePDFThemes godoc
//
//	@Summary	Get Available PDF Themes
//	@Tags		Items
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/v1/items/export/pdf/themes [GET]
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
