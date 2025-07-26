// Package labelmaker provides functionality for generating and printing labels for items, locations and assets stored in Homebox
package labelmaker

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
)

type GenerateParameters struct {
	Width                 int
	Height                int
	QrSize                int
	Margin                int
	ComponentPadding      int
	TitleText             string
	TitleFontSize         float64
	DescriptionText       string
	DescriptionFontSize   float64
	AdditionalInformation *string
	Dpi                   float64
	URL                   string
	DynamicLength         bool
}

func (p *GenerateParameters) Validate() error {
	if p.Width <= 0 {
		return fmt.Errorf("invalid width")
	}
	if p.Height <= 0 {
		return fmt.Errorf("invalid height")
	}
	if p.Margin < 0 {
		return fmt.Errorf("invalid margin")
	}
	if p.ComponentPadding < 0 {
		return fmt.Errorf("invalid component padding")
	}
	return nil
}

func NewGenerateParams(width int, height int, margin int, padding int, fontSize float64, title string, description string, url string, dynamicLength bool, additionalInformation *string) GenerateParameters {
	return GenerateParameters{
		Width:                 width,
		Height:                height,
		QrSize:                height - (padding * 2),
		Margin:                margin,
		ComponentPadding:      padding,
		TitleText:             title,
		DescriptionText:       description,
		TitleFontSize:         fontSize,
		DescriptionFontSize:   fontSize * 0.8,
		Dpi:                   72,
		URL:                   url,
		AdditionalInformation: additionalInformation,
		DynamicLength:         dynamicLength,
	}
}

func measureString(text string, face font.Face, ctx *freetype.Context) int {
	width := 0
	for _, r := range text {
		awidth, _ := face.GlyphAdvance(r)
		width += awidth.Round()
	}
	return ctx.PointToFixed(float64(width)).Round()
}

func wrapText(text string, face font.Face, maxWidth int, maxHeight int, lineHeight int, ctx *freetype.Context) ([]string, string) {
	lines := strings.Split(text, "\n")
	unlimitedHeight := maxHeight == -1
	var wrappedLines []string
	currentHeight := 0
	processedChars := 0

	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, "")
			processedChars += 1
			if !unlimitedHeight {
				currentHeight += lineHeight
				if currentHeight > maxHeight {
					return wrappedLines[:len(wrappedLines)-1], text[processedChars:]
				}
			}
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			testLine := currentLine + " " + word
			width := measureString(testLine, face, ctx)

			if width <= maxWidth {
				currentLine = testLine
			} else {
				wrappedLines = append(wrappedLines, currentLine)
				processedChars += len(currentLine) + 1
				if !unlimitedHeight {
					currentHeight += lineHeight
					if currentHeight > maxHeight {
						return wrappedLines[:len(wrappedLines)-1], text[processedChars-len(currentLine)-1:]
					}
				}
				currentLine = word
			}
		}

		wrappedLines = append(wrappedLines, currentLine)
		processedChars += len(currentLine) + 1
		if !unlimitedHeight {
			currentHeight += lineHeight
			if currentHeight > maxHeight {
				return wrappedLines[:len(wrappedLines)-1], text[processedChars-len(currentLine)-1:]
			}
		}
	}

	return wrappedLines, ""
}

func GenerateLabel(w io.Writer, params *GenerateParameters, cfg *config.Config) error {
	if err := params.Validate(); err != nil {
		return err
	}

	// If LabelServiceUrl is configured, fetch the label from the URL instead of generating it
	if cfg != nil && cfg.LabelMaker.LabelServiceUrl != nil && *cfg.LabelMaker.LabelServiceUrl != "" {
		log.Printf("LabelServiceUrl configured: %s", *cfg.LabelMaker.LabelServiceUrl)

		return fetchLabelFromURL(w, *cfg.LabelMaker.LabelServiceUrl, params, cfg)
	}

	bodyText := params.DescriptionText
	if params.AdditionalInformation != nil {
		bodyText = bodyText + "\n" + *params.AdditionalInformation
	}

	// Create QR code
	qr, err := qrcode.New(params.URL, qrcode.Medium)
	if err != nil {
		return err
	}
	qr.DisableBorder = true
	qrImage := qr.Image(params.QrSize)

	regularFont, err := truetype.Parse(gomedium.TTF)
	if err != nil {
		return err
	}

	boldFont, err := truetype.Parse(gobold.TTF)
	if err != nil {
		return err
	}

	regularFace := truetype.NewFace(regularFont, &truetype.Options{
		Size: params.DescriptionFontSize,
		DPI:  params.Dpi,
	})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{
		Size: params.TitleFontSize,
		DPI:  params.Dpi,
	})

	// Calculate text area dimensions
	maxWidth := params.Width - (params.Margin * 2) - params.ComponentPadding

	// Create temporary contexts for text measurement
	tmpImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	boldContext := createContext(boldFont, params.TitleFontSize, tmpImg, params.Dpi)
	regularContext := createContext(regularFont, params.DescriptionFontSize, tmpImg, params.Dpi)

	// Calculate total height needed
	totalHeight := params.Margin
	titleLineSpacing := boldContext.PointToFixed(params.TitleFontSize).Round()

	titleLines, _ := wrapText(params.TitleText, boldFace, maxWidth-params.QrSize, -1, titleLineSpacing, boldContext)
	titleHeight := titleLineSpacing * len(titleLines)
	totalHeight += titleHeight

	totalHeight += params.ComponentPadding / 4

	regularLineSpacing := regularContext.PointToFixed(params.DescriptionFontSize).Round()
	descriptionLinesRight, descriptionRemaining := wrapText(bodyText, regularFace, maxWidth-params.QrSize, params.QrSize-titleHeight, regularLineSpacing, regularContext)
	totalHeight += regularLineSpacing * len(descriptionLinesRight)

	var textYBottomText int
	var descriptionLinesBottom []string
	hasBottomText := descriptionRemaining != ""
	if hasBottomText {
		totalHeight = max(params.Margin+params.QrSize+params.ComponentPadding/2, totalHeight)
		textYBottomText = totalHeight
		descriptionLinesBottom, _ = wrapText(descriptionRemaining, regularFace, maxWidth, -1, regularLineSpacing, regularContext)
		totalHeight += regularLineSpacing * len(descriptionLinesBottom)
		totalHeight += params.Margin
	}

	var requiredHeight int
	if params.DynamicLength {
		requiredHeight = max(totalHeight, params.QrSize+(params.Margin*2))
	} else {
		requiredHeight = params.Height
	}

	// Create the actual image with calculated height
	bounds := image.Rect(0, 0, params.Width, requiredHeight)
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw QR code onto the image
	draw.Draw(img,
		image.Rect(params.Margin, params.Margin, params.QrSize+params.Margin, params.QrSize+params.Margin),
		qrImage,
		image.Point{},
		draw.Over)

	// Create final drawing contexts
	boldContext = createContext(boldFont, params.TitleFontSize, img, params.Dpi)
	regularContext = createContext(regularFont, params.DescriptionFontSize, img, params.Dpi)

	textXRight := params.Margin + params.ComponentPadding + params.QrSize
	textY := params.Margin - 8

	// Draw title
	for _, line := range titleLines {
		pt := freetype.Pt(textXRight, textY+titleLineSpacing)
		if _, err = boldContext.DrawString(line, pt); err != nil {
			return err
		}
		textY += titleLineSpacing
	}

	// Draw description right from QR Code
	textY += params.ComponentPadding / 4
	for _, line := range descriptionLinesRight {
		pt := freetype.Pt(textXRight, textY+regularLineSpacing)
		if _, err = regularContext.DrawString(line, pt); err != nil {
			return err
		}
		textY += regularLineSpacing
	}

	// Draw description below QR Code
	if hasBottomText {
		for _, line := range descriptionLinesBottom {
			pt := freetype.Pt(params.Margin, textYBottomText+regularLineSpacing)
			if _, err = regularContext.DrawString(line, pt); err != nil {
				return err
			}
			textYBottomText += regularLineSpacing
		}
	}

	return png.Encode(w, img)
}

// Helper function to create freetype context
func createContext(font *truetype.Font, size float64, img *image.RGBA, dpi float64) *freetype.Context {
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(font)
	c.SetFontSize(size)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.Black))
	return c
}

// fetchLabelFromURL fetches an image from the specified URL and writes it to the writer
func fetchLabelFromURL(w io.Writer, serviceURL string, params *GenerateParameters, cfg *config.Config) error {
	// Parse the base URL
	baseURL, err := url.Parse(serviceURL)
	if err != nil {
		return fmt.Errorf("failed to parse service URL %s: %w", serviceURL, err)
	}

	// Build query parameters with the same attributes passed to print command
	query := url.Values{}
	query.Set("Width", fmt.Sprintf("%d", params.Width))
	query.Set("Height", fmt.Sprintf("%d", params.Height))
	query.Set("QrSize", fmt.Sprintf("%d", params.QrSize))
	query.Set("Margin", fmt.Sprintf("%d", params.Margin))
	query.Set("ComponentPadding", fmt.Sprintf("%d", params.ComponentPadding))
	query.Set("TitleText", params.TitleText)
	query.Set("TitleFontSize", fmt.Sprintf("%f", params.TitleFontSize))
	query.Set("DescriptionText", params.DescriptionText)
	query.Set("DescriptionFontSize", fmt.Sprintf("%f", params.DescriptionFontSize))
	query.Set("Dpi", fmt.Sprintf("%f", params.Dpi))
	query.Set("URL", params.URL)
	query.Set("DynamicLength", fmt.Sprintf("%t", params.DynamicLength))

	// Add AdditionalInformation if it exists
	if params.AdditionalInformation != nil {
		query.Set("AdditionalInformation", *params.AdditionalInformation)
	}

	// Set the query parameters
	baseURL.RawQuery = query.Encode()
	finalServiceURL := baseURL.String()

	log.Printf("Fetching label from URL: %s", finalServiceURL)

	// Use configured timeout or default to 30 seconds
	timeout := 30 * time.Second
	if cfg != nil && cfg.LabelMaker.LabelServiceTimeout != nil {
		timeout = *cfg.LabelMaker.LabelServiceTimeout
	}

	// Create HTTP client with configurable timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create HTTP request with custom headers
	req, err := http.NewRequest("GET", finalServiceURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for URL %s: %w", finalServiceURL, err)
	}

	// Set custom headers
	req.Header.Set("User-Agent", "Homebox-LabelMaker/1.0")
	req.Header.Set("Accept", "image/*")

	// Make HTTP request to the label service
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch label from URL %s: %w", finalServiceURL, err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("label service returned status %d for URL %s", resp.StatusCode, finalServiceURL)
	}

	// Check if the response is an image
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("label service returned invalid content type %s, expected image/*", contentType)
	}

	// Set default max response size (10MB)
	maxResponseSize := int64(10 << 20)
	if cfg != nil {
		maxResponseSize = cfg.Web.MaxUploadSize << 20
	}
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	// Copy the response body to the writer
	_, err = io.Copy(w, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to write fetched label data: %w", err)
	}

	return nil
}

func PrintLabel(cfg *config.Config, params *GenerateParameters) error {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("label-%d.png", time.Now().UnixNano()))
	f, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("failed to remove temporary label file: %v", err)
		}
	}()

	err = GenerateLabel(f, params, cfg)
	if err != nil {
		return err
	}

	if cfg.LabelMaker.PrintCommand == nil {
		return fmt.Errorf("no print command specified")
	}

	commandTemplate := template.Must(template.New("command").Parse(*cfg.LabelMaker.PrintCommand))
	builder := &strings.Builder{}
	additionalInformation := func() string {
		if params.AdditionalInformation != nil {
			return *params.AdditionalInformation
		}
		return ""
	}()
	if err := commandTemplate.Execute(builder, map[string]string{
		"FileName":              f.Name(),
		"Width":                 fmt.Sprintf("%d", params.Width),
		"Height":                fmt.Sprintf("%d", params.Height),
		"QrSize":                fmt.Sprintf("%d", params.QrSize),
		"Margin":                fmt.Sprintf("%d", params.Margin),
		"ComponentPadding":      fmt.Sprintf("%d", params.ComponentPadding),
		"TitleText":             params.TitleText,
		"TitleFontSize":         fmt.Sprintf("%f", params.TitleFontSize),
		"DescriptionText":       params.DescriptionText,
		"DescriptionFontSize":   fmt.Sprintf("%f", params.DescriptionFontSize),
		"AdditionalInformation": additionalInformation,
		"Dpi":                   fmt.Sprintf("%f", params.Dpi),
		"URL":                   params.URL,
		"DynamicLength":         fmt.Sprintf("%t", params.DynamicLength),
	}); err != nil {
		return err
	}

	commandParts := strings.Fields(builder.String())
	if len(commandParts) == 0 {
		return nil
	}

	command := exec.Command(commandParts[0], commandParts[1:]...)

	_, err = command.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
