// Package labelmaker provides functionality for generating and printing labels
package labelmaker

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// ItemData contains all the data fields available for label rendering
type ItemData struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	AssetID         string            `json:"assetId"`
	SerialNumber    string            `json:"serialNumber"`
	ModelNumber     string            `json:"modelNumber"`
	Manufacturer    string            `json:"manufacturer"`
	LocationName    string            `json:"locationName"`
	LocationPath    []string          `json:"locationPath"`
	Labels          []string          `json:"labels"`
	CustomFields    map[string]string `json:"customFields"`
	PrimaryImageURL *string           `json:"primaryImageUrl,omitempty"`
	ItemURL         string            `json:"itemUrl"`
	Quantity        int               `json:"quantity"`
	Notes           string            `json:"notes"`
}

// LocationData contains all the data fields available for location label rendering
type LocationData struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        []string  `json:"path"`        // Hierarchy path (parent names)
	FullPath    string    `json:"fullPath"`    // Full path as string
	ItemCount   int       `json:"itemCount"`   // Number of items in this location
	LocationURL string    `json:"locationUrl"` // URL to view the location
}

// TemplateData contains the template configuration
type TemplateData struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Width        float64                `json:"width"`  // mm
	Height       float64                `json:"height"` // mm
	DPI          int                    `json:"dpi"`
	CanvasData   map[string]interface{} `json:"canvasData"`
	OutputFormat string                 `json:"outputFormat"`
}

// RenderContext contains all data needed to render a label
type RenderContext struct {
	Item     *ItemData     // Item data for item labels (mutually exclusive with Location)
	Location *LocationData // Location data for location labels (mutually exclusive with Item)
	Template *TemplateData
}

// CanvasElement represents an element from Fabric.js canvas JSON
type CanvasElement struct {
	Type        string                 `json:"type"`
	Left        float64                `json:"left"`
	Top         float64                `json:"top"`
	Width       float64                `json:"width"`
	Height      float64                `json:"height"`
	ScaleX      float64                `json:"scaleX"`
	ScaleY      float64                `json:"scaleY"`
	Angle       float64                `json:"angle"`
	Text        string                 `json:"text"`
	FontSize    float64                `json:"fontSize"`
	FontFamily  string                 `json:"fontFamily"`
	FontWeight  interface{}            `json:"fontWeight"` // Can be string or number
	Fill        string                 `json:"fill"`
	Stroke      string                 `json:"stroke"`
	StrokeWidth float64                `json:"strokeWidth"`
	TextAlign   string                 `json:"textAlign"`
	OriginX     string                 `json:"originX"` // "left", "center", "right"
	OriginY     string                 `json:"originY"` // "top", "center", "bottom"
	X1          float64                `json:"x1"`      // For lines
	Y1          float64                `json:"y1"`
	X2          float64                `json:"x2"`
	Y2          float64                `json:"y2"`
	Data        map[string]interface{} `json:"data"`    // Custom data (field type, barcode config, etc.)
	Objects     []interface{}          `json:"objects"` // For groups
}

// TemplateRenderer handles rendering of label templates
type TemplateRenderer struct {
	regularFont *truetype.Font
	boldFont    *truetype.Font
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() (*TemplateRenderer, error) {
	regularFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regular font: %w", err)
	}

	boldFont, err := truetype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bold font: %w", err)
	}

	return &TemplateRenderer{
		regularFont: regularFont,
		boldFont:    boldFont,
	}, nil
}

// mmToPixels converts millimeters to pixels at the given DPI
func mmToPixels(mm float64, dpi int) int {
	return int((mm / 25.4) * float64(dpi))
}

// screenDPI is the DPI used by the frontend canvas for display
const screenDPI = 96.0

// scaleCoord scales a coordinate from screen DPI to render DPI
func scaleCoord(coord float64, renderDPI int) int {
	return int(coord * float64(renderDPI) / screenDPI)
}

// scaleSize scales a size value from screen DPI to render DPI
func scaleSize(size float64, renderDPI int) float64 {
	return size * float64(renderDPI) / screenDPI
}

// RenderTemplate renders a label template with item data
func (r *TemplateRenderer) RenderTemplate(ctx RenderContext) ([]byte, error) {
	if ctx.Template == nil {
		return nil, fmt.Errorf("template is required")
	}

	dpi := ctx.Template.DPI
	if dpi <= 0 {
		dpi = 300
	}

	width := mmToPixels(ctx.Template.Width, dpi)
	height := mmToPixels(ctx.Template.Height, dpi)

	// Create blank white image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	// Parse and render canvas elements
	if ctx.Template.CanvasData != nil {
		err := r.renderCanvasElements(img, ctx.Template.CanvasData, ctx.Item, ctx.Location, dpi)
		if err != nil {
			return nil, fmt.Errorf("failed to render canvas elements: %w", err)
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// renderCanvasElements renders all elements from the canvas data
func (r *TemplateRenderer) renderCanvasElements(img *image.RGBA, canvasData map[string]interface{}, item *ItemData, location *LocationData, dpi int) error {
	objects, ok := canvasData["objects"].([]interface{})
	if !ok {
		return nil // No objects to render
	}

	for _, obj := range objects {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			continue
		}

		elem := parseCanvasElement(objMap)
		if err := r.renderElement(img, elem, item, location, dpi); err != nil {
			continue
		}
	}

	return nil
}

// parseCanvasElement parses a canvas element from a map
func parseCanvasElement(m map[string]interface{}) CanvasElement {
	elem := CanvasElement{
		ScaleX:      1,
		ScaleY:      1,
		StrokeWidth: 1,
		OriginX:     "left",
		OriginY:     "top",
	}

	if v, ok := m["type"].(string); ok {
		elem.Type = v
	}
	if v, ok := m["left"].(float64); ok {
		elem.Left = v
	}
	if v, ok := m["top"].(float64); ok {
		elem.Top = v
	}
	if v, ok := m["width"].(float64); ok {
		elem.Width = v
	}
	if v, ok := m["height"].(float64); ok {
		elem.Height = v
	}
	if v, ok := m["scaleX"].(float64); ok {
		elem.ScaleX = v
	}
	if v, ok := m["scaleY"].(float64); ok {
		elem.ScaleY = v
	}
	if v, ok := m["angle"].(float64); ok {
		elem.Angle = v
	}
	if v, ok := m["text"].(string); ok {
		elem.Text = v
	}
	if v, ok := m["fontSize"].(float64); ok {
		elem.FontSize = v
	}
	if v, ok := m["fontFamily"].(string); ok {
		elem.FontFamily = v
	}
	// fontWeight can be string ("bold", "normal") or number (400, 700)
	if v, ok := m["fontWeight"]; ok {
		elem.FontWeight = v
	}
	if v, ok := m["fill"].(string); ok {
		elem.Fill = v
	}
	if v, ok := m["stroke"].(string); ok {
		elem.Stroke = v
	}
	if v, ok := m["strokeWidth"].(float64); ok {
		elem.StrokeWidth = v
	}
	if v, ok := m["textAlign"].(string); ok {
		elem.TextAlign = v
	}
	if v, ok := m["originX"].(string); ok {
		elem.OriginX = v
	}
	if v, ok := m["originY"].(string); ok {
		elem.OriginY = v
	}
	if v, ok := m["x1"].(float64); ok {
		elem.X1 = v
	}
	if v, ok := m["y1"].(float64); ok {
		elem.Y1 = v
	}
	if v, ok := m["x2"].(float64); ok {
		elem.X2 = v
	}
	if v, ok := m["y2"].(float64); ok {
		elem.Y2 = v
	}
	if v, ok := m["data"].(map[string]interface{}); ok {
		elem.Data = v
	}
	if v, ok := m["objects"].([]interface{}); ok {
		elem.Objects = v
	}

	return elem
}

// renderElement renders a single element onto the image
func (r *TemplateRenderer) renderElement(img *image.RGBA, elem CanvasElement, item *ItemData, location *LocationData, dpi int) error {
	// First check custom data type - this takes priority for data_field and barcode
	if elem.Data != nil {
		if dataType, ok := elem.Data["type"].(string); ok {
			switch dataType {
			case "data_field":
				return r.renderDataField(img, elem, item, location, dpi)
			case "barcode":
				return r.renderBarcode(img, elem, item, location, dpi)
			}
		}
	}

	// Normalize type to handle both Fabric.js v5 (lowercase) and v6 (PascalCase)
	elemType := strings.ToLower(elem.Type)

	switch elemType {
	case "itext", "i-text", "text", "textbox":
		return r.renderText(img, elem, item, location, dpi)
	case "rect":
		return r.renderRect(img, elem, dpi)
	case "line":
		return r.renderLine(img, elem, dpi)
	case "group":
		return r.renderGroup(img, elem, item, location, dpi)
	case "image":
		return r.renderBarcode(img, elem, item, location, dpi)
	}

	return nil
}

// isBoldFont checks if the fontWeight indicates bold
func isBoldFont(fontWeight interface{}) bool {
	if fontWeight == nil {
		return false
	}
	switch v := fontWeight.(type) {
	case string:
		return v == "bold" || v == "700" || v == "800" || v == "900"
	case float64:
		return v >= 700
	case int:
		return v >= 700
	}
	return false
}

// autofitConfig holds autofit settings extracted from element data
type autofitConfig struct {
	enabled     bool
	fixedWidth  float64
	fixedHeight float64
	maxFontSize float64
}

// textRenderContext holds all computed values needed for text rendering
type textRenderContext struct {
	text                 string
	lines                []string
	fontSize             float64
	scaledFontSize       float64
	scaleY               float64
	font                 *truetype.Font
	face                 font.Face
	textColor            color.Color
	textAlign            string
	lineHeight           int
	textWidth            int
	textHeight           int
	containerWidth       float64
	scaledContainerWidth int
	metrics              font.Metrics
	autofit              autofitConfig
}

// getAutofitConfig extracts autofit settings from element data
func getAutofitConfig(elem CanvasElement) autofitConfig {
	cfg := autofitConfig{}
	if elem.Data == nil {
		return cfg
	}
	if v, ok := elem.Data["autofit"].(bool); ok {
		cfg.enabled = v
	}
	if v, ok := elem.Data["fixedWidth"].(float64); ok {
		cfg.fixedWidth = v
	}
	if v, ok := elem.Data["fixedHeight"].(float64); ok {
		cfg.fixedHeight = v
	}
	if v, ok := elem.Data["maxFontSize"].(float64); ok {
		cfg.maxFontSize = v
	}
	return cfg
}

// resolveTextContent replaces placeholders in text based on item or location data
func (r *TemplateRenderer) resolveTextContent(text string, item *ItemData, location *LocationData) string {
	if item != nil {
		return r.replaceDataPlaceholders(text, item)
	}
	if location != nil {
		return r.replaceLocationPlaceholders(text, location)
	}
	return text
}

// prepareTextRender computes all values needed for text rendering
func (r *TemplateRenderer) prepareTextRender(elem CanvasElement, text string, dpi int) textRenderContext {
	ctx := textRenderContext{text: text}

	// Font size with default
	ctx.fontSize = elem.FontSize
	if ctx.fontSize <= 0 {
		ctx.fontSize = 16
	}

	// Scale with default
	ctx.scaleY = elem.ScaleY
	if ctx.scaleY <= 0 {
		ctx.scaleY = 1
	}

	// Autofit configuration
	ctx.autofit = getAutofitConfig(elem)
	if ctx.autofit.fixedWidth <= 0 && elem.Width > 0 {
		ctx.autofit.fixedWidth = elem.Width
	}

	// Font selection
	ctx.font = r.regularFont
	if isBoldFont(elem.FontWeight) {
		ctx.font = r.boldFont
	}

	// Apply autofit if enabled
	if ctx.autofit.enabled && ctx.autofit.fixedWidth > 0 && ctx.autofit.fixedHeight > 0 {
		maxSize := ctx.autofit.maxFontSize
		if maxSize <= 0 {
			maxSize = ctx.fontSize
		}
		ctx.fontSize = r.findOptimalFontSize(text, ctx.font, ctx.autofit.fixedWidth, ctx.autofit.fixedHeight, maxSize, dpi)
	}

	// Scale font size for DPI
	ctx.scaledFontSize = scaleSize(ctx.fontSize*ctx.scaleY, dpi)

	// Create font face
	ctx.face = truetype.NewFace(ctx.font, &truetype.Options{
		Size: ctx.scaledFontSize,
		DPI:  72,
	})

	// Parse color
	fillColor := elem.Fill
	if fillColor == "" {
		fillColor = "#000000"
	}
	ctx.textColor = parseColor(fillColor)

	// Text metrics
	ctx.metrics = ctx.face.Metrics()
	ctx.lineHeight = (ctx.metrics.Ascent + ctx.metrics.Descent).Round()

	// Word wrapping
	if ctx.autofit.enabled && ctx.autofit.fixedWidth > 0 {
		scaledWidth := scaleCoord(ctx.autofit.fixedWidth, dpi)
		ctx.lines = wrapTextSimple(ctx.face, text, scaledWidth)
	} else {
		ctx.lines = []string{text}
	}

	// Calculate text dimensions
	ctx.textWidth = 0
	for _, line := range ctx.lines {
		w := measureTextWidth(ctx.face, line)
		if w > ctx.textWidth {
			ctx.textWidth = w
		}
	}
	ctx.textHeight = ctx.lineHeight * len(ctx.lines)

	// Text alignment
	ctx.textAlign = strings.ToLower(elem.TextAlign)
	if ctx.textAlign == "" {
		ctx.textAlign = "left"
	}

	// Container width for alignment
	ctx.containerWidth = ctx.autofit.fixedWidth
	if ctx.containerWidth <= 0 {
		ctx.containerWidth = elem.Width
	}
	ctx.scaledContainerWidth = scaleCoord(ctx.containerWidth, dpi)

	return ctx
}

// calculateAlignedX computes the x position for a line based on text alignment
func calculateAlignedX(align string, containerWidth, lineWidth, baseX int) int {
	switch align {
	case "center":
		if containerWidth > 0 {
			return baseX + (containerWidth-lineWidth)/2
		}
		return baseX
	case "right":
		if containerWidth > 0 {
			return baseX + containerWidth - lineWidth
		}
		return baseX
	default: // "left"
		return baseX
	}
}

// renderTextDirect renders non-rotated text directly onto the image
func (r *TemplateRenderer) renderTextDirect(img *image.RGBA, elem CanvasElement, ctx textRenderContext, dpi int) error {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ctx.font)
	c.SetFontSize(ctx.scaledFontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(ctx.textColor))

	baseX := scaleCoord(elem.Left, dpi)
	baseY := scaleCoord(elem.Top, dpi) + ctx.metrics.Ascent.Round()

	fmt.Printf("DEBUG renderTextDirect: drawing %d lines at baseX=%d, baseY=%d, textAlign=%s\n",
		len(ctx.lines), baseX, baseY, ctx.textAlign)

	for i, line := range ctx.lines {
		lineWidth := measureTextWidth(ctx.face, line)
		x := calculateAlignedX(ctx.textAlign, ctx.scaledContainerWidth, lineWidth, baseX)
		y := baseY + (ctx.lineHeight * i)

		pt := freetype.Pt(x, y)
		if _, err := c.DrawString(line, pt); err != nil {
			return err
		}
	}
	return nil
}

// renderTextRotated renders rotated text by drawing to a temp image, rotating, and compositing
func (r *TemplateRenderer) renderTextRotated(img *image.RGBA, elem CanvasElement, ctx textRenderContext, dpi int) error {
	// Create temporary image - use container width if alignment matters
	margin := 2
	tempWidth := ctx.textWidth + margin*2
	if ctx.scaledContainerWidth > 0 && (ctx.textAlign == "center" || ctx.textAlign == "right") {
		tempWidth = ctx.scaledContainerWidth + margin*2
	}
	tempHeight := ctx.textHeight + margin*2

	// Create temporary image with transparent background
	tempImg := image.NewRGBA(image.Rect(0, 0, tempWidth, tempHeight))

	// Draw text on temporary image
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ctx.font)
	c.SetFontSize(ctx.scaledFontSize)
	c.SetClip(tempImg.Bounds())
	c.SetDst(tempImg)
	c.SetSrc(image.NewUniform(ctx.textColor))

	baseY := margin + ctx.metrics.Ascent.Round()

	for i, line := range ctx.lines {
		lineWidth := measureTextWidth(ctx.face, line)

		// Calculate x position based on alignment (relative to temp image)
		var textX int
		switch ctx.textAlign {
		case "center":
			textX = margin + (tempWidth-margin*2-lineWidth)/2
		case "right":
			textX = tempWidth - margin - lineWidth
		default: // "left"
			textX = margin
		}

		y := baseY + (ctx.lineHeight * i)
		pt := freetype.Pt(textX, y)
		if _, err := c.DrawString(line, pt); err != nil {
			return err
		}
	}

	// Rotate the temporary image
	// Fabric.js uses clockwise rotation, bild's transform.Rotate uses counter-clockwise
	rotatedImg := transform.Rotate(tempImg, -elem.Angle, nil)

	// Get rotated image dimensions and composite onto main image
	rotatedBounds := rotatedImg.Bounds()
	x := scaleCoord(elem.Left, dpi)
	y := scaleCoord(elem.Top, dpi)
	destRect := image.Rect(x, y, x+rotatedBounds.Dx(), y+rotatedBounds.Dy())
	draw.Draw(img, destRect, rotatedImg, rotatedBounds.Min, draw.Over)

	fmt.Printf("DEBUG renderTextRotated: destRect=%v, rotatedSize=%dx%d\n",
		destRect, rotatedBounds.Dx(), rotatedBounds.Dy())

	return nil
}

// renderText renders a text element
func (r *TemplateRenderer) renderText(img *image.RGBA, elem CanvasElement, item *ItemData, location *LocationData, dpi int) error {
	fmt.Printf("DEBUG renderText ENTRY: elem.Text=%q, elem.Angle=%.1f, elem.Left=%.1f, elem.Top=%.1f, dpi=%d\n",
		elem.Text, elem.Angle, elem.Left, elem.Top, dpi)

	// Resolve text content with placeholder replacement
	text := r.resolveTextContent(elem.Text, item, location)
	if text == "" {
		fmt.Printf("DEBUG renderText: text is empty after placeholder replacement, returning\n")
		return nil
	}

	fmt.Printf("DEBUG renderText RENDERING: text=%q, angle=%.1f, fontSize=%.1f, scaleX=%.2f, scaleY=%.2f\n",
		text, elem.Angle, elem.FontSize, elem.ScaleX, elem.ScaleY)

	// Prepare all rendering context
	ctx := r.prepareTextRender(elem, text, dpi)

	// Dispatch to appropriate renderer based on rotation
	if elem.Angle == 0 {
		return r.renderTextDirect(img, elem, ctx, dpi)
	}
	return r.renderTextRotated(img, elem, ctx, dpi)
}

// measureTextWidth calculates the width of text in pixels
func measureTextWidth(face font.Face, text string) int {
	var width int
	for _, ch := range text {
		awidth, ok := face.GlyphAdvance(ch)
		if ok {
			width += awidth.Round()
		}
	}
	return width
}

// wrapTextSimple wraps text to fit within a given width, returning lines
func wrapTextSimple(face font.Face, text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			// First word on the line
			currentLine = word
		} else {
			// Try adding the word to current line
			testLine := currentLine + " " + word
			testWidth := measureTextWidth(face, testLine)

			if testWidth <= maxWidth {
				currentLine = testLine
			} else {
				// Word doesn't fit, start a new line
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
	}

	// Don't forget the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// calculateTextHeight calculates the total height of wrapped text
func calculateTextHeight(face font.Face, lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	metrics := face.Metrics()
	lineHeight := (metrics.Ascent + metrics.Descent).Round()
	return lineHeight * len(lines)
}

// findOptimalFontSize finds the largest font size that fits text within bounds
func (r *TemplateRenderer) findOptimalFontSize(text string, fontFace *truetype.Font, maxWidth, maxHeight, maxFontSize float64, dpi int) float64 {
	const minFontSize = 6.0

	// Scale bounds from screen DPI to pixels
	// The bounds are in screen pixels (96 DPI), we need to find font size in screen DPI
	// that will fit, then it gets scaled to render DPI later
	widthPx := int(maxWidth)
	heightPx := int(maxHeight)

	// Binary search for optimal font size
	low := minFontSize
	high := maxFontSize
	bestSize := minFontSize

	for high-low > 0.5 {
		mid := (low + high) / 2

		// Create face at this size (in screen coordinates)
		face := truetype.NewFace(fontFace, &truetype.Options{
			Size: mid,
			DPI:  72,
		})

		// Wrap text and calculate dimensions
		lines := wrapTextSimple(face, text, widthPx)
		textHeight := calculateTextHeight(face, lines)

		// Check if all lines fit width-wise
		allLinesFit := true
		for _, line := range lines {
			lineWidth := measureTextWidth(face, line)
			if lineWidth > widthPx {
				allLinesFit = false
				break
			}
		}

		if allLinesFit && textHeight <= heightPx {
			// This size fits, try larger
			bestSize = mid
			low = mid + 0.5
		} else {
			// Too big, try smaller
			high = mid - 0.5
		}
	}

	fmt.Printf("DEBUG findOptimalFontSize: text=%q maxWidth=%.1f maxHeight=%.1f maxFontSize=%.1f -> bestSize=%.1f\n",
		text, maxWidth, maxHeight, maxFontSize, bestSize)

	return bestSize
}

// renderDataField renders a data field element with word wrap and auto-fit support
func (r *TemplateRenderer) renderDataField(img *image.RGBA, elem CanvasElement, item *ItemData, location *LocationData, dpi int) error {
	if elem.Data == nil {
		fmt.Printf("DEBUG renderDataField: elem.Data is nil - returning early\n")
		return nil
	}

	field, _ := elem.Data["field"].(string)
	var value string

	// Get field value from item or location
	switch {
	case item != nil:
		value = r.getFieldValue(field, item)
		fmt.Printf("DEBUG renderDataField: field=%q, value=%q, item.Name=%q, item.Description=%q\n",
			field, value, item.Name, item.Description)
	case location != nil:
		value = r.getLocationFieldValue(field, location)
		fmt.Printf("DEBUG renderDataField: field=%q, value=%q, location.Name=%q\n",
			field, value, location.Name)
	default:
		fmt.Printf("DEBUG renderDataField: no item or location data - returning early\n")
		return nil
	}

	if value == "" {
		fmt.Printf("DEBUG renderDataField: value is empty, returning without rendering\n")
		return nil
	}

	// Create a modified element with the field value as text and render as normal text
	textElem := elem
	textElem.Text = value

	return r.renderText(img, textElem, item, location, dpi)
}

// renderBarcode renders a barcode element
func (r *TemplateRenderer) renderBarcode(img *image.RGBA, elem CanvasElement, item *ItemData, location *LocationData, dpi int) error {
	if elem.Data == nil {
		return nil
	}

	format, _ := elem.Data["format"].(string)
	contentSource, _ := elem.Data["contentSource"].(string)

	// Get barcode content based on source - try item first, then location
	var content string
	if item != nil {
		content = r.getBarcodeContent(contentSource, item)
	} else if location != nil {
		content = r.getLocationBarcodeContent(contentSource, location)
	}

	if content == "" {
		return nil
	}

	// Calculate dimensions
	width := scaleCoord(elem.Width*elem.ScaleX, dpi)
	height := scaleCoord(elem.Height*elem.ScaleY, dpi)

	if width <= 0 || height <= 0 {
		width = 100
		height = 100
	}

	// Generate barcode
	opts := BarcodeOptions{
		Width:  width,
		Height: height,
	}

	barcodeImg, err := GenerateBarcode(BarcodeFormat(format), content, opts)
	if err != nil {
		// Draw placeholder on error
		barcodeImg = CreatePlaceholderBarcode(width, height, BarcodeFormat(format))
	}

	// Calculate position
	x := scaleCoord(elem.Left, dpi)
	y := scaleCoord(elem.Top, dpi)

	// Draw barcode onto image
	destRect := image.Rect(x, y, x+width, y+height)
	draw.Draw(img, destRect, barcodeImg, image.Point{}, draw.Over)

	return nil
}

// renderRect renders a rectangle element
func (r *TemplateRenderer) renderRect(img *image.RGBA, elem CanvasElement, dpi int) error {
	x := scaleCoord(elem.Left, dpi)
	y := scaleCoord(elem.Top, dpi)
	width := scaleCoord(elem.Width*elem.ScaleX, dpi)
	height := scaleCoord(elem.Height*elem.ScaleY, dpi)

	// Draw fill if not transparent
	if elem.Fill != "" && elem.Fill != "transparent" {
		fillColor := parseColor(elem.Fill)
		for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
			for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
				if px >= 0 && py >= 0 {
					img.Set(px, py, fillColor)
				}
			}
		}
	}

	// Draw stroke if present
	if elem.Stroke != "" && elem.Stroke != "transparent" {
		strokeColor := parseColor(elem.Stroke)
		strokeWidth := scaleCoord(elem.StrokeWidth, dpi)
		if strokeWidth < 1 {
			strokeWidth = 1
		}

		// Draw horizontal lines (top and bottom)
		for sw := 0; sw < strokeWidth; sw++ {
			for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
				if px >= 0 {
					if y+sw >= 0 && y+sw < img.Bounds().Max.Y {
						img.Set(px, y+sw, strokeColor)
					}
					if y+height-1-sw >= 0 && y+height-1-sw < img.Bounds().Max.Y {
						img.Set(px, y+height-1-sw, strokeColor)
					}
				}
			}
		}

		// Draw vertical lines (left and right)
		for sw := 0; sw < strokeWidth; sw++ {
			for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
				if py >= 0 {
					if x+sw >= 0 && x+sw < img.Bounds().Max.X {
						img.Set(x+sw, py, strokeColor)
					}
					if x+width-1-sw >= 0 && x+width-1-sw < img.Bounds().Max.X {
						img.Set(x+width-1-sw, py, strokeColor)
					}
				}
			}
		}
	}

	return nil
}

// renderLine renders a line element
func (r *TemplateRenderer) renderLine(img *image.RGBA, elem CanvasElement, dpi int) error {
	// Get line endpoints relative to element position
	x1 := scaleCoord(elem.Left+elem.X1, dpi)
	y1 := scaleCoord(elem.Top+elem.Y1, dpi)
	x2 := scaleCoord(elem.Left+elem.X2, dpi)
	y2 := scaleCoord(elem.Top+elem.Y2, dpi)

	strokeColor := parseColor(elem.Stroke)
	if elem.Stroke == "" {
		strokeColor = color.Black
	}

	strokeWidth := scaleCoord(elem.StrokeWidth, dpi)
	if strokeWidth < 1 {
		strokeWidth = 1
	}

	// Bresenham's line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		// Draw a thick point
		for sw := -strokeWidth / 2; sw <= strokeWidth/2; sw++ {
			for sh := -strokeWidth / 2; sh <= strokeWidth/2; sh++ {
				px, py := x1+sw, y1+sh
				if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
					img.Set(px, py, strokeColor)
				}
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}

	return nil
}

// renderGroup renders a group element (which may contain a barcode)
func (r *TemplateRenderer) renderGroup(img *image.RGBA, elem CanvasElement, item *ItemData, location *LocationData, dpi int) error {
	// Check if this group is a barcode
	if elem.Data != nil {
		if dataType, ok := elem.Data["type"].(string); ok && dataType == "barcode" {
			return r.renderBarcode(img, elem, item, location, dpi)
		}
	}

	// For other groups, we'd need to render child objects
	// This is a basic implementation that doesn't handle nested objects
	return nil
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// replaceDataPlaceholders replaces {{field}} placeholders with actual values
func (r *TemplateRenderer) replaceDataPlaceholders(text string, item *ItemData) string {
	if item == nil {
		return text
	}

	replacements := map[string]string{
		"{{item_name}}":     item.Name,
		"{{name}}":          item.Name,
		"{{description}}":   item.Description,
		"{{asset_id}}":      item.AssetID,
		"{{serial_number}}": item.SerialNumber,
		"{{model_number}}":  item.ModelNumber,
		"{{manufacturer}}":  item.Manufacturer,
		"{{location}}":      item.LocationName,
		"{{location_name}}": item.LocationName,
		"{{location_path}}": strings.Join(item.LocationPath, " > "),
		"{{labels}}":        strings.Join(item.Labels, ", "),
		"{{item_url}}":      item.ItemURL,
		"{{quantity}}":      fmt.Sprintf("%d", item.Quantity),
		"{{notes}}":         item.Notes,
		"{{current_date}}":  time.Now().Format("2006-01-02"),
		"{{current_time}}":  time.Now().Format("15:04"),
	}

	result := text
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Handle custom fields
	for key, value := range item.CustomFields {
		placeholder := fmt.Sprintf("{{custom_%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// getFieldValue returns the value for a specific field name
func (r *TemplateRenderer) getFieldValue(field string, item *ItemData) string {
	if item == nil {
		return ""
	}

	switch field {
	case "item_name", "name":
		return item.Name
	case "description":
		return item.Description
	case "asset_id":
		return item.AssetID
	case "serial_number":
		return item.SerialNumber
	case "model_number":
		return item.ModelNumber
	case "manufacturer":
		return item.Manufacturer
	case "location", "location_name":
		return item.LocationName
	case "location_path":
		return strings.Join(item.LocationPath, " > ")
	case "labels":
		return strings.Join(item.Labels, ", ")
	case "item_url":
		return item.ItemURL
	case "quantity":
		return fmt.Sprintf("%d", item.Quantity)
	case "notes":
		return item.Notes
	case "current_date":
		return time.Now().Format("2006-01-02")
	case "current_time":
		return time.Now().Format("15:04")
	default:
		// Check custom fields
		if strings.HasPrefix(field, "custom_") {
			key := strings.TrimPrefix(field, "custom_")
			if value, ok := item.CustomFields[key]; ok {
				return value
			}
		}
		return ""
	}
}

// replaceLocationPlaceholders replaces {{field}} placeholders with location data
func (r *TemplateRenderer) replaceLocationPlaceholders(text string, location *LocationData) string {
	if location == nil {
		return text
	}

	replacements := map[string]string{
		"{{location_name}}": location.Name,
		"{{name}}":          location.Name,
		"{{description}}":   location.Description,
		"{{location_path}}": location.FullPath,
		"{{path}}":          location.FullPath,
		"{{item_count}}":    fmt.Sprintf("%d", location.ItemCount),
		"{{location_url}}":  location.LocationURL,
		"{{current_date}}":  time.Now().Format("2006-01-02"),
		"{{current_time}}":  time.Now().Format("15:04"),
	}

	result := text
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// getLocationFieldValue returns the value for a specific field name from location data
func (r *TemplateRenderer) getLocationFieldValue(field string, location *LocationData) string {
	if location == nil {
		return ""
	}

	switch field {
	case "location_name", "name":
		return location.Name
	case "description":
		return location.Description
	case "location_path", "path":
		return location.FullPath
	case "item_count":
		return fmt.Sprintf("%d", location.ItemCount)
	case "location_url":
		return location.LocationURL
	case "current_date":
		return time.Now().Format("2006-01-02")
	case "current_time":
		return time.Now().Format("15:04")
	default:
		return ""
	}
}

// getLocationBarcodeContent returns the content for a barcode based on source type for locations
func (r *TemplateRenderer) getLocationBarcodeContent(source string, location *LocationData) string {
	if location == nil {
		return ""
	}

	switch source {
	case "location_url":
		return location.LocationURL
	case "location_name", "name":
		return location.Name
	case "location_path", "path":
		return location.FullPath
	case "id":
		return location.ID.String()
	default:
		return location.LocationURL // Default to URL for QR codes
	}
}

// getBarcodeContent returns the content for a barcode based on source type
func (r *TemplateRenderer) getBarcodeContent(source string, item *ItemData) string {
	if item == nil {
		return ""
	}

	switch source {
	case "url", "item_url":
		return item.ItemURL
	case "asset_id":
		return item.AssetID
	case "serial_number":
		return item.SerialNumber
	case "model_number":
		return item.ModelNumber
	case "id":
		return item.ID.String()
	default:
		// If source looks like a custom value, return it directly
		if source != "" && !strings.HasPrefix(source, "{{") {
			return source
		}
		return item.ItemURL
	}
}

// parseColor parses a color string (hex or named) into a color.Color
func parseColor(s string) color.Color {
	if s == "" {
		return color.Black
	}

	// Handle hex colors
	if strings.HasPrefix(s, "#") {
		s = s[1:]
		var r, g, b uint8

		switch len(s) {
		case 3:
			_, _ = fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
			r *= 17
			g *= 17
			b *= 17
		case 6:
			_, _ = fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		default:
			return color.Black
		}

		return color.RGBA{R: r, G: g, B: b, A: 255}
	}

	// Handle named colors
	switch strings.ToLower(s) {
	case "black":
		return color.Black
	case "white":
		return color.White
	case "red":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}
	case "green":
		return color.RGBA{R: 0, G: 128, B: 0, A: 255}
	case "blue":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}
	case "gray", "grey":
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	default:
		return color.Black
	}
}

// RenderPreview renders a template preview with sample data
func (r *TemplateRenderer) RenderPreview(template *TemplateData) ([]byte, error) {
	sampleItem := &ItemData{
		ID:           uuid.New(),
		Name:         "Sample Item",
		Description:  "This is a sample item for preview",
		AssetID:      "A-001234",
		SerialNumber: "SN123456789",
		ModelNumber:  "MODEL-X100",
		Manufacturer: "Acme Corp",
		LocationName: "Storage Room",
		LocationPath: []string{"Building A", "Floor 2", "Storage Room"},
		Labels:       []string{"Electronics", "Office"},
		CustomFields: map[string]string{
			"color": "Blue",
			"size":  "Medium",
		},
		ItemURL:  "https://example.com/items/sample",
		Quantity: 1,
		Notes:    "Sample notes for preview",
	}

	return r.RenderTemplate(RenderContext{
		Item:     sampleItem,
		Template: template,
	})
}

// RenderLocationPreview renders a template preview with sample location data
func (r *TemplateRenderer) RenderLocationPreview(template *TemplateData) ([]byte, error) {
	sampleLocation := &LocationData{
		ID:          uuid.New(),
		Name:        "Storage Room",
		Description: "Main storage area for electronics",
		Path:        []string{"Building A", "Floor 2"},
		FullPath:    "Building A > Floor 2 > Storage Room",
		ItemCount:   42,
		LocationURL: "https://example.com/locations/sample",
	}

	return r.RenderTemplate(RenderContext{
		Location: sampleLocation,
		Template: template,
	})
}
