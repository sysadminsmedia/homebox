package printer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"net"
	"time"
)

// BrotherRasterClient implements PrinterClient for Brother QL label printers
// using the Brother raster protocol over raw TCP (port 9100)
type BrotherRasterClient struct {
	address string
	port    string
	media   BrotherMediaType
}

// BrotherPrintOptions contains options for a Brother print job
type BrotherPrintOptions struct {
	MediaType     string  // Media type identifier (e.g., "DK-22251")
	LabelLengthMM float64 // Label length in mm (for continuous rolls)
}

// NewBrotherRasterClient creates a new Brother raster client
// Address should be the IP or hostname of the printer
func NewBrotherRasterClient(address string) (*BrotherRasterClient, error) {
	return &BrotherRasterClient{
		address: address,
		port:    "9100",
		media:   BrotherMediaTypes["DK-22251"], // Default to DK-22251
	}, nil
}

// SetMediaType configures the media type for the printer
func (c *BrotherRasterClient) SetMediaType(mediaType string) error {
	media, ok := GetBrotherMedia(mediaType)
	if !ok {
		return fmt.Errorf("unknown media type: %s", mediaType)
	}
	c.media = media
	return nil
}

// GetPrinterInfo returns information about the printer
func (c *BrotherRasterClient) GetPrinterInfo(ctx context.Context) (*PrinterInfo, error) {
	return &PrinterInfo{
		Name:  "Brother QL",
		Make:  "Brother",
		Model: "QL-820NWB",
		State: StatusOnline,
	}, nil
}

// Print sends a document to the printer
func (c *BrotherRasterClient) Print(ctx context.Context, job *PrintJob) (*PrintResult, error) {
	// Decode PNG image
	img, err := png.Decode(bytes.NewReader(job.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Convert to Brother raster format
	rasterData := c.convertToRaster(img)

	// Connect to printer
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.address, c.port), 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to printer: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send data
	_, err = conn.Write(rasterData)
	if err != nil {
		return nil, fmt.Errorf("failed to send data: %w", err)
	}

	return &PrintResult{
		Success: true,
		Message: "Print job sent successfully",
	}, nil
}

// convertToRaster converts an image to Brother raster format
func (c *BrotherRasterClient) convertToRaster(img image.Image) []byte {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Brother QL uses 90 bytes per line (720 dots) for 62mm, less for narrower
	bytesPerLine := 90
	printableWidth := c.media.PrintableDotsWidth
	if printableWidth == 0 {
		printableWidth = 696 // Default for 62mm
	}

	// Feed lines to add before and after content (about 3mm at 300 DPI)
	feedLines := 35
	// Total raster lines = feed + image + feed
	totalRasterLines := feedLines + imgHeight + feedLines

	var buf bytes.Buffer

	// 1. Invalidate - 200 null bytes
	buf.Write(make([]byte, 200))

	// 2. Initialize - ESC @
	buf.Write([]byte{0x1B, 0x40})

	// 3. Switch to raster mode - ESC i a 1
	buf.Write([]byte{0x1B, 0x69, 0x61, 0x01})

	// 4. Print info - ESC i z
	// Flags byte: 0x80 = quality, 0x08 = length valid, 0x04 = width valid, 0x02 = type valid
	flags := byte(0x80 | 0x08 | 0x04 | 0x02) // 0x8E
	buf.Write([]byte{0x1B, 0x69, 0x7A})
	buf.WriteByte(flags)
	buf.WriteByte(c.media.MediaCode) // Media type (0x0A=continuous, 0x0B=die-cut)
	buf.WriteByte(byte(c.media.WidthMM))
	buf.WriteByte(byte(c.media.LengthMM)) // Length in mm (0 for continuous)
	// Number of raster lines (4 bytes, little endian) - includes feed lines
	buf.WriteByte(byte(totalRasterLines & 0xFF))
	buf.WriteByte(byte((totalRasterLines >> 8) & 0xFF))
	buf.WriteByte(byte((totalRasterLines >> 16) & 0xFF))
	buf.WriteByte(byte((totalRasterLines >> 24) & 0xFF))
	buf.WriteByte(0) // First page (0 = first page only)
	buf.WriteByte(0) // Reserved

	// 5. Expanded mode - ESC i K
	// Bit 3 (0x08) = cut at end
	// Bit 0 (0x01) = two-color mode (for DK-22251, etc.)
	expandedFlags := byte(0x08)
	if c.media.TwoColor {
		expandedFlags |= 0x01
	}
	buf.Write([]byte{0x1B, 0x69, 0x4B, expandedFlags})

	// 6. Auto cut - ESC i M
	// 0x40 = auto cut enabled
	buf.Write([]byte{0x1B, 0x69, 0x4D, 0x40})

	// 7. Margin - ESC i d
	// Set margin to 35 dots (standard for continuous rolls)
	marginDots := uint16(35)
	buf.Write([]byte{0x1B, 0x69, 0x64, byte(marginDots & 0xFF), byte(marginDots >> 8)})

	// 8. Compression mode - M 0 (no compression)
	buf.Write([]byte{0x4D, 0x00})

	// 9. Raster data
	// Center the image in the printable area
	offsetDots := (printableWidth - imgWidth) / 2
	if offsetDots < 0 {
		offsetDots = 0
	}

	// Helper to write a raster line (blank or with data)
	// redData is optional - if nil, an empty red plane is sent in two-color mode
	writeRasterLine := func(blackData []byte, redData []byte) {
		if c.media.TwoColor {
			// Two-color mode: send BOTH planes for each line
			// First the black plane (0x01)
			buf.Write([]byte{0x77, 0x01, byte(bytesPerLine)})
			buf.Write(blackData)
			// Then the red plane (0x02) - empty if not provided
			buf.Write([]byte{0x77, 0x02, byte(bytesPerLine)})
			if redData != nil {
				buf.Write(redData)
			} else {
				buf.Write(make([]byte, bytesPerLine)) // Empty red plane
			}
		} else {
			// Standard mode: use 0x67 0x00
			buf.Write([]byte{0x67, 0x00, byte(bytesPerLine)})
			buf.Write(blackData)
		}
	}

	// Add leading blank lines for proper feed (about 3mm worth at 300 DPI = ~35 lines)
	blankLine := make([]byte, bytesPerLine)
	for i := 0; i < feedLines; i++ {
		writeRasterLine(blankLine, nil)
	}

	// Write the actual image data
	// Note: Brother QL printers require horizontally mirrored data
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		blackData := make([]byte, bytesPerLine)
		var redData []byte
		if c.media.TwoColor {
			redData = make([]byte, bytesPerLine)
		}

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rv, gv, bv, _ := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit
			r8 := float64(rv >> 8)
			g8 := float64(gv >> 8)
			b8 := float64(bv >> 8)

			// Calculate grayscale value
			gray := 0.299*r8 + 0.587*g8 + 0.114*b8

			// Mirror horizontally: map x to (imgWidth - 1 - x)
			mirroredX := imgWidth - 1 - (x - bounds.Min.X)
			dotPos := offsetDots + mirroredX
			byteIdx := dotPos / 8
			bitIdx := 7 - (dotPos % 8)

			if byteIdx < 0 || byteIdx >= bytesPerLine {
				continue
			}

			if gray < 200 { // Not white (threshold for printing)
				if c.media.TwoColor {
					// For two-color mode, detect if pixel is "red"
					// Red pixels have high R, low G, low B
					isRed := r8 > 150 && g8 < 100 && b8 < 100
					if isRed {
						redData[byteIdx] |= (1 << bitIdx)
					} else {
						blackData[byteIdx] |= (1 << bitIdx)
					}
				} else if gray < 128 {
					// Single color mode: all dark pixels are black
					blackData[byteIdx] |= (1 << bitIdx)
				}
			}
		}

		writeRasterLine(blackData, redData)
	}

	// Add trailing blank lines for proper feed before cut (about 3mm = ~35 lines)
	for i := 0; i < feedLines; i++ {
		writeRasterLine(blankLine, nil)
	}

	// 10. Print command
	buf.WriteByte(0x1A)

	return buf.Bytes()
}

// GetJobStatus checks the status of a print job
// Brother raster protocol doesn't support job status queries
func (c *BrotherRasterClient) GetJobStatus(ctx context.Context, jobID int) (*JobStatus, error) {
	return &JobStatus{
		JobID:     jobID,
		State:     "completed",
		Completed: true,
	}, nil
}

// CancelJob cancels a pending print job
// Brother raster protocol doesn't support job cancellation
func (c *BrotherRasterClient) CancelJob(ctx context.Context, jobID int) error {
	return nil
}
