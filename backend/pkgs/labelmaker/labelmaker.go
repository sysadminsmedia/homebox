package labelmaker

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

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
	Padding             int
	TitleText           string
	TitleFontSize       float64
	DescriptionText     string
	DescriptionFontSize float64
	Dpi                 float64
	Url                 string
}

func NewGenerateParams(width int, height int, padding int, fontSize float64, title string, description string, url string) GenerateParameters {
	return GenerateParameters{
		Width:               width,
		Height:              height,
		QrSize:              height - (padding * 2),
		Padding:             padding,
		TitleText:           title,
		DescriptionText:     description,
		TitleFontSize:       fontSize,
		DescriptionFontSize: fontSize * 0.8,
		Dpi:                 72,
		Url:                 url,
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
	// Create QR code
	qr, err := qrcode.New(params.Url, qrcode.Medium)
	if err != nil {
		return err
	}
	qrImage := qr.Image(params.QrSize)

	// Create a new white background image
	bounds := image.Rect(0, 0, params.Width, params.Height)
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw QR code onto the image
	draw.Draw(img,
		image.Rect(params.Padding, params.Padding, params.QrSize+params.Padding, params.QrSize+params.Padding), // Position QR code at (50,50)
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

	maxWidth := params.Width - (params.Padding * 3) - params.QrSize
	lineSpacing := int(boldContext.PointToFixed(params.TitleFontSize).Round())
	textX := (params.Padding * 2) + params.QrSize
	textY := params.Padding + (lineSpacing / 2)

	titleLines := wrapText(params.TitleText, boldFace, maxWidth, boldContext)
	for _, line := range titleLines {
		pt := freetype.Pt(textX, textY+lineSpacing)
		_, err = boldContext.DrawString(line, pt)
		if err != nil {
			return err
		}
		textY += lineSpacing
	}

	lineSpacing = int(regularContext.PointToFixed(params.DescriptionFontSize).Round())
	textY += lineSpacing / 2

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
	f, err := os.CreateTemp("", "label-*.jpg")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
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
