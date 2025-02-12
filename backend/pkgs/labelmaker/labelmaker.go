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

func GenerateLabel(w io.Writer, params *GenerateParameters) error {
	if err := params.Validate(); err != nil {
		return err
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

	err = GenerateLabel(f, params)
	if err != nil {
		return err
	}

	if cfg.LabelMaker.PrintCommand == nil {
		return fmt.Errorf("no print command specified")
	}

	commandTemplate := template.Must(template.New("command").Parse(*cfg.LabelMaker.PrintCommand))
	builder := &strings.Builder{}
	if err := commandTemplate.Execute(builder, map[string]string{
		"FileName": f.Name(),
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
