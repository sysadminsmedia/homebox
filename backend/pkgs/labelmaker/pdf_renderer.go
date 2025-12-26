package labelmaker

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"codeberg.org/go-pdf/fpdf"
	"github.com/rs/zerolog/log"
)

// PDFPageSize represents common paper sizes
type PDFPageSize string

const (
	PDFPageLetter PDFPageSize = "Letter"
	PDFPageA4     PDFPageSize = "A4"
	PDFPageCustom PDFPageSize = "Custom"
)

// PDFRenderOptions configures PDF output
type PDFRenderOptions struct {
	PageSize      PDFPageSize
	PageWidth     float64 // mm (for custom size)
	PageHeight    float64 // mm (for custom size)
	Orientation   string  // "P" for portrait, "L" for landscape
	MarginTop     float64 // mm
	MarginBottom  float64 // mm
	MarginLeft    float64 // mm
	MarginRight   float64 // mm
	LabelSpacingX float64 // mm spacing between labels horizontally
	LabelSpacingY float64 // mm spacing between labels vertically
	LabelsPerRow  int     // 0 = auto-calculate
	LabelsPerCol  int     // 0 = auto-calculate
	ShowCutGuides bool    // Draw light borders around labels for cutting (disable for pre-cut sheets)
}

// DefaultPDFOptions returns sensible defaults for PDF generation
func DefaultPDFOptions() PDFRenderOptions {
	return PDFRenderOptions{
		PageSize:      PDFPageLetter,
		Orientation:   "P",
		MarginTop:     10,
		MarginBottom:  10,
		MarginLeft:    10,
		MarginRight:   10,
		LabelSpacingX: 5,
		LabelSpacingY: 5,
		LabelsPerRow:  0, // auto
		LabelsPerCol:  0, // auto
	}
}

// getPageDimensions returns page width and height in mm
func getPageDimensions(opts PDFRenderOptions) (width, height float64) {
	switch opts.PageSize {
	case PDFPageLetter:
		width, height = 215.9, 279.4 // 8.5 x 11 inches in mm
	case PDFPageA4:
		width, height = 210, 297
	case PDFPageCustom:
		width, height = opts.PageWidth, opts.PageHeight
	default:
		width, height = 215.9, 279.4 // default to letter
	}

	if opts.Orientation == "L" {
		width, height = height, width
	}

	return width, height
}

// RenderToPDFWithSheetLayout renders labels to PDF using Avery-style sheet layout if available
func (r *TemplateRenderer) RenderToPDFWithSheetLayout(template *TemplateData, items []*ItemData, opts PDFRenderOptions, sheetLayout *SheetLayout) ([]byte, error) {
	if sheetLayout != nil {
		return r.renderPDFWithSheetLayout(template, items, opts, sheetLayout)
	}
	return r.RenderToPDF(template, items, opts)
}

// renderPDFWithSheetLayout renders labels using exact Avery-style sheet layout
func (r *TemplateRenderer) renderPDFWithSheetLayout(template *TemplateData, items []*ItemData, opts PDFRenderOptions, layout *SheetLayout) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}
	if layout.Columns <= 0 || layout.Rows <= 0 {
		return nil, fmt.Errorf("sheet layout must have positive columns (%d) and rows (%d)", layout.Columns, layout.Rows)
	}

	// Create PDF with custom page size matching the sheet
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: opts.Orientation,
		UnitStr:        "mm",
		SizeStr:        "",
		Size:           fpdf.SizeType{Wd: layout.PageWidth, Ht: layout.PageHeight},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	labelsPerPage := layout.Columns * layout.Rows
	labelWidth := template.Width
	labelHeight := template.Height

	// Render labels page by page
	skipped := 0
	for i, item := range items {
		// Add new page if needed
		if i%labelsPerPage == 0 {
			pdf.AddPage()
		}

		// Calculate position on page
		posInPage := i % labelsPerPage
		col := posInPage % layout.Columns
		row := posInPage / layout.Columns

		x := layout.MarginLeft + float64(col)*(labelWidth+layout.GutterH)
		y := layout.MarginTop + float64(row)*(labelHeight+layout.GutterV)

		// Render label to image
		imgData, err := r.RenderTemplate(RenderContext{
			Item:     item,
			Template: template,
		})
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("failed to render label template, skipping")
			skipped++
			continue
		}

		// Decode PNG
		img, err := png.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("failed to decode rendered PNG, skipping")
			skipped++
			continue
		}

		// Register image with PDF
		imgName := fmt.Sprintf("label_%d", i)
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgData))

		// Place image on PDF
		pdf.ImageOptions(imgName, x, y, labelWidth, labelHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		_ = img
	}

	if skipped > 0 {
		log.Warn().Int("skipped", skipped).Int("total", len(items)).Msg("some labels were skipped during PDF generation")
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderLocationsToPDFWithSheetLayout renders location labels to PDF using Avery-style sheet layout if available
func (r *TemplateRenderer) RenderLocationsToPDFWithSheetLayout(template *TemplateData, locations []*LocationData, opts PDFRenderOptions, sheetLayout *SheetLayout) ([]byte, error) {
	if sheetLayout != nil {
		return r.renderLocationsPDFWithSheetLayout(template, locations, opts, sheetLayout)
	}
	return r.RenderLocationsToPDF(template, locations, opts)
}

// renderLocationsPDFWithSheetLayout renders location labels using exact Avery-style sheet layout
func (r *TemplateRenderer) renderLocationsPDFWithSheetLayout(template *TemplateData, locations []*LocationData, opts PDFRenderOptions, layout *SheetLayout) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}
	if len(locations) == 0 {
		return nil, fmt.Errorf("at least one location is required")
	}
	if layout.Columns <= 0 || layout.Rows <= 0 {
		return nil, fmt.Errorf("sheet layout must have positive columns (%d) and rows (%d)", layout.Columns, layout.Rows)
	}

	// Create PDF with custom page size matching the sheet
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: opts.Orientation,
		UnitStr:        "mm",
		SizeStr:        "",
		Size:           fpdf.SizeType{Wd: layout.PageWidth, Ht: layout.PageHeight},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	labelsPerPage := layout.Columns * layout.Rows
	labelWidth := template.Width
	labelHeight := template.Height

	// Render labels page by page
	skipped := 0
	for i, location := range locations {
		// Add new page if needed
		if i%labelsPerPage == 0 {
			pdf.AddPage()
		}

		// Calculate position on page
		posInPage := i % labelsPerPage
		col := posInPage % layout.Columns
		row := posInPage / layout.Columns

		x := layout.MarginLeft + float64(col)*(labelWidth+layout.GutterH)
		y := layout.MarginTop + float64(row)*(labelHeight+layout.GutterV)

		// Render label to image
		imgData, err := r.RenderTemplate(RenderContext{
			Location: location,
			Template: template,
		})
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("failed to render location label template, skipping")
			skipped++
			continue
		}

		// Decode PNG
		img, err := png.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("failed to decode rendered PNG, skipping")
			skipped++
			continue
		}

		// Register image with PDF
		imgName := fmt.Sprintf("location_label_%d", i)
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgData))

		// Place image on PDF
		pdf.ImageOptions(imgName, x, y, labelWidth, labelHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		_ = img
	}

	if skipped > 0 {
		log.Warn().Int("skipped", skipped).Int("total", len(locations)).Msg("some location labels were skipped during PDF generation")
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderLocationsToPDF renders multiple location labels to a multi-page PDF
func (r *TemplateRenderer) RenderLocationsToPDF(template *TemplateData, locations []*LocationData, opts PDFRenderOptions) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("at least one location is required")
	}

	// Get page dimensions
	pageWidth, pageHeight := getPageDimensions(opts)

	// Calculate usable area
	usableWidth := pageWidth - opts.MarginLeft - opts.MarginRight
	usableHeight := pageHeight - opts.MarginTop - opts.MarginBottom

	// Label dimensions
	labelWidth := template.Width
	labelHeight := template.Height

	// Calculate labels per row/column if not specified
	labelsPerRow := opts.LabelsPerRow
	labelsPerCol := opts.LabelsPerCol

	if labelsPerRow <= 0 {
		labelsPerRow = int((usableWidth + opts.LabelSpacingX) / (labelWidth + opts.LabelSpacingX))
		if labelsPerRow <= 0 {
			labelsPerRow = 1
		}
	}

	if labelsPerCol <= 0 {
		labelsPerCol = int((usableHeight + opts.LabelSpacingY) / (labelHeight + opts.LabelSpacingY))
		if labelsPerCol <= 0 {
			labelsPerCol = 1
		}
	}

	labelsPerPage := labelsPerRow * labelsPerCol

	// Create PDF
	pdf := fpdf.New(opts.Orientation, "mm", string(opts.PageSize), "")
	pdf.SetMargins(opts.MarginLeft, opts.MarginTop, opts.MarginRight)
	pdf.SetAutoPageBreak(false, opts.MarginBottom)

	// Render each label as an image and place on PDF
	for i, location := range locations {
		// Add new page if needed
		if i%labelsPerPage == 0 {
			pdf.AddPage()
		}

		// Calculate position on page
		posInPage := i % labelsPerPage
		col := posInPage % labelsPerRow
		row := posInPage / labelsPerRow

		x := opts.MarginLeft + float64(col)*(labelWidth+opts.LabelSpacingX)
		y := opts.MarginTop + float64(row)*(labelHeight+opts.LabelSpacingY)

		// Render label to image
		imgData, err := r.RenderTemplate(RenderContext{
			Location: location,
			Template: template,
		})
		if err != nil {
			// Skip this label on error but continue with others
			continue
		}

		// Decode PNG
		img, err := png.Decode(bytes.NewReader(imgData))
		if err != nil {
			continue
		}

		// Register image with PDF
		imgName := fmt.Sprintf("location_label_%d", i)
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgData))

		// Place image on PDF
		pdf.ImageOptions(imgName, x, y, labelWidth, labelHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		// Draw a light border around the label for cutting guides (optional)
		if opts.ShowCutGuides {
			pdf.SetDrawColor(200, 200, 200)
			pdf.Rect(x, y, labelWidth, labelHeight, "D")
		}

		_ = img
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderToPDF renders multiple labels to a multi-page PDF
func (r *TemplateRenderer) RenderToPDF(template *TemplateData, items []*ItemData, opts PDFRenderOptions) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	// Get page dimensions
	pageWidth, pageHeight := getPageDimensions(opts)

	// Calculate usable area
	usableWidth := pageWidth - opts.MarginLeft - opts.MarginRight
	usableHeight := pageHeight - opts.MarginTop - opts.MarginBottom

	// Label dimensions
	labelWidth := template.Width
	labelHeight := template.Height

	// Calculate labels per row/column if not specified
	labelsPerRow := opts.LabelsPerRow
	labelsPerCol := opts.LabelsPerCol

	if labelsPerRow <= 0 {
		labelsPerRow = int((usableWidth + opts.LabelSpacingX) / (labelWidth + opts.LabelSpacingX))
		if labelsPerRow <= 0 {
			labelsPerRow = 1
		}
	}

	if labelsPerCol <= 0 {
		labelsPerCol = int((usableHeight + opts.LabelSpacingY) / (labelHeight + opts.LabelSpacingY))
		if labelsPerCol <= 0 {
			labelsPerCol = 1
		}
	}

	labelsPerPage := labelsPerRow * labelsPerCol

	// Create PDF
	pdf := fpdf.New(opts.Orientation, "mm", string(opts.PageSize), "")
	pdf.SetMargins(opts.MarginLeft, opts.MarginTop, opts.MarginRight)
	pdf.SetAutoPageBreak(false, opts.MarginBottom)

	// Render each label as an image and place on PDF
	for i, item := range items {
		// Add new page if needed
		if i%labelsPerPage == 0 {
			pdf.AddPage()
		}

		// Calculate position on page
		posInPage := i % labelsPerPage
		col := posInPage % labelsPerRow
		row := posInPage / labelsPerRow

		x := opts.MarginLeft + float64(col)*(labelWidth+opts.LabelSpacingX)
		y := opts.MarginTop + float64(row)*(labelHeight+opts.LabelSpacingY)

		// Render label to image
		imgData, err := r.RenderTemplate(RenderContext{
			Item:     item,
			Template: template,
		})
		if err != nil {
			// Skip this label on error but continue with others
			continue
		}

		// Decode PNG
		img, err := png.Decode(bytes.NewReader(imgData))
		if err != nil {
			continue
		}

		// Register image with PDF
		imgName := fmt.Sprintf("label_%d", i)
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgData))

		// Place image on PDF
		pdf.ImageOptions(imgName, x, y, labelWidth, labelHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		// Draw a light border around the label for cutting guides (optional)
		if opts.ShowCutGuides {
			pdf.SetDrawColor(200, 200, 200)
			pdf.Rect(x, y, labelWidth, labelHeight, "D")
		}

		_ = img // use the decoded image bounds if needed
	}

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderFirstItemToImage renders the first item from a list to an image.
// Note: Despite accepting multiple items, only the first item is rendered.
// This is a simplified implementation; use RenderToPDF for multiple labels.
func (r *TemplateRenderer) RenderFirstItemToImage(template *TemplateData, items []*ItemData) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	return r.RenderTemplate(RenderContext{
		Item:     items[0],
		Template: template,
	})
}

// RenderToSheet renders multiple labels arranged on a sheet (for Avery-style labels)
// Returns PNG data for a single sheet. If more items than fit on a sheet, only renders first sheet.
func (r *TemplateRenderer) RenderToSheet(template *TemplateData, items []*ItemData, layout *SheetLayout) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}
	if layout == nil {
		return nil, fmt.Errorf("sheet layout is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	// Calculate sheet dimensions in pixels (use 300 DPI for high quality)
	sheetDPI := 300
	sheetWidthPx := int(layout.PageWidth * float64(sheetDPI) / 25.4)
	sheetHeightPx := int(layout.PageHeight * float64(sheetDPI) / 25.4)

	// Create a new image for the sheet
	sheetImg := image.NewRGBA(image.Rect(0, 0, sheetWidthPx, sheetHeightPx))

	// Fill with white background
	draw.Draw(sheetImg, sheetImg.Bounds(), image.White, image.Point{}, draw.Src)

	// Calculate label size in pixels
	labelWidthPx := int(template.Width * float64(sheetDPI) / 25.4)
	labelHeightPx := int(template.Height * float64(sheetDPI) / 25.4)

	// Calculate layout positions
	marginTopPx := int(layout.MarginTop * float64(sheetDPI) / 25.4)
	marginLeftPx := int(layout.MarginLeft * float64(sheetDPI) / 25.4)
	gutterHPx := int(layout.GutterH * float64(sheetDPI) / 25.4)
	gutterVPx := int(layout.GutterV * float64(sheetDPI) / 25.4)

	labelsPerPage := layout.Columns * layout.Rows
	numLabels := len(items)
	if numLabels > labelsPerPage {
		numLabels = labelsPerPage
	}

	// Render and place each label
	for i := 0; i < numLabels; i++ {
		col := i % layout.Columns
		row := i / layout.Columns

		// Calculate position
		x := marginLeftPx + col*(labelWidthPx+gutterHPx)
		y := marginTopPx + row*(labelHeightPx+gutterVPx)

		// Render the label
		labelData, err := r.RenderTemplate(RenderContext{
			Item:     items[i],
			Template: template,
		})
		if err != nil {
			continue // Skip labels that fail to render
		}

		// Decode the label PNG
		labelImg, err := png.Decode(bytes.NewReader(labelData))
		if err != nil {
			continue
		}

		// Scale the label to fit the target size if needed
		labelBounds := labelImg.Bounds()
		if labelBounds.Dx() != labelWidthPx || labelBounds.Dy() != labelHeightPx {
			// Simple nearest-neighbor scaling
			scaledImg := image.NewRGBA(image.Rect(0, 0, labelWidthPx, labelHeightPx))
			for py := 0; py < labelHeightPx; py++ {
				for px := 0; px < labelWidthPx; px++ {
					srcX := px * labelBounds.Dx() / labelWidthPx
					srcY := py * labelBounds.Dy() / labelHeightPx
					scaledImg.Set(px, py, labelImg.At(srcX, srcY))
				}
			}
			labelImg = scaledImg
		}

		// Draw label onto sheet
		destRect := image.Rect(x, y, x+labelWidthPx, y+labelHeightPx)
		draw.Draw(sheetImg, destRect, labelImg, labelImg.Bounds().Min, draw.Over)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, sheetImg); err != nil {
		return nil, fmt.Errorf("failed to encode sheet image: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderToSheets renders multiple labels arranged on multiple sheets
// Returns a slice of PNG data, one for each sheet
func (r *TemplateRenderer) RenderToSheets(template *TemplateData, items []*ItemData, layout *SheetLayout) ([][]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is required")
	}
	if layout == nil {
		return nil, fmt.Errorf("sheet layout is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	labelsPerPage := layout.Columns * layout.Rows
	numSheets := (len(items) + labelsPerPage - 1) / labelsPerPage

	sheets := make([][]byte, 0, numSheets)

	for sheet := 0; sheet < numSheets; sheet++ {
		startIdx := sheet * labelsPerPage
		endIdx := startIdx + labelsPerPage
		if endIdx > len(items) {
			endIdx = len(items)
		}

		sheetItems := items[startIdx:endIdx]
		sheetData, err := r.RenderToSheet(template, sheetItems, layout)
		if err != nil {
			return nil, fmt.Errorf("failed to render sheet %d: %w", sheet+1, err)
		}

		sheets = append(sheets, sheetData)
	}

	return sheets, nil
}
