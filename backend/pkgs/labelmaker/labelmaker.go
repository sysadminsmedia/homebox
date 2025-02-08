// labelmaker package provides functionality for generating and printing labels for items, locations and assets stored in Homebox
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
	"golang.org/x/image/font/gofont/goregular"
)

type GenerateParameters struct {
	Width               int
	Height              int
	QrSize              int
	Margin              int
	ComponentPadding    int
	TitleText           string
	TitleFontSize       float64
	DescriptionText     string
	DescriptionFontSize float64
	Dpi                 float64
	URL                 string
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

func NewGenerateParams(width int, height int, margin int, padding int, fontSize float64, title string, description string, url string) GenerateParameters {
	return GenerateParameters{
		Width:               width,
		Height:              height,
		QrSize:              height - (padding * 2),
		Margin:              margin,
		ComponentPadding:    padding,
		TitleText:           title,
		DescriptionText:     description,
		TitleFontSize:       fontSize,
		DescriptionFontSize: fontSize * 0.8,
		Dpi:                 72,
		URL:                 url,
	}
}

func measureString(text string, face font.Face, ctx *freetype.Context) int {
	width := 0
	for _, r := range text {
		awidth, _ := face.GlyphAdvance(r)
		width += awidth.Round()
	}
	return int(ctx.PointToFixed(float64(width)).Round())
}

// wrapText breaks text into lines that fit within maxWidth
func wrapText(text string, face font.Face, maxWidth int, ctx *freetype.Context) []string {
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, "")
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
				currentLine = word
			}
		}
		wrappedLines = append(wrappedLines, currentLine)
	}

	// Handle lines that are too long and have no spaces
	for i, line := range wrappedLines {
		width := measureString(line, face, ctx)
		if width > maxWidth {
			var splitLines []string
			currentLine := ""
			for _, r := range line {
				testLine := currentLine + string(r)
				width := measureString(testLine, face, ctx)
				if width <= maxWidth {
					currentLine = testLine
				} else {
					splitLines = append(splitLines, currentLine)
					currentLine = string(r)
				}
			}
			splitLines = append(splitLines, currentLine)
			wrappedLines = append(wrappedLines[:i], append(splitLines, wrappedLines[i+1:]...)...)
		}
	}

	return wrappedLines
}

func GenerateLabel(w io.Writer, params *GenerateParameters) error {
	if err := params.Validate(); err != nil {
		return err
	}

	// Create QR code
	qr, err := qrcode.New(params.URL, qrcode.Medium)
	if err != nil {
		return err
	}
	qr.DisableBorder = true
	qrImage := qr.Image(params.QrSize)

	// Create a new white background image
	bounds := image.Rect(0, 0, params.Width, params.Height)
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw QR code onto the image
	draw.Draw(img,
		image.Rect(params.Margin, params.Margin, params.QrSize+params.Margin, params.QrSize+params.Margin),
		qrImage,
		image.Point{},
		draw.Over)

	regularFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	boldFont, err := truetype.Parse(gobold.TTF)
	if err != nil {
		return err
	}

	regularFace := truetype.NewFace(regularFont, &truetype.Options{
		Size: params.TitleFontSize,
		DPI:  params.Dpi,
	})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{
		Size: params.DescriptionFontSize,
		DPI:  params.Dpi,
	})

	createContext := func(font *truetype.Font, size float64) *freetype.Context {
		c := freetype.NewContext()
		c.SetDPI(params.Dpi)
		c.SetFont(font)
		c.SetFontSize(size)
		c.SetClip(img.Bounds())
		c.SetDst(img)
		c.SetSrc(image.NewUniform(color.Black))
		return c
	}

	boldContext := createContext(boldFont, params.TitleFontSize)
	regularContext := createContext(regularFont, params.DescriptionFontSize)

	maxWidth := params.Width - (params.Margin * 2) - params.QrSize - params.ComponentPadding
	lineSpacing := int(boldContext.PointToFixed(params.TitleFontSize).Round())
	textX := params.Margin + params.ComponentPadding + params.QrSize
	textY := params.Margin - 8

	titleLines := wrapText(params.TitleText, boldFace, maxWidth, boldContext)
	for _, line := range titleLines {
		pt := freetype.Pt(textX, textY+lineSpacing)
		_, err = boldContext.DrawString(line, pt)
		if err != nil {
			return err
		}
		textY += lineSpacing
	}

	textY += params.ComponentPadding / 4
	lineSpacing = int(regularContext.PointToFixed(params.DescriptionFontSize).Round())

	descriptionLines := wrapText(params.DescriptionText, regularFace, maxWidth, regularContext)
	for _, line := range descriptionLines {
		pt := freetype.Pt(textX, textY+lineSpacing)
		_, err = regularContext.DrawString(line, pt)
		if err != nil {
			return err
		}
		textY += lineSpacing
	}

	err = png.Encode(w, img)
	if err != nil {
		return err
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
