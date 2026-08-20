package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"

	"gocloud.dev/blob"
)

const (
	// MaxImageBytes is the maximum size for embedded images (10 MB).
	// Images larger than this are skipped to prevent excessive memory usage.
	MaxImageBytes = 10 * 1024 * 1024
)

// PDFTheme defines the color scheme and styling for PDF exports.
// Multiple themes are available to suit different presentation needs.
type PDFTheme struct {
	Name            string  // Display name of the theme
	HeaderR         int     // Header background red component (0-255)
	HeaderG         int     // Header background green component
	HeaderB         int     // Header background blue component
	AccentR         int     // Accent/divider line red component
	AccentG         int     // Accent/divider line green component
	AccentB         int     // Accent/divider line blue component
	AltRowR         int     // Alternating table row red component
	AltRowG         int     // Alternating table row green component
	AltRowB         int     // Alternating table row blue component
	HeaderFontSize  float64 // Font size for section headers
	BodyFontSize    float64 // Font size for body text
	CoverTitleSize  float64 // Font size for cover page title
	CoverDetailSize float64 // Font size for cover page details
}

// Available PDF themes — keyed by name for user selection via query parameter.
var PDFThemes = map[string]PDFTheme{
	// Navy: professional insurance-style look with navy blue headers
	"navy": {
		Name: "Navy", HeaderR: 26, HeaderG: 54, HeaderB: 93,
		AccentR: 41, AccentG: 98, AccentB: 168,
		AltRowR: 245, AltRowG: 245, AltRowB: 245,
		HeaderFontSize: 14, BodyFontSize: 10, CoverTitleSize: 28, CoverDetailSize: 14,
	},
	// Modern: clean dark slate with teal accents
	"modern": {
		Name: "Modern", HeaderR: 45, HeaderG: 55, HeaderB: 72,
		AccentR: 56, AccentG: 178, AccentB: 172,
		AltRowR: 248, AltRowG: 250, AltRowB: 252,
		HeaderFontSize: 14, BodyFontSize: 10, CoverTitleSize: 28, CoverDetailSize: 14,
	},
	// Minimal: light gray theme for a clean, understated appearance
	"minimal": {
		Name: "Minimal", HeaderR: 75, HeaderG: 85, HeaderB: 99,
		AccentR: 148, AccentG: 163, AccentB: 184,
		AltRowR: 249, AltRowG: 250, AltRowB: 251,
		HeaderFontSize: 13, BodyFontSize: 10, CoverTitleSize: 26, CoverDetailSize: 13,
	},
	// Forest: earthy green tones suited for organic/outdoor inventories
	"forest": {
		Name: "Forest", HeaderR: 34, HeaderG: 87, HeaderB: 59,
		AccentR: 56, AccentG: 142, AccentB: 93,
		AltRowR: 244, AltRowG: 249, AltRowB: 245,
		HeaderFontSize: 14, BodyFontSize: 10, CoverTitleSize: 28, CoverDetailSize: 14,
	},
}

// PDFExportOptions configures what is included in the generated PDF.
type PDFExportOptions struct {
	Theme         string // Theme name (navy, modern, minimal, forest)
	IncludePhotos bool   // Whether to embed item photos
	OwnerName     string // Name shown on cover page
}

// PDFExportService handles generating PDF reports from item data.
// It reads item details from the repository and attachment images from blob storage.
type PDFExportService struct {
	repo *repo.AllRepos
}

// NewPDFExportService creates a new PDF export service instance.
func NewPDFExportService(repos *repo.AllRepos) *PDFExportService {
	return &PDFExportService{repo: repos}
}

// getTheme resolves the theme by name, defaulting to "navy" if not found.
func getTheme(name string) PDFTheme {
	if t, ok := PDFThemes[strings.ToLower(name)]; ok {
		return t
	}
	return PDFThemes["navy"]
}

// ExportSingleItem generates a PDF report for a single item.
// Returns the PDF bytes and a suggested filename.
func (svc *PDFExportService) ExportSingleItem(
	ctx context.Context, groupID uuid.UUID, itemID uuid.UUID, opts PDFExportOptions,
) ([]byte, string, error) {
	// Fetch the full item details including attachments, fields, location, tags
	item, err := svc.repo.Entities.GetOneByGroup(ctx, groupID, itemID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get item: %w", err)
	}

	// Fetch maintenance log for this item (all entries, no filter)
	maintenance, err := svc.repo.MaintEntry.GetMaintenanceByItemID(ctx, groupID, itemID, repo.MaintenanceFilters{})
	if err != nil {
		log.Warn().Err(err).Msg("failed to get maintenance entries for PDF export, continuing without")
		maintenance = nil
	}

	theme := getTheme(opts.Theme)
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// Generate the single-item cover page
	svc.addCoverPage(pdf, theme, opts, []repo.EntityOut{item})

	// Generate the detailed item page(s)
	svc.addItemPages(ctx, pdf, theme, opts, item, maintenance)

	// Write PDF to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Build filename with asset ID when available, otherwise just the date
	date := time.Now().Format("2006-01-02")
	var filename string
	if !item.AssetID.Nil() {
		filename = fmt.Sprintf("HomeBox Asset Export - %s - %s.pdf", item.AssetID.String(), date)
	} else {
		filename = fmt.Sprintf("HomeBox Asset Export - %s.pdf", date)
	}

	return buf.Bytes(), filename, nil
}

// ExportMultipleItems generates a PDF report containing multiple items.
// Includes a summary page with a table of all items followed by per-item detail pages.
func (svc *PDFExportService) ExportMultipleItems(
	ctx context.Context, groupID uuid.UUID, itemIDs []uuid.UUID, opts PDFExportOptions,
) ([]byte, string, error) {
	// Fetch all requested items individually.
	// NOTE: This is an N+1 query pattern. A batch-fetch method (e.g., GetManyByGroup)
	// would be more efficient but does not currently exist in the repository layer.
	// This is acceptable for typical export sizes but should be optimized if exports
	// of hundreds of items become common.
	var items []repo.EntityOut
	for _, id := range itemIDs {
		item, err := svc.repo.Entities.GetOneByGroup(ctx, groupID, id)
		if err != nil {
			log.Warn().Err(err).Str("itemID", id.String()).Msg("skipping item in PDF export")
			continue
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, "", fmt.Errorf("no valid items found for export")
	}

	theme := getTheme(opts.Theme)
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// Cover page
	svc.addCoverPage(pdf, theme, opts, items)

	// Summary page with item table (only for multi-item exports)
	svc.addSummaryPage(pdf, theme, items)

	// Per-item detail pages
	for _, item := range items {
		maintenance, err := svc.repo.MaintEntry.GetMaintenanceByItemID(ctx, groupID, item.ID, repo.MaintenanceFilters{})
		if err != nil {
			log.Warn().Err(err).Msg("failed to get maintenance for item, continuing")
			maintenance = nil
		}
		svc.addItemPages(ctx, pdf, theme, opts, item, maintenance)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("HomeBox Asset Export - %s.pdf", date)

	return buf.Bytes(), filename, nil
}

// addCoverPage renders the title/cover page of the PDF report.
// Shows the title, owner name, generation date, and purpose statement.
func (svc *PDFExportService) addCoverPage(pdf *fpdf.Fpdf, theme PDFTheme, opts PDFExportOptions, items []repo.EntityOut) {
	pdf.AddPage()

	pageW, pageH := pdf.GetPageSize()

	// Large colored header band across the top of the cover page
	pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
	pdf.Rect(0, 0, pageW, 80, "F")

	// Title text centered in the header band
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", theme.CoverTitleSize)
	pdf.SetY(25)
	pdf.CellFormat(pageW, 12, "HomeBox Asset Report", "", 1, "C", false, 0, "")

	// Subtitle with item count
	pdf.SetFont("Helvetica", "", theme.CoverDetailSize)
	subtitle := "Single Item Report"
	if len(items) > 1 {
		subtitle = fmt.Sprintf("%d Items", len(items))
	}
	pdf.CellFormat(pageW, 10, subtitle, "", 1, "C", false, 0, "")

	// Accent divider line below the header
	pdf.SetDrawColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.SetLineWidth(1.5)
	pdf.Line(40, 85, pageW-40, 85)

	// Owner name and date centered below the divider
	pdf.SetTextColor(60, 60, 60)
	pdf.SetFont("Helvetica", "", theme.CoverDetailSize)
	yPos := 100.0

	if opts.OwnerName != "" {
		pdf.SetY(yPos)
		pdf.CellFormat(pageW, 10, fmt.Sprintf("Prepared for: %s", opts.OwnerName), "", 1, "C", false, 0, "")
		yPos += 12
	}

	pdf.SetY(yPos)
	pdf.CellFormat(pageW, 10, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006")), "", 1, "C", false, 0, "")
	yPos += 20

	// Purpose statement for insurance documentation
	pdf.SetY(yPos)
	pdf.SetFont("Helvetica", "I", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(pageW, 10, "For Insurance & Documentation Purposes", "", 1, "C", false, 0, "")

	// Calculate and show total estimated value across all items
	totalValue := 0.0
	insuredCount := 0
	for _, item := range items {
		totalValue += item.PurchasePrice
		if item.Insured {
			insuredCount++
		}
	}

	if totalValue > 0 {
		pdf.SetY(yPos + 20)
		pdf.SetFont("Helvetica", "B", 16)
		pdf.SetTextColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
		pdf.CellFormat(pageW, 10, fmt.Sprintf("Total Estimated Value: $%.2f", totalValue), "", 1, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 12)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(pageW, 8, fmt.Sprintf("%d of %d items insured", insuredCount, len(items)), "", 1, "C", false, 0, "")
	}

	// Footer branding at the bottom of the cover page
	pdf.SetY(pageH - 30)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(pageW, 6, "Generated by HomeBox — Home Inventory Management", "", 1, "C", false, 0, "")
}

// addSummaryPage renders a table-based summary of all items in a multi-item export.
// Columns: Asset ID, Name, Location, Value, Insured status.
func (svc *PDFExportService) addSummaryPage(pdf *fpdf.Fpdf, theme PDFTheme, items []repo.EntityOut) {
	pdf.AddPage()

	// Section header
	svc.drawSectionHeader(pdf, theme, "Item Summary")

	// Table header row
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
	pdf.SetTextColor(255, 255, 255)

	// Column widths proportional to page width (with margins)
	colWidths := []float64{25, 65, 45, 30, 25}
	headers := []string{"Asset ID", "Name", "Location", "Value", "Insured"}
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data rows with alternating background colors
	pdf.SetFont("Helvetica", "", 8)
	for rowIdx, item := range items {
		// Check if we need a new page (leave room for footer)
		if pdf.GetY() > 260 {
			pdf.AddPage()
			// Re-draw table header on new page
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
			pdf.SetTextColor(255, 255, 255)
			for i, h := range headers {
				pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Helvetica", "", 8)
		}

		// Apply alternating row shading for readability
		if rowIdx%2 == 1 {
			pdf.SetFillColor(theme.AltRowR, theme.AltRowG, theme.AltRowB)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(50, 50, 50)

		locationName := ""
		if item.Location != nil {
			locationName = item.Location.Name
		}

		insuredStr := "No"
		if item.Insured {
			insuredStr = "Yes"
		}

		pdf.CellFormat(colWidths[0], 7, item.AssetID.String(), "1", 0, "C", true, 0, "")
		pdf.CellFormat(colWidths[1], 7, truncateStr(item.Name, 35), "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[2], 7, truncateStr(locationName, 22), "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[3], 7, fmt.Sprintf("$%.2f", item.PurchasePrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[4], 7, insuredStr, "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
	}

	// Calculate totals for the summary footer row
	total := 0.0
	insured := 0
	for _, item := range items {
		total += item.PurchasePrice
		if item.Insured {
			insured++
		}
	}

	// Summary totals row at the bottom of the table
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(colWidths[0]+colWidths[1]+colWidths[2], 8, fmt.Sprintf("Total: %d items", len(items)), "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidths[3], 8, fmt.Sprintf("$%.2f", total), "1", 0, "R", true, 0, "")
	pdf.CellFormat(colWidths[4], 8, fmt.Sprintf("%d", insured), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)
}

// addItemPages renders one or more pages of detailed information for a single item.
// Includes: header bar, primary photo, details, purchase/warranty/sold info,
// custom fields, notes, additional photos, receipts, and maintenance history.
func (svc *PDFExportService) addItemPages(
	ctx context.Context, pdf *fpdf.Fpdf, theme PDFTheme, opts PDFExportOptions,
	item repo.EntityOut, maintenance []repo.MaintenanceEntryWithDetails,
) {
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	marginL := 10.0
	contentW := pageW - 2*marginL

	svc.addItemHeaderBar(pdf, theme, item, marginL, contentW)

	pdf.SetY(30)
	photoY := pdf.GetY()
	photoEmbedded := false
	if opts.IncludePhotos {
		photoEmbedded = svc.addItemPrimaryPhoto(ctx, pdf, item, marginL, photoY)
	}

	svc.addItemDetailRows(pdf, theme, item, marginL, contentW, photoY, photoEmbedded)
	svc.addItemDescription(pdf, theme, item, marginL, contentW)
	svc.addItemPurchaseInfo(pdf, theme, item, marginL, contentW)
	svc.addItemWarranty(pdf, theme, item, marginL, contentW)
	svc.addItemCustomFields(pdf, theme, item, marginL, contentW)
	svc.addItemNotes(pdf, theme, item, marginL, contentW)

	if opts.IncludePhotos {
		svc.addItemPhotoAppendix(ctx, pdf, theme, item, marginL, contentW)
	}

	svc.addMaintenanceHistory(pdf, theme, maintenance, marginL)
	svc.addItemSoldInfo(pdf, theme, item, marginL, contentW)
}

// addItemHeaderBar renders the colored banner with the item name and asset ID.
func (svc *PDFExportService) addItemHeaderBar(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	pageW, _ := pdf.GetPageSize()
	pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
	pdf.Rect(0, 10, pageW, 16, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", theme.HeaderFontSize)
	pdf.SetY(12)
	pdf.SetX(marginL)
	headerText := item.Name
	if !item.AssetID.Nil() {
		headerText = fmt.Sprintf("%s  |  Asset ID: %s", item.Name, item.AssetID.String())
	}
	pdf.CellFormat(contentW, 12, headerText, "", 1, "L", false, 0, "")
}

// addItemPrimaryPhoto embeds the primary photo attachment on the right side of
// the page. Returns true when a photo was embedded.
func (svc *PDFExportService) addItemPrimaryPhoto(ctx context.Context, pdf *fpdf.Fpdf, item repo.EntityOut, marginL, photoY float64) bool {
	pageW, _ := pdf.GetPageSize()
	for _, att := range item.Attachments {
		if !att.Primary || att.Type != attachment.TypePhoto.String() {
			continue
		}
		imgBytes, imgType, err := svc.readAttachment(ctx, att)
		if err != nil {
			log.Warn().Err(err).Msg("failed to read primary photo for PDF")
			return false
		}
		// Skip images that exceed the size limit to prevent excessive memory usage
		if len(imgBytes) > MaxImageBytes {
			log.Warn().Int("bytes", len(imgBytes)).Msg("primary photo exceeds max image size, skipping embed")
			return false
		}
		// Register the image and place it on the right side, 60mm wide, proportional height
		imgName := fmt.Sprintf("primary_%s", att.ID.String())
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: imgType}, bytes.NewReader(imgBytes))
		imgW := 60.0
		pdf.ImageOptions(imgName, pageW-marginL-imgW, photoY, imgW, 0, false, fpdf.ImageOptions{ImageType: imgType}, 0, "")
		return true
	}
	return false
}

// addItemDetailRows renders the label-value detail rows (location, quantity,
// identification, insured status, tags) in the left column next to the photo.
func (svc *PDFExportService) addItemDetailRows(
	pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut,
	marginL, contentW, photoY float64, photoEmbedded bool,
) {
	detailW := contentW
	if photoEmbedded {
		detailW = contentW - 65 // Leave room for the photo on the right
	}

	pdf.SetY(photoY)
	pdf.SetX(marginL)

	if item.Location != nil {
		svc.drawDetailRow(pdf, theme, marginL, detailW, "Location", item.Location.Name)
	}

	svc.drawDetailRow(pdf, theme, marginL, detailW, "Quantity", fmt.Sprintf("%g", item.Quantity))

	if item.SerialNumber != "" {
		svc.drawDetailRow(pdf, theme, marginL, detailW, "Serial Number", item.SerialNumber)
	}
	if item.ModelNumber != "" {
		svc.drawDetailRow(pdf, theme, marginL, detailW, "Model Number", item.ModelNumber)
	}
	if item.Manufacturer != "" {
		svc.drawDetailRow(pdf, theme, marginL, detailW, "Manufacturer", item.Manufacturer)
	}

	insuredStr := "No"
	if item.Insured {
		insuredStr = "Yes"
	}
	svc.drawDetailRow(pdf, theme, marginL, detailW, "Insured", insuredStr)

	if len(item.Tags) > 0 {
		tagNames := make([]string, len(item.Tags))
		for i, t := range item.Tags {
			tagNames[i] = t.Name
		}
		svc.drawDetailRow(pdf, theme, marginL, detailW, "Tags", strings.Join(tagNames, ", "))
	}

	// Ensure we move below the photo before continuing
	if photoEmbedded && pdf.GetY() < photoY+65 {
		pdf.SetY(photoY + 65)
	}
}

// addItemDescription renders the description section when present.
func (svc *PDFExportService) addItemDescription(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if item.Description == "" {
		return
	}
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Description")
	pdf.SetFont("Helvetica", "", theme.BodyFontSize)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetX(marginL)
	pdf.MultiCell(contentW, 5, item.Description, "", "L", false)
}

// addItemPurchaseInfo renders the purchase information section when present.
func (svc *PDFExportService) addItemPurchaseInfo(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if item.PurchaseFrom == "" && item.PurchasePrice <= 0 && item.PurchaseDate.Time().IsZero() {
		return
	}
	svc.ensureSpace(pdf, 30)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Purchase Information")

	if item.PurchaseFrom != "" {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Purchased From", item.PurchaseFrom)
	}
	if !item.PurchaseDate.Time().IsZero() {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Purchase Date", item.PurchaseDate.Time().Format("January 2, 2006"))
	}
	if item.PurchasePrice > 0 {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Purchase Price", fmt.Sprintf("$%.2f", item.PurchasePrice))
	}
}

// addItemWarranty renders the warranty section when present.
func (svc *PDFExportService) addItemWarranty(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if !item.LifetimeWarranty && item.WarrantyExpires.Time().IsZero() && item.WarrantyDetails == "" {
		return
	}
	svc.ensureSpace(pdf, 25)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Warranty")

	if item.LifetimeWarranty {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Warranty", "Lifetime")
	} else if !item.WarrantyExpires.Time().IsZero() {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Warranty Expires", item.WarrantyExpires.Time().Format("January 2, 2006"))
	}
	if item.WarrantyDetails != "" {
		pdf.SetFont("Helvetica", "", theme.BodyFontSize)
		pdf.SetTextColor(50, 50, 50)
		pdf.SetX(marginL)
		pdf.MultiCell(contentW, 5, item.WarrantyDetails, "", "L", false)
	}
}

// addItemCustomFields renders the custom fields section when present.
func (svc *PDFExportService) addItemCustomFields(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if len(item.Fields) == 0 {
		return
	}
	svc.ensureSpace(pdf, 20)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Custom Fields")

	for _, field := range item.Fields {
		var value string
		switch field.Type {
		case "text":
			value = field.TextValue
		case "number":
			value = fmt.Sprintf("%d", field.NumberValue)
		case "boolean":
			if field.BooleanValue {
				value = "Yes"
			} else {
				value = "No"
			}
		default:
			value = field.TextValue
		}
		if value != "" {
			svc.drawDetailRow(pdf, theme, marginL, contentW, field.Name, value)
		}
	}
}

// addItemNotes renders the notes section when present, stripped of Markdown.
func (svc *PDFExportService) addItemNotes(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if item.Notes == "" {
		return
	}
	svc.ensureSpace(pdf, 20)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Notes")
	pdf.SetFont("Helvetica", "", theme.BodyFontSize)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetX(marginL)
	pdf.MultiCell(contentW, 5, stripMarkdown(item.Notes), "", "L", false)
}

// addItemPhotoAppendix renders additional (non-primary) photos and receipt
// images in grid layouts on their own pages.
func (svc *PDFExportService) addItemPhotoAppendix(
	ctx context.Context, pdf *fpdf.Fpdf, theme PDFTheme,
	item repo.EntityOut, marginL, contentW float64,
) {
	var additionalPhotos []repo.ItemAttachment
	var receipts []repo.ItemAttachment
	for _, att := range item.Attachments {
		switch {
		case att.Type == attachment.TypePhoto.String() && !att.Primary:
			// All photo attachments except the primary one (already shown above)
			additionalPhotos = append(additionalPhotos, att)
		case att.Type == attachment.TypeReceipt.String():
			receipts = append(receipts, att)
		}
	}

	if len(additionalPhotos) > 0 {
		pdf.AddPage()
		svc.drawSectionHeader(pdf, theme, "Additional Photos")
		svc.drawPhotoGrid(ctx, pdf, marginL, contentW, additionalPhotos)
	}

	if len(receipts) > 0 {
		pdf.AddPage()
		svc.drawSectionHeader(pdf, theme, "Receipts")
		svc.drawPhotoGrid(ctx, pdf, marginL, contentW, receipts)
	}
}

// addMaintenanceHistory renders the maintenance log as a table.
func (svc *PDFExportService) addMaintenanceHistory(
	pdf *fpdf.Fpdf, theme PDFTheme,
	maintenance []repo.MaintenanceEntryWithDetails, marginL float64,
) {
	if len(maintenance) == 0 {
		return
	}
	svc.ensureSpace(pdf, 30)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Maintenance History")

	maintCols := []float64{40, 50, 30, 30, 25}
	maintHeaders := []string{"Task", "Description", "Scheduled", "Completed", "Cost"}

	drawTableHeader := func() {
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetFillColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetX(marginL)
		for i, h := range maintHeaders {
			pdf.CellFormat(maintCols[i], 7, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 7)
	}
	drawTableHeader()

	for rowIdx, entry := range maintenance {
		if pdf.GetY() > 260 {
			pdf.AddPage()
			drawTableHeader()
		}

		if rowIdx%2 == 1 {
			pdf.SetFillColor(theme.AltRowR, theme.AltRowG, theme.AltRowB)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(50, 50, 50)

		scheduled := ""
		if !entry.ScheduledDate.Time().IsZero() {
			scheduled = entry.ScheduledDate.Time().Format("2006-01-02")
		}
		completed := ""
		if !entry.CompletedDate.Time().IsZero() {
			completed = entry.CompletedDate.Time().Format("2006-01-02")
		}
		costStr := ""
		if entry.Cost > 0 {
			costStr = fmt.Sprintf("$%.2f", entry.Cost)
		}

		pdf.SetX(marginL)
		pdf.CellFormat(maintCols[0], 6, truncateStr(entry.Name, 22), "1", 0, "L", true, 0, "")
		pdf.CellFormat(maintCols[1], 6, truncateStr(entry.Description, 28), "1", 0, "L", true, 0, "")
		pdf.CellFormat(maintCols[2], 6, scheduled, "1", 0, "C", true, 0, "")
		pdf.CellFormat(maintCols[3], 6, completed, "1", 0, "C", true, 0, "")
		pdf.CellFormat(maintCols[4], 6, costStr, "1", 0, "R", true, 0, "")
		pdf.Ln(-1)
	}
}

// addItemSoldInfo renders the sold information section when the item has been sold.
func (svc *PDFExportService) addItemSoldInfo(pdf *fpdf.Fpdf, theme PDFTheme, item repo.EntityOut, marginL, contentW float64) {
	if item.SoldTo == "" && item.SoldPrice <= 0 && item.SoldDate.Time().IsZero() {
		return
	}
	svc.ensureSpace(pdf, 25)
	pdf.Ln(4)
	svc.drawSectionHeader(pdf, theme, "Sold Information")

	if item.SoldTo != "" {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Sold To", item.SoldTo)
	}
	if !item.SoldDate.Time().IsZero() {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Sold Date", item.SoldDate.Time().Format("January 2, 2006"))
	}
	if item.SoldPrice > 0 {
		svc.drawDetailRow(pdf, theme, marginL, contentW, "Sold Price", fmt.Sprintf("$%.2f", item.SoldPrice))
	}
	if item.SoldNotes != "" {
		pdf.SetFont("Helvetica", "", theme.BodyFontSize)
		pdf.SetTextColor(50, 50, 50)
		pdf.SetX(marginL)
		pdf.MultiCell(contentW, 5, item.SoldNotes, "", "L", false)
	}
}

// drawSectionHeader renders a themed section title with an accent underline.
func (svc *PDFExportService) drawSectionHeader(pdf *fpdf.Fpdf, theme PDFTheme, title string) {
	marginL := 10.0
	pageW, _ := pdf.GetPageSize()
	contentW := pageW - 2*marginL

	pdf.SetFont("Helvetica", "B", theme.HeaderFontSize-2)
	pdf.SetTextColor(theme.HeaderR, theme.HeaderG, theme.HeaderB)
	pdf.SetX(marginL)
	pdf.CellFormat(contentW, 8, title, "", 1, "L", false, 0, "")

	// Draw a colored accent line under the section title
	pdf.SetDrawColor(theme.AccentR, theme.AccentG, theme.AccentB)
	pdf.SetLineWidth(0.5)
	y := pdf.GetY()
	pdf.Line(marginL, y, marginL+contentW, y)
	pdf.Ln(3)
}

// drawDetailRow renders a single label-value pair as a row.
// The label is bold and the value is normal weight.
func (svc *PDFExportService) drawDetailRow(pdf *fpdf.Fpdf, theme PDFTheme, marginL, width float64, label, value string) {
	labelW := 45.0
	valueW := width - labelW

	pdf.SetX(marginL)
	pdf.SetFont("Helvetica", "B", theme.BodyFontSize)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(labelW, 6, label+":", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", theme.BodyFontSize)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(valueW, 6, value, "", 1, "L", false, 0, "")
}

// drawPhotoGrid renders a set of images in a 2-column grid layout.
// Each image is scaled to fit within a cell while maintaining aspect ratio.
func (svc *PDFExportService) drawPhotoGrid(
	ctx context.Context, pdf *fpdf.Fpdf,
	marginL, contentW float64, attachments []repo.ItemAttachment,
) {
	colW := (contentW - 5) / 2 // 5mm gap between columns
	imgH := 70.0               // Fixed height for grid cells

	for i, att := range attachments {
		// Start a new row every 2 images
		col := i % 2
		if col == 0 && i > 0 {
			pdf.Ln(imgH + 5)
		}
		if col == 0 && pdf.GetY()+imgH > 270 {
			pdf.AddPage()
		}

		imgBytes, imgType, err := svc.readAttachment(ctx, att)
		if err != nil {
			log.Warn().Err(err).Str("attachment", att.ID.String()).Msg("failed to read attachment for photo grid")
			continue
		}
		// Skip images that exceed the size limit to prevent excessive memory usage
		if len(imgBytes) > MaxImageBytes {
			log.Warn().Str("attachment", att.ID.String()).Int("bytes", len(imgBytes)).Msg("image exceeds max size, skipping")
			continue
		}

		imgName := fmt.Sprintf("grid_%s", att.ID.String())
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: imgType}, bytes.NewReader(imgBytes))

		x := marginL + float64(col)*(colW+5)
		y := pdf.GetY()

		pdf.ImageOptions(imgName, x, y, colW, 0, false, fpdf.ImageOptions{ImageType: imgType}, 0, "")

		// Draw a light border around the image
		pdf.SetDrawColor(200, 200, 200)
		pdf.SetLineWidth(0.3)
		pdf.Rect(x, y, colW, imgH, "D")

		// Caption below the image with the attachment title
		if att.Title != "" {
			pdf.SetFont("Helvetica", "I", 7)
			pdf.SetTextColor(120, 120, 120)
			pdf.Text(x+2, y+imgH-2, truncateStr(att.Title, 40))
		}
	}

	pdf.Ln(imgH + 5)
}

// readAttachment reads the binary content of an attachment from blob storage.
// Returns the bytes, detected image type (for fpdf), and any error.
func (svc *PDFExportService) readAttachment(ctx context.Context, att repo.ItemAttachment) ([]byte, string, error) {
	// Defensive path traversal check: ensure the attachment path does not
	// contain ".." components that could escape the expected storage directory.
	cleanPath := filepath.ToSlash(filepath.Clean(att.Path))
	if strings.Contains(cleanPath, "..") {
		return nil, "", fmt.Errorf("invalid attachment path (directory traversal detected): %s", att.Path)
	}

	// Open the blob storage bucket using the configured connection string
	bucket, err := blob.OpenBucket(ctx, svc.repo.Attachments.GetConnString())
	if err != nil {
		return nil, "", fmt.Errorf("failed to open bucket: %w", err)
	}
	defer func() {
		if cerr := bucket.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("failed to close blob bucket after attachment read")
		}
	}()

	// Read the full file from storage
	reader, err := bucket.NewReader(ctx, svc.repo.Attachments.GetFullPath(att.Path), nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read attachment: %w", err)
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			log.Warn().Err(cerr).Str("attachment", att.ID.String()).Msg("failed to close attachment reader")
		}
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read attachment data: %w", err)
	}

	// Determine image type from MIME type for fpdf registration.
	// fpdf natively supports jpg, png, and gif. For unsupported formats
	// (webp, heic, avif), we log a warning and skip the image since fpdf
	// cannot embed them without a conversion step.
	mime := strings.ToLower(att.MimeType)
	var imgType string
	switch {
	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		imgType = "jpg"
	case strings.Contains(mime, "png"):
		imgType = "png"
	case strings.Contains(mime, "gif"):
		imgType = "gif"
	case strings.Contains(mime, "webp"),
		strings.Contains(mime, "heic"), strings.Contains(mime, "heif"),
		strings.Contains(mime, "avif"):
		log.Warn().Str("mimeType", att.MimeType).Str("attachmentID", att.ID.String()).
			Msg("unsupported image format for PDF embed, skipping")
		return nil, "", fmt.Errorf("unsupported image format for PDF: %s", att.MimeType)
	default:
		log.Warn().Str("mimeType", att.MimeType).Str("attachmentID", att.ID.String()).
			Msg("unknown MIME type for PDF image embed, defaulting to jpg")
		imgType = "jpg"
	}

	return data, imgType, nil
}

// ensureSpace checks if enough vertical space remains on the current page.
// If not, it starts a new page to prevent content from being cut off.
func (svc *PDFExportService) ensureSpace(pdf *fpdf.Fpdf, minSpace float64) {
	_, pageH := pdf.GetPageSize()
	if pdf.GetY()+minSpace > pageH-20 {
		pdf.AddPage()
	}
}

// truncateStr shortens a string to maxLen runes, appending "..." if truncated.
// Uses rune-based slicing to safely handle multi-byte UTF-8 characters.
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// stripMarkdown removes common Markdown formatting for plain-text rendering in PDFs.
// Targets bold/italic markers, header prefixes, and code backticks while preserving
// underscores that appear in identifiers (e.g., serial_number).
func stripMarkdown(s string) string {
	// Remove bold/italic markers (must remove ** before * to avoid partial matches)
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")

	// Remove header markers from line starts (e.g., "## Title" -> "Title")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		stripped := strings.TrimLeft(line, "#")
		if stripped != line {
			lines[i] = strings.TrimLeft(stripped, " ")
		}
	}
	s = strings.Join(lines, "\n")

	// Remove code backticks (triple first, then single)
	s = strings.ReplaceAll(s, "```", "")
	s = strings.ReplaceAll(s, "`", "")

	return s
}
